#!/usr/bin/env bash
set -Eeuo pipefail

VERSION="${1:-8.0}"
IMAGE_PREFIX="${IMAGE_PREFIX:-swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io}"
MYSQL_IMAGE="${MYSQL_IMAGE:-$IMAGE_PREFIX/library/mysql:$VERSION}"
WORK_ROOT="${WORK_ROOT:-/tmp/dbtrue-mysql-pitr-verify}"
KEEP_ARTIFACTS="${KEEP_ARTIFACTS:-0}"
KEEP_CONTAINERS="${KEEP_CONTAINERS:-0}"
ROOT_PASSWORD="${ROOT_PASSWORD:-RootPass_12345!}"
DB_NAME="${DB_NAME:-dbtrue_pitr}"
BACKUP_MODE="${BACKUP_MODE:-stream}"
DOCKER_LIMITS_VALUE="${DOCKER_LIMITS:---cpus=1 --memory=1g}"
read -r -a DOCKER_LIMIT_ARGS <<< "$DOCKER_LIMITS_VALUE"
MYSQLD_EXTRA_ARGS_VALUE="${MYSQLD_EXTRA_ARGS:---innodb-buffer-pool-size=64M --performance-schema=OFF}"
read -r -a MYSQLD_EXTRA_ARGS <<< "$MYSQLD_EXTRA_ARGS_VALUE"
TOOL_USER="${TOOL_USER:-0:0}"
STEP_TIMEOUT="${STEP_TIMEOUT:-240s}"
TOOL_RUN_INDEX=0

case "$VERSION" in
  8.0|8.4) DEFAULT_XTRABACKUP_TAG="$VERSION" ;;
  *)
    echo "Unsupported MySQL version '$VERSION'. Expected 8.0 or 8.4." >&2
    exit 2
    ;;
esac
XTRABACKUP_TAG="${XTRABACKUP_TAG:-$DEFAULT_XTRABACKUP_TAG}"
XTRABACKUP_IMAGE="${XTRABACKUP_IMAGE:-$IMAGE_PREFIX/percona/percona-xtrabackup:$XTRABACKUP_TAG}"

case "$BACKUP_MODE" in
  stream|directory|coldcopy|hot-incremental) ;;
  *)
    echo "Unsupported BACKUP_MODE '$BACKUP_MODE'. Expected stream, directory, coldcopy, or hot-incremental." >&2
    exit 2
    ;;
esac
if [ "$BACKUP_MODE" = "coldcopy" ]; then
  TOOL_IMAGE="${TOOL_IMAGE:-$XTRABACKUP_IMAGE}"
else
  TOOL_IMAGE="${TOOL_IMAGE:-$XTRABACKUP_IMAGE}"
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required" >&2
  exit 2
fi
if ! command -v timeout >/dev/null 2>&1; then
  echo "GNU timeout is required" >&2
  exit 2
fi

PREFIX="dbtrue-mysql-physical-${VERSION//./}-$(date +%s)"
BASE="$WORK_ROOT/$PREFIX"
SRC="${PREFIX}-src"
DST="${PREFIX}-dst"
NET="${PREFIX}-net"
BINLOG_STREAM_CONTAINER="${PREFIX}-binlog-stream"
SRC_HOST="source-mysql"
DST_HOST="restore-mysql"

log() {
  printf '\n[%s] %s\n' "$(date -u +%H:%M:%S)" "$*"
}

cleanup() {
  set +e
  stop_binlog_stream >/dev/null 2>&1 || true
  docker rm -f "$BINLOG_STREAM_CONTAINER" >/dev/null 2>&1 || true
  docker ps -aq --filter "name=${PREFIX}-tool-" | xargs -r docker rm -f >/dev/null 2>&1 || true
  if [ "$KEEP_CONTAINERS" = "1" ]; then
    echo "Keeping verification containers with prefix $PREFIX"
  else
    docker rm -f "$SRC" "$DST" >/dev/null 2>&1 || true
    docker network rm "$NET" >/dev/null 2>&1 || true
  fi
  if [ "$KEEP_ARTIFACTS" != "1" ]; then
    rm -rf "$BASE"
  else
    echo "Keeping verification artifacts at $BASE"
  fi
}
trap cleanup EXIT

mysql_exec() {
  local container="$1"
  local sql="$2"
  docker exec "$container" mysql -uroot -p"$ROOT_PASSWORD" --protocol=socket -Nse "$sql"
}

container_ip() {
  local container="$1"
  docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$container"
}

wait_mysql() {
  local container="$1"
  for _ in $(seq 1 90); do
    if docker exec "$container" mysql -uroot -p"$ROOT_PASSWORD" --protocol=socket -Nse "SELECT 1" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  docker logs "$container" >&2 || true
  return 1
}

wait_mysql_tcp() {
  local host="$1"
  local container="${2:-}"
  log "Waiting for tool-container TCP access to $host"
  if run_tool --network "$NET" --entrypoint bash "$TOOL_IMAGE" -lc \
    "for _ in \$(seq 1 90); do if mysql --connect-timeout=2 --host=$host --port=3306 --user=root --password='$ROOT_PASSWORD' -Nse 'SELECT 1' >/dev/null 2>&1; then exit 0; fi; sleep 2; done; exit 1"; then
    return 0
  fi
  if [ -n "$container" ]; then
    docker logs "$container" >&2 || true
  fi
  return 1
}

run_tool() {
  TOOL_RUN_INDEX=$((TOOL_RUN_INDEX + 1))
  local tool_container="${PREFIX}-tool-${TOOL_RUN_INDEX}"
  timeout --kill-after=30s "$STEP_TIMEOUT" docker run --name "$tool_container" --rm --user "$TOOL_USER" "${DOCKER_LIMIT_ARGS[@]}" "$@"
  local status=$?
  docker rm -f "$tool_container" >/dev/null 2>&1 || true
  return "$status"
}

read_binlog_status() {
  local container="$1"
  if mysql_exec "$container" "SHOW BINARY LOG STATUS;" 2>/dev/null; then
    return 0
  fi
  mysql_exec "$container" "SHOW MASTER STATUS;"
}

start_binlog_stream() {
  local start_file="$1"
  log "Starting realtime binlog stream from $start_file"
  docker rm -f "$BINLOG_STREAM_CONTAINER" >/dev/null 2>&1 || true
  docker run -d --name "$BINLOG_STREAM_CONTAINER" --user "$TOOL_USER" --network "$NET" "${DOCKER_LIMIT_ARGS[@]}" \
    -v "$BASE/binlogs:/work/binlogs" \
    --entrypoint bash "$TOOL_IMAGE" -lc \
    "mysqlbinlog --read-from-remote-server --raw --stop-never --to-last-log --host=$SRC_CONNECT_HOST --port=3306 --user=root --password='$ROOT_PASSWORD' --result-file=/work/binlogs/ --connection-server-id=1703 $start_file" \
    >/dev/null
}

stop_binlog_stream() {
  set +e
  if docker ps -aq --filter "name=^/${BINLOG_STREAM_CONTAINER}$" | grep -q .; then
    docker logs "$BINLOG_STREAM_CONTAINER" > "$BASE/mysqlbinlog-stream.out" 2> "$BASE/mysqlbinlog-stream.log" || true
    docker stop -t 10 "$BINLOG_STREAM_CONTAINER" >/dev/null 2>&1 || true
    docker rm -f "$BINLOG_STREAM_CONTAINER" >/dev/null 2>&1 || true
  fi
  set -e
}

wait_for_binlog_file() {
  local file="$1"
  for _ in $(seq 1 90); do
    if [ -s "$BASE/binlogs/$file" ]; then
      return 0
    fi
    sleep 1
  done
  echo "timed out waiting for realtime binlog file $file" >&2
  ls -l "$BASE/binlogs" >&2 || true
  docker logs "$BINLOG_STREAM_CONTAINER" >&2 || true
  return 1
}

rm -rf "$BASE"
mkdir -p \
  "$BASE/source-data" \
  "$BASE/full" \
  "$BASE/inc1" \
  "$BASE/full-tmp" \
  "$BASE/meta" \
  "$BASE/full-meta" \
  "$BASE/inc1-meta" \
  "$BASE/binlogs" \
  "$BASE/restore-extract" \
  "$BASE/restore-data"

log "Using MySQL image: $MYSQL_IMAGE"
log "Using XtraBackup image: $XTRABACKUP_IMAGE"
log "Using backup tool image: $TOOL_IMAGE"
docker pull "$MYSQL_IMAGE" >/dev/null
if [ "$TOOL_IMAGE" != "$MYSQL_IMAGE" ]; then
  docker pull "$TOOL_IMAGE" >/dev/null
fi

log "Checking backup tools"
if [ "$BACKUP_MODE" = "coldcopy" ]; then
  run_tool --entrypoint bash "$TOOL_IMAGE" -lc \
    'mysqlbinlog --version && mysql --version'
else
  run_tool --entrypoint bash "$XTRABACKUP_IMAGE" -lc \
    'xtrabackup --version && xbstream --version && mysqlbinlog --version && mysql --version'
fi

log "Starting source MySQL $VERSION"
docker network create "$NET" >/dev/null
docker run -d --name "$SRC" --network "$NET" --network-alias "$SRC_HOST" "${DOCKER_LIMIT_ARGS[@]}" \
  -e MYSQL_ROOT_PASSWORD="$ROOT_PASSWORD" \
  -v "$BASE/source-data:/var/lib/mysql" \
  "$MYSQL_IMAGE" \
  --server-id=701 \
  --log-bin=mysql-bin \
  --binlog-format=ROW \
  --binlog-row-image=FULL \
  --sync-binlog=1 \
  "${MYSQLD_EXTRA_ARGS[@]}" >/dev/null
wait_mysql "$SRC"
SRC_CONNECT_HOST="$(container_ip "$SRC")"
wait_mysql_tcp "$SRC_CONNECT_HOST" "$SRC"
mysql_exec "$SRC" "SELECT VERSION();"

log "Creating seed data before full backup"
mysql_exec "$SRC" "
  CREATE DATABASE $DB_NAME;
  CREATE TABLE $DB_NAME.events(
    id INT PRIMARY KEY,
    phase VARCHAR(64) NOT NULL,
    note VARCHAR(64),
    created_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6)
  );
  INSERT INTO $DB_NAME.events(id, phase, note) VALUES (1, 'before_full', 'must_exist');
  FLUSH LOGS;
"

if [ "$BACKUP_MODE" = "hot-incremental" ]; then
  STREAM_START_FILE="$(read_binlog_status "$SRC" | awk '{print $1}')"
  if [ -z "$STREAM_START_FILE" ]; then
    echo "missing binlog stream start file" >&2
    exit 1
  fi
  start_binlog_stream "$STREAM_START_FILE"
  wait_for_binlog_file "$STREAM_START_FILE"
fi

log "Running $BACKUP_MODE physical full backup"
if [ "$BACKUP_MODE" = "coldcopy" ]; then
  read_binlog_status "$SRC" > "$BASE/meta/xtrabackup_binlog_info"
  log "Stopping source MySQL for cold-copy physical backup"
  docker stop "$SRC" >/dev/null
  cp -a "$BASE/source-data/." "$BASE/restore-extract/"
  log "Restarting source MySQL after cold-copy backup"
  docker start "$SRC" >/dev/null
  wait_mysql "$SRC"
  SRC_CONNECT_HOST="$(container_ip "$SRC")"
  wait_mysql_tcp "$SRC_CONNECT_HOST" "$SRC"
elif [ "$BACKUP_MODE" = "stream" ]; then
  if ! run_tool --network "$NET" \
    -v "$BASE:/work" \
    --volumes-from "$SRC" \
    --entrypoint bash "$XTRABACKUP_IMAGE" -lc \
    "xtrabackup --backup --host=$SRC_CONNECT_HOST --port=3306 --user=root --password='$ROOT_PASSWORD' --datadir=/var/lib/mysql --target-dir=/work/full-tmp --extra-lsndir=/work/meta --stream=xbstream" \
    > "$BASE/full.xbstream" 2> "$BASE/xtrabackup-backup.log"; then
    cat "$BASE/xtrabackup-backup.log" >&2 || true
    exit 1
  fi
  if [ ! -s "$BASE/full.xbstream" ]; then
    cat "$BASE/xtrabackup-backup.log" >&2 || true
    echo "full.xbstream is empty" >&2
    exit 1
  fi
elif [ "$BACKUP_MODE" = "hot-incremental" ]; then
  if ! run_tool --network "$NET" \
    -v "$BASE:/work" \
    --volumes-from "$SRC" \
    --entrypoint bash "$XTRABACKUP_IMAGE" -lc \
    "xtrabackup --backup --host=$SRC_CONNECT_HOST --port=3306 --user=root --password='$ROOT_PASSWORD' --datadir=/var/lib/mysql --target-dir=/work/full --extra-lsndir=/work/full-meta" \
    > "$BASE/xtrabackup-full.out" 2> "$BASE/xtrabackup-full.log"; then
    cat "$BASE/xtrabackup-full.log" >&2 || true
    exit 1
  fi
else
  if ! run_tool --network "$NET" \
    -v "$BASE:/work" \
    --volumes-from "$SRC" \
    --entrypoint bash "$XTRABACKUP_IMAGE" -lc \
    "xtrabackup --backup --host=$SRC_CONNECT_HOST --port=3306 --user=root --password='$ROOT_PASSWORD' --datadir=/var/lib/mysql --target-dir=/work/restore-extract --extra-lsndir=/work/meta" \
    > "$BASE/xtrabackup-backup.out" 2> "$BASE/xtrabackup-backup.log"; then
    cat "$BASE/xtrabackup-backup.log" >&2 || true
    exit 1
  fi
fi

BACKUP_BINLOG_INFO="$BASE/meta/xtrabackup_binlog_info"
if [ "$BACKUP_MODE" = "hot-incremental" ]; then
  log "Reading full backup binlog coordinate"
  cat "$BASE/full/xtrabackup_binlog_info"

  log "Writing row before incremental backup"
  sleep 2
  mysql_exec "$SRC" "INSERT INTO $DB_NAME.events(id, phase, note) VALUES (2, 'after_full_before_incremental', 'must_exist'); FLUSH LOGS;"

  log "Running hot incremental physical backup"
  if ! run_tool --network "$NET" \
    -v "$BASE:/work" \
    --volumes-from "$SRC" \
    --entrypoint bash "$XTRABACKUP_IMAGE" -lc \
    "xtrabackup --backup --host=$SRC_CONNECT_HOST --port=3306 --user=root --password='$ROOT_PASSWORD' --datadir=/var/lib/mysql --target-dir=/work/inc1 --incremental-basedir=/work/full --extra-lsndir=/work/inc1-meta" \
    > "$BASE/xtrabackup-inc1.out" 2> "$BASE/xtrabackup-inc1.log"; then
    cat "$BASE/xtrabackup-inc1.log" >&2 || true
    exit 1
  fi
  BACKUP_BINLOG_INFO="$BASE/inc1/xtrabackup_binlog_info"
fi

log "Reading backup binlog coordinate"
cat "$BACKUP_BINLOG_INFO"
BINLOG_FILE="$(awk '{print $1}' "$BACKUP_BINLOG_INFO")"
BINLOG_POS="$(awk '{print $2}' "$BACKUP_BINLOG_INFO")"
if [ -z "$BINLOG_FILE" ] || [ -z "$BINLOG_POS" ]; then
  echo "missing binlog coordinate" >&2
  exit 1
fi

log "Writing rows around PITR target"
sleep 2
if [ "$BACKUP_MODE" = "hot-incremental" ]; then
  mysql_exec "$SRC" "INSERT INTO $DB_NAME.events(id, phase, note) VALUES (3, 'before_target', 'must_exist');"
else
  mysql_exec "$SRC" "INSERT INTO $DB_NAME.events(id, phase, note) VALUES (2, 'before_target', 'must_exist');"
fi
sleep 2
TARGET_TIME="$(date -u '+%Y-%m-%d %H:%M:%S')"
log "PITR target time: $TARGET_TIME"
sleep 2
if [ "$BACKUP_MODE" = "hot-incremental" ]; then
  mysql_exec "$SRC" "INSERT INTO $DB_NAME.events(id, phase, note) VALUES (4, 'after_target', 'must_not_exist'); FLUSH LOGS;"
else
  mysql_exec "$SRC" "INSERT INTO $DB_NAME.events(id, phase, note) VALUES (3, 'after_target', 'must_not_exist'); FLUSH LOGS;"
fi

SRC_CONNECT_HOST="$(container_ip "$SRC")"
if [ "$BACKUP_MODE" = "hot-incremental" ]; then
  log "Waiting for realtime binlogs from $BINLOG_FILE"
  wait_for_binlog_file "$BINLOG_FILE"
  sleep 3
  stop_binlog_stream
else
  log "Capturing remote binlogs from $BINLOG_FILE"
  if ! run_tool --network "$NET" \
    -v "$BASE/binlogs:/work/binlogs" \
    --entrypoint bash "$TOOL_IMAGE" -lc \
    "mysqlbinlog --read-from-remote-server --raw --to-last-log --host=$SRC_CONNECT_HOST --port=3306 --user=root --password='$ROOT_PASSWORD' --result-file=/work/binlogs/ $BINLOG_FILE" \
    > "$BASE/mysqlbinlog-raw.out" 2> "$BASE/mysqlbinlog-raw.log"; then
    cat "$BASE/mysqlbinlog-raw.log" >&2 || true
    exit 1
  fi
fi
ls -l "$BASE/binlogs"

log "Stopping source MySQL before restore validation"
docker stop "$SRC" >/dev/null

if [ "$BACKUP_MODE" = "stream" ]; then
  log "Extracting xbstream"
  run_tool -v "$BASE:/work" --entrypoint bash "$XTRABACKUP_IMAGE" -lc \
    "cd /work/restore-extract && xbstream -x < /work/full.xbstream" \
    > "$BASE/xbstream-extract.out" 2> "$BASE/xbstream-extract.log"
fi

if [ "$BACKUP_MODE" = "hot-incremental" ]; then
  log "Preparing full backup with apply-log-only"
  if ! run_tool -v "$BASE:/work" --entrypoint bash "$XTRABACKUP_IMAGE" -lc \
    "xtrabackup --prepare --apply-log-only --target-dir=/work/full" \
    > "$BASE/xtrabackup-prepare-full.out" 2> "$BASE/xtrabackup-prepare-full.log"; then
    cat "$BASE/xtrabackup-prepare-full.log" >&2 || true
    exit 1
  fi

  log "Applying incremental backup"
  if ! run_tool -v "$BASE:/work" --entrypoint bash "$XTRABACKUP_IMAGE" -lc \
    "xtrabackup --prepare --target-dir=/work/full --incremental-dir=/work/inc1" \
    > "$BASE/xtrabackup-prepare-inc1.out" 2> "$BASE/xtrabackup-prepare-inc1.log"; then
    cat "$BASE/xtrabackup-prepare-inc1.log" >&2 || true
    exit 1
  fi
elif [ "$BACKUP_MODE" != "coldcopy" ]; then
  log "Preparing backup"
  if ! run_tool -v "$BASE:/work" --entrypoint bash "$XTRABACKUP_IMAGE" -lc \
    "xtrabackup --prepare --target-dir=/work/restore-extract" \
    > "$BASE/xtrabackup-prepare.out" 2> "$BASE/xtrabackup-prepare.log"; then
    cat "$BASE/xtrabackup-prepare.log" >&2 || true
    exit 1
  fi
else
  log "Cold-copy backup contains a consistent stopped datadir; skipping xtrabackup prepare"
fi

log "Copying prepared datadir"
if [ "$BACKUP_MODE" = "hot-incremental" ]; then
  cp -a "$BASE/full/." "$BASE/restore-data/"
else
  cp -a "$BASE/restore-extract/." "$BASE/restore-data/"
fi
chown -R 999:999 "$BASE/restore-data"

log "Starting restored MySQL $VERSION"
docker run -d --name "$DST" --network "$NET" --network-alias "$DST_HOST" "${DOCKER_LIMIT_ARGS[@]}" \
  -e MYSQL_ROOT_PASSWORD="$ROOT_PASSWORD" \
  -v "$BASE/restore-data:/var/lib/mysql" \
  "$MYSQL_IMAGE" \
  --server-id=702 \
  --skip-log-bin \
  "${MYSQLD_EXTRA_ARGS[@]}" >/dev/null
wait_mysql "$DST"
DST_CONNECT_HOST="$(container_ip "$DST")"
wait_mysql_tcp "$DST_CONNECT_HOST" "$DST"

log "Rows after full restore before PITR"
mysql_exec "$DST" "SELECT id, phase FROM $DB_NAME.events ORDER BY id;"
FULL_ROWS="$(mysql_exec "$DST" "SELECT GROUP_CONCAT(CONCAT(id, ':', phase) ORDER BY id SEPARATOR ',') FROM $DB_NAME.events;")"
EXPECTED_FULL_ROWS="1:before_full"
if [ "$BACKUP_MODE" = "hot-incremental" ]; then
  EXPECTED_FULL_ROWS="1:before_full,2:after_full_before_incremental"
fi
if [ "$FULL_ROWS" != "$EXPECTED_FULL_ROWS" ]; then
  echo "unexpected full restore rows: $FULL_ROWS" >&2
  echo "expected: $EXPECTED_FULL_ROWS" >&2
  exit 1
fi

log "Applying PITR binlogs through $TARGET_TIME from $BINLOG_FILE:$BINLOG_POS"
mapfile -t BINLOG_FILES < <(find "$BASE/binlogs" -maxdepth 1 -type f -name 'mysql-bin.*' | sort)
BINLOG_ARGS=()
emit=0
for path in "${BINLOG_FILES[@]}"; do
  base_name="$(basename "$path")"
  if [ "$base_name" = "$BINLOG_FILE" ]; then
    emit=1
  fi
  if [ "$emit" = "1" ]; then
    BINLOG_ARGS+=("/work/binlogs/$base_name")
  fi
done
if [ "${#BINLOG_ARGS[@]}" -eq 0 ]; then
  echo "no binlog files captured from $BINLOG_FILE" >&2
  exit 1
fi

printf '%q ' "${BINLOG_ARGS[@]}" > "$BASE/binlog-args.txt"
DST_CONNECT_HOST="$(container_ip "$DST")"
if ! run_tool --network "$NET" \
  -v "$BASE/binlogs:/work/binlogs:ro" \
  -v "$BASE/binlog-args.txt:/work/binlog-args.txt:ro" \
  --entrypoint bash "$TOOL_IMAGE" -lc \
  "mysqlbinlog --start-position=$BINLOG_POS --stop-datetime='$TARGET_TIME' \$(cat /work/binlog-args.txt) | mysql --host=$DST_CONNECT_HOST --port=3306 --user=root --password='$ROOT_PASSWORD'" \
  > "$BASE/mysqlbinlog-apply.out" 2> "$BASE/mysqlbinlog-apply.log"; then
  cat "$BASE/mysqlbinlog-apply.log" >&2 || true
  exit 1
fi

log "Rows after PITR"
mysql_exec "$DST" "SELECT id, phase FROM $DB_NAME.events ORDER BY id;"
PITR_ROWS="$(mysql_exec "$DST" "SELECT GROUP_CONCAT(CONCAT(id, ':', phase) ORDER BY id SEPARATOR ',') FROM $DB_NAME.events;")"
EXPECTED_PITR_ROWS="1:before_full,2:before_target"
if [ "$BACKUP_MODE" = "hot-incremental" ]; then
  EXPECTED_PITR_ROWS="1:before_full,2:after_full_before_incremental,3:before_target"
fi
if [ "$PITR_ROWS" != "$EXPECTED_PITR_ROWS" ]; then
  echo "unexpected PITR rows: $PITR_ROWS" >&2
  echo "expected: $EXPECTED_PITR_ROWS" >&2
  cat "$BASE/mysqlbinlog-apply.log" >&2 || true
  exit 1
fi

log "PASS mysql $VERSION $BACKUP_MODE physical backup + PITR verified"

package mysql_physical_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMySQL80PhysicalBackupPITR(t *testing.T) {
	runMySQLPhysicalPITRVerification(t, "8.0")
}

func TestMySQL84PhysicalBackupPITR(t *testing.T) {
	runMySQLPhysicalPITRVerification(t, "8.4")
}

func TestVerificationScriptWaitsForAuthenticatedSQL(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	if strings.Contains(scriptText, "mysqladmin") && strings.Contains(scriptText, " ping") {
		t.Fatalf("wait_mysql must not rely on mysqladmin ping; it can succeed before authenticated SQL is ready")
	}
	if !strings.Contains(scriptText, "SELECT 1") {
		t.Fatalf("wait_mysql should verify authenticated SQL with SELECT 1")
	}
}

func TestVerificationScriptWaitsForToolContainerTCPAccess(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	requiredSnippets := []string{
		`wait_mysql_tcp()`,
		`mysql --connect-timeout=2 --host=$host --port=3306`,
		`wait_mysql_tcp "$SRC_CONNECT_HOST"`,
		`wait_mysql_tcp "$DST_CONNECT_HOST"`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(scriptText, snippet) {
			t.Fatalf("verification script should contain %q so external backup tools wait for TCP readiness", snippet)
		}
	}
}

func TestVerificationScriptRunsBackupToolsAsRootForDatadirAccess(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	if !strings.Contains(scriptText, `TOOL_USER="${TOOL_USER:-0:0}"`) {
		t.Fatalf("verification script should default tool containers to root so xtrabackup can read bind-mounted datadirs")
	}
	if !strings.Contains(scriptText, `--user "$TOOL_USER"`) {
		t.Fatalf("verification script should pass TOOL_USER to docker run for backup tools")
	}
}

func TestVerificationScriptCleansUpTimedOutToolContainers(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	requiredSnippets := []string{
		"TOOL_RUN_INDEX=0",
		`--name "$tool_container"`,
		`timeout --kill-after=30s "$STEP_TIMEOUT" docker run`,
		`docker rm -f "$tool_container"`,
		`run_tool --network "$NET"`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(scriptText, snippet) {
			t.Fatalf("verification script should contain %q to keep tool containers bounded and clean", snippet)
		}
	}
}

func TestVerificationScriptUsesLowMemoryMySQLDefaults(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	requiredSnippets := []string{
		`MYSQLD_EXTRA_ARGS_VALUE="${MYSQLD_EXTRA_ARGS:---innodb-buffer-pool-size=64M --performance-schema=OFF}"`,
		`read -r -a MYSQLD_EXTRA_ARGS <<< "$MYSQLD_EXTRA_ARGS_VALUE"`,
		`"${MYSQLD_EXTRA_ARGS[@]}"`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(scriptText, snippet) {
			t.Fatalf("verification script should contain %q to keep MySQL containers lightweight during E2E verification", snippet)
		}
	}
}

func TestVerificationScriptStopsSourceBeforeRestoreValidation(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	stopSnippet := `log "Stopping source MySQL before restore validation"`
	startRestoreSnippet := `log "Starting restored MySQL $VERSION"`
	stopIndex := strings.Index(scriptText, stopSnippet)
	if stopIndex < 0 {
		t.Fatalf("verification script should contain %q to avoid running source and restored MySQL at the same time", stopSnippet)
	}
	startRestoreIndex := strings.Index(scriptText, startRestoreSnippet)
	if startRestoreIndex < 0 {
		t.Fatalf("verification script should contain %q", startRestoreSnippet)
	}
	if stopIndex > startRestoreIndex {
		t.Fatalf("verification script should stop the source MySQL before starting restored MySQL")
	}
}

func TestVerificationScriptUsesVolumesFromForXtraBackupDatadir(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	if !strings.Contains(scriptText, `--volumes-from "$SRC"`) {
		t.Fatalf("verification script should let xtrabackup access the MySQL datadir through --volumes-from")
	}
	if strings.Contains(scriptText, `-v "$BASE/source-data:/var/lib/mysql:ro"`) {
		t.Fatalf("verification script should not bind the source datadir separately for xtrabackup")
	}
}

func TestVerificationScriptSupportsDirectoryBackupMode(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	requiredSnippets := []string{
		`BACKUP_MODE="${BACKUP_MODE:-stream}"`,
		`stream|directory|coldcopy|hot-incremental) ;;`,
		`if [ "$BACKUP_MODE" = "stream" ]; then`,
		`--stream=xbstream`,
		`else`,
		`--target-dir=/work/restore-extract`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(scriptText, snippet) {
			t.Fatalf("verification script should contain %q to support directory backup mode", snippet)
		}
	}
}

func TestVerificationScriptAllowsXtraBackupImageOverrides(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	requiredSnippets := []string{
		`DEFAULT_XTRABACKUP_TAG="$VERSION"`,
		`XTRABACKUP_TAG="${XTRABACKUP_TAG:-$DEFAULT_XTRABACKUP_TAG}"`,
		`XTRABACKUP_IMAGE="${XTRABACKUP_IMAGE:-$IMAGE_PREFIX/percona/percona-xtrabackup:$XTRABACKUP_TAG}"`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(scriptText, snippet) {
			t.Fatalf("verification script should contain %q to allow site-specific xtrabackup image selection", snippet)
		}
	}
}

func TestVerificationScriptSupportsColdCopyPhysicalBackupMode(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	requiredSnippets := []string{
		`stream|directory|coldcopy|hot-incremental) ;;`,
		"if [ \"$BACKUP_MODE\" = \"coldcopy\" ]; then\n  TOOL_IMAGE=\"${TOOL_IMAGE:-$XTRABACKUP_IMAGE}\"",
		`if [ "$BACKUP_MODE" = "coldcopy" ]; then`,
		`docker stop "$SRC"`,
		`cp -a "$BASE/source-data/." "$BASE/restore-extract/"`,
		`docker start "$SRC"`,
		`if [ "$BACKUP_MODE" != "coldcopy" ]; then`,
		`run_tool -v "$BASE:/work" --entrypoint bash "$XTRABACKUP_IMAGE" -lc`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(scriptText, snippet) {
			t.Fatalf("verification script should contain %q to support cold-copy physical backup/PITR verification", snippet)
		}
	}
}

func TestVerificationScriptSupportsHotIncrementalBackupMode(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	requiredSnippets := []string{
		`stream|directory|coldcopy|hot-incremental) ;;`,
		`if [ "$BACKUP_MODE" = "hot-incremental" ]; then`,
		`--target-dir=/work/full`,
		`--target-dir=/work/inc1`,
		`--incremental-basedir=/work/full`,
		`cat "$BASE/full/xtrabackup_binlog_info"`,
		`BACKUP_BINLOG_INFO="$BASE/inc1/xtrabackup_binlog_info"`,
		`xtrabackup --prepare --apply-log-only --target-dir=/work/full`,
		`xtrabackup --prepare --target-dir=/work/full --incremental-dir=/work/inc1`,
		`cp -a "$BASE/full/." "$BASE/restore-data/"`,
		`2:after_full_before_incremental`,
		`3:before_target`,
		`VALUES (4, 'after_target', 'must_not_exist')`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(scriptText, snippet) {
			t.Fatalf("verification script should contain %q to support hot full plus incremental backup verification", snippet)
		}
	}
}

func TestVerificationScriptStreamsBinlogsInRealtimeForHotIncrementalMode(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	requiredSnippets := []string{
		`BINLOG_STREAM_CONTAINER="${PREFIX}-binlog-stream"`,
		`docker run -d --name "$BINLOG_STREAM_CONTAINER" --user "$TOOL_USER"`,
		`mysqlbinlog --read-from-remote-server --raw --stop-never --to-last-log`,
		`--connection-server-id=1703`,
		`docker rm -f "$BINLOG_STREAM_CONTAINER"`,
		`wait_for_binlog_file "$BINLOG_FILE"`,
		`stop_binlog_stream`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(scriptText, snippet) {
			t.Fatalf("verification script should contain %q to stream remote binlogs continuously", snippet)
		}
	}
}

func TestVerificationScriptCanKeepContainersForFailureDiagnostics(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	requiredSnippets := []string{
		`KEEP_CONTAINERS="${KEEP_CONTAINERS:-0}"`,
		`if [ "$KEEP_CONTAINERS" = "1" ]; then`,
		`echo "Keeping verification containers with prefix $PREFIX"`,
		`else`,
		`docker rm -f "$SRC" "$DST"`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(scriptText, snippet) {
			t.Fatalf("verification script should contain %q to preserve containers for failure diagnostics", snippet)
		}
	}
}

func TestVerificationScriptUsesInspectedContainerIPsForToolConnections(t *testing.T) {
	script, err := os.ReadFile("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to read verification script: %v", err)
	}

	scriptText := string(script)
	requiredSnippets := []string{
		`container_ip()`,
		`SRC_CONNECT_HOST="$(container_ip "$SRC")"`,
		`DST_CONNECT_HOST="$(container_ip "$DST")"`,
		`--network-alias "$SRC_HOST"`,
		`--network-alias "$DST_HOST"`,
		`--host=$SRC_CONNECT_HOST`,
		`--host=$DST_CONNECT_HOST`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(scriptText, snippet) {
			t.Fatalf("verification script should contain %q so tool containers do not rely on Docker DNS", snippet)
		}
	}
}

func runMySQLPhysicalPITRVerification(t *testing.T, version string) {
	t.Helper()

	if os.Getenv("RUN_MYSQL_PHYSICAL_PITR_E2E") != "1" {
		t.Skip("set RUN_MYSQL_PHYSICAL_PITR_E2E=1 to run real Docker MySQL physical backup/PITR verification")
	}

	scriptPath, err := filepath.Abs("verify_mysql_physical_pitr.sh")
	if err != nil {
		t.Fatalf("failed to resolve verification script path: %v", err)
	}

	cmd := exec.Command("bash", scriptPath, version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		t.Fatalf("mysql %s physical backup/PITR verification failed: %v", version, err)
	}
}

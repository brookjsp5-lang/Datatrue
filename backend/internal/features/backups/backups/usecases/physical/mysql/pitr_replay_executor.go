package usecases_physical_mysql

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type PITRReplayRequest struct {
	MysqlbinlogBin string
	MysqlBin       string
	WorkDir        string
	Host           string
	Port           int
	Username       string
	Password       string
	StartPosition  string
	TargetTime     time.Time
	BinlogFiles    []string
}

func ReplayBinlogs(ctx context.Context, request PITRReplayRequest, runner PipelineRunner) error {
	if runner == nil {
		return errors.New("pipeline runner is required")
	}
	if request.MysqlbinlogBin == "" {
		return errors.New("mysqlbinlog binary is required")
	}
	if request.MysqlBin == "" {
		return errors.New("mysql binary is required")
	}
	if request.WorkDir == "" {
		return errors.New("work dir is required")
	}
	if request.StartPosition == "" {
		return errors.New("start position is required")
	}
	if request.TargetTime.IsZero() {
		return errors.New("target time is required")
	}
	if len(request.BinlogFiles) == 0 {
		return errors.New("binlog files are required")
	}

	if err := os.MkdirAll(request.WorkDir, 0o700); err != nil {
		return err
	}

	defaultsFilePath := filepath.Join(request.WorkDir, "mysql-replay-defaults.cnf")
	if err := writeDefaultsFile(defaultsFilePath, request.Username, request.Password); err != nil {
		return err
	}

	mysqlbinlogArgs := []string{
		"--defaults-file=" + defaultsFilePath,
		"--start-position=" + request.StartPosition,
		"--stop-datetime=" + request.TargetTime.UTC().Format("2006-01-02 15:04:05"),
	}
	mysqlbinlogArgs = append(mysqlbinlogArgs, request.BinlogFiles...)

	mysqlArgs := []string{
		"--defaults-file=" + defaultsFilePath,
		"--host=" + request.Host,
		"--port=" + strconv.Itoa(request.Port),
		"--user=" + request.Username,
	}

	var stderr bytes.Buffer
	if err := runner.RunPipeline(ctx, CommandExecution{
		Executable: request.MysqlbinlogBin,
		Args:       mysqlbinlogArgs,
		Stderr:     &stderr,
	}, CommandExecution{
		Executable: request.MysqlBin,
		Args:       mysqlArgs,
		Stderr:     &stderr,
	}); err != nil {
		return fmt.Errorf("mysql binlog replay failed: %w; stderr: %s", err, stderr.String())
	}

	return nil
}

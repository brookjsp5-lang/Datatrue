package usecases_physical_mysql

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"databasus-backend/internal/util/tools"
)

type CommandExecution struct {
	Executable string
	Args       []string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
}

type CommandRunner interface {
	Run(ctx context.Context, command CommandExecution) error
}

type PipelineRunner interface {
	RunPipeline(ctx context.Context, producer CommandExecution, consumer CommandExecution) error
}

type FullBackupExecutionRequest struct {
	XtraBackupBin   string
	WorkDir         string
	Host            string
	Port            int
	Username        string
	Password        string
	TargetDir       string
	DataDir         string
	Parallel        int
	CompressThreads int
	StreamWriter    io.Writer
}

type FullBackupExecutionResult struct {
	BinlogFile     string
	BinlogPosition string
}

func ExecuteFullBackup(
	ctx context.Context,
	request FullBackupExecutionRequest,
	runner CommandRunner,
) (FullBackupExecutionResult, error) {
	if runner == nil {
		return FullBackupExecutionResult{}, errors.New("command runner is required")
	}
	if request.WorkDir == "" {
		return FullBackupExecutionResult{}, errors.New("work dir is required")
	}
	if request.StreamWriter == nil {
		return FullBackupExecutionResult{}, errors.New("stream writer is required")
	}

	if err := os.MkdirAll(request.WorkDir, 0o700); err != nil {
		return FullBackupExecutionResult{}, err
	}

	defaultsFilePath := filepath.Join(request.WorkDir, "mysql-defaults.cnf")
	if err := writeDefaultsFile(defaultsFilePath, request.Username, request.Password); err != nil {
		return FullBackupExecutionResult{}, err
	}

	extraLsnDir := filepath.Join(request.WorkDir, "xtrabackup-lsn")
	if err := os.MkdirAll(extraLsnDir, 0o700); err != nil {
		return FullBackupExecutionResult{}, err
	}

	plan, err := BuildFullBackupCommand(FullBackupCommandRequest{
		XtraBackupBin:    request.XtraBackupBin,
		DefaultsFilePath: defaultsFilePath,
		TargetDir:        request.TargetDir,
		ExtraLsnDir:      extraLsnDir,
		DataDir:          request.DataDir,
		Parallel:         request.Parallel,
		CompressThreads:  request.CompressThreads,
	})
	if err != nil {
		return FullBackupExecutionResult{}, err
	}

	plan.Args = append(plan.Args,
		"--host="+request.Host,
		fmt.Sprintf("--port=%d", request.Port),
		"--user="+request.Username,
	)

	var stderr bytes.Buffer
	if err := runner.Run(ctx, CommandExecution{
		Executable: plan.Executable,
		Args:       plan.Args,
		Stdout:     request.StreamWriter,
		Stderr:     &stderr,
	}); err != nil {
		return FullBackupExecutionResult{}, fmt.Errorf("xtrabackup failed: %w; stderr: %s", err, strings.TrimSpace(stderr.String()))
	}

	return readBinlogInfo(filepath.Join(extraLsnDir, "xtrabackup_binlog_info"))
}

func writeDefaultsFile(path, username, password string) error {
	if username == "" {
		return errors.New("username is required")
	}

	content := fmt.Sprintf(
		"[client]\nuser=%s\npassword=\"%s\"\n",
		username,
		tools.EscapeMysqlPassword(password),
	)

	return os.WriteFile(path, []byte(content), 0o600)
}

func readBinlogInfo(path string) (FullBackupExecutionResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return FullBackupExecutionResult{}, err
	}

	fields := strings.Fields(string(content))
	if len(fields) < 2 {
		return FullBackupExecutionResult{}, fmt.Errorf("invalid xtrabackup_binlog_info: %q", string(content))
	}

	return FullBackupExecutionResult{
		BinlogFile:     fields[0],
		BinlogPosition: fields[1],
	}, nil
}

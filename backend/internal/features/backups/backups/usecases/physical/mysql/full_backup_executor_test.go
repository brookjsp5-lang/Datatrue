package usecases_physical_mysql

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestExecuteFullBackupStreamsBackupAndReturnsBinlogCoordinate(t *testing.T) {
	var stream bytes.Buffer
	runner := &recordingCommandRunner{
		t: t,
		onRun: func(command CommandExecution) error {
			if strings.Contains(strings.Join(command.Args, " "), "secret") {
				t.Fatalf("password must not be exposed in command arguments: %v", command.Args)
			}

			defaultsFile := argValue(command.Args, "--defaults-file=")
			if defaultsFile == "" {
				t.Fatalf("expected defaults file argument, got %v", command.Args)
			}
			defaultsContent, err := os.ReadFile(defaultsFile)
			if err != nil {
				t.Fatalf("failed to read defaults file: %v", err)
			}
			defaultsText := string(defaultsContent)
			if !strings.Contains(defaultsText, "user=backup_user") {
				t.Fatalf("defaults file does not contain user: %s", defaultsText)
			}
			if !strings.Contains(defaultsText, `password="top\\\"secret"`) {
				t.Fatalf("defaults file does not contain escaped password: %s", defaultsText)
			}

			extraLsnDir := argValue(command.Args, "--extra-lsndir=")
			if extraLsnDir == "" {
				t.Fatalf("expected extra LSN dir argument, got %v", command.Args)
			}
			if err := os.MkdirAll(extraLsnDir, 0o700); err != nil {
				t.Fatalf("failed to create extra LSN dir: %v", err)
			}
			if err := os.WriteFile(
				filepath.Join(extraLsnDir, "xtrabackup_binlog_info"),
				[]byte("mysql-bin.000003\t456\n"),
				0o600,
			); err != nil {
				t.Fatalf("failed to write binlog info: %v", err)
			}

			_, err = command.Stdout.Write([]byte("xbstream-bytes"))
			return err
		},
	}

	result, err := ExecuteFullBackup(context.Background(), FullBackupExecutionRequest{
		XtraBackupBin:   "/usr/bin/xtrabackup",
		WorkDir:         t.TempDir(),
		Host:            "mysql.example.com",
		Port:            3306,
		Username:        "backup_user",
		Password:        `top\"secret`,
		TargetDir:       "/tmp/full",
		DataDir:         "/var/lib/mysql",
		Parallel:        2,
		CompressThreads: 1,
		StreamWriter:    &stream,
	}, runner)

	if err != nil {
		t.Fatalf("expected full backup execution to succeed, got error: %v", err)
	}
	if result.BinlogFile != "mysql-bin.000003" {
		t.Fatalf("expected binlog file mysql-bin.000003, got %s", result.BinlogFile)
	}
	if result.BinlogPosition != "456" {
		t.Fatalf("expected binlog position 456, got %s", result.BinlogPosition)
	}
	if stream.String() != "xbstream-bytes" {
		t.Fatalf("expected streamed xbstream bytes, got %q", stream.String())
	}
	if runner.command.Executable != "/usr/bin/xtrabackup" {
		t.Fatalf("unexpected executable: %s", runner.command.Executable)
	}
	if !slices.Contains(runner.command.Args, "--stream=xbstream") {
		t.Fatalf("expected streaming backup argument, got %v", runner.command.Args)
	}
}

type recordingCommandRunner struct {
	t       *testing.T
	command CommandExecution
	onRun   func(command CommandExecution) error
}

func (r *recordingCommandRunner) Run(_ context.Context, command CommandExecution) error {
	r.command = command
	if r.onRun == nil {
		return nil
	}
	return r.onRun(command)
}

func argValue(args []string, prefix string) string {
	for _, arg := range args {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}
	return ""
}

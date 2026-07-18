package usecases_physical_mysql

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestReplayBinlogsBuildsPipelineWithoutExposingPasswords(t *testing.T) {
	runner := &recordingPipelineRunner{
		t: t,
		onRunPipeline: func(producer CommandExecution, consumer CommandExecution) error {
			commandParts := append([]string{producer.Executable}, producer.Args...)
			commandParts = append(commandParts, consumer.Executable)
			commandParts = append(commandParts, consumer.Args...)
			commandText := strings.Join(commandParts, " ")
			if strings.Contains(commandText, "top-secret") {
				t.Fatalf("password must not be exposed in command text: %s", commandText)
			}

			defaultsFile := argValue(producer.Args, "--defaults-file=")
			if defaultsFile == "" {
				t.Fatalf("expected defaults file argument, got %v", producer.Args)
			}
			defaultsContent, err := os.ReadFile(defaultsFile)
			if err != nil {
				t.Fatalf("failed to read defaults file: %v", err)
			}
			if !strings.Contains(string(defaultsContent), `password="top-secret"`) {
				t.Fatalf("defaults file does not contain password: %s", string(defaultsContent))
			}

			return nil
		},
	}

	target := time.Date(2026, 7, 17, 10, 20, 30, 0, time.UTC)
	err := ReplayBinlogs(context.Background(), PITRReplayRequest{
		MysqlbinlogBin: "/usr/bin/mysqlbinlog",
		MysqlBin:       "/usr/bin/mysql",
		WorkDir:        t.TempDir(),
		Host:           "restore-mysql",
		Port:           3306,
		Username:       "root",
		Password:       "top-secret",
		StartPosition:  "456",
		TargetTime:     target,
		BinlogFiles: []string{
			"/backup/mysql-bin.000003",
			"/backup/mysql-bin.000004",
		},
	}, runner)

	if err != nil {
		t.Fatalf("expected PITR replay execution to succeed, got error: %v", err)
	}

	if runner.producer.Executable != "/usr/bin/mysqlbinlog" {
		t.Fatalf("expected mysqlbinlog executable, got %s", runner.producer.Executable)
	}
	if runner.consumer.Executable != "/usr/bin/mysql" {
		t.Fatalf("expected mysql executable, got %s", runner.consumer.Executable)
	}

	requiredProducerArgs := []string{
		"--start-position=456",
		"--stop-datetime=2026-07-17 10:20:30",
		"/backup/mysql-bin.000003",
		"/backup/mysql-bin.000004",
	}
	for _, requiredArg := range requiredProducerArgs {
		if !containsArg(runner.producer.Args, requiredArg) {
			t.Fatalf("expected mysqlbinlog args to contain %s, got %v", requiredArg, runner.producer.Args)
		}
	}
	if containsArg(runner.producer.Args, "|") {
		t.Fatalf("mysqlbinlog args must not contain a shell pipe: %v", runner.producer.Args)
	}

	requiredConsumerArgs := []string{
		"--host=restore-mysql",
		"--port=3306",
		"--user=root",
	}
	for _, requiredArg := range requiredConsumerArgs {
		if !containsArg(runner.consumer.Args, requiredArg) {
			t.Fatalf("expected mysql args to contain %s, got %v", requiredArg, runner.consumer.Args)
		}
	}
}

type recordingPipelineRunner struct {
	t             *testing.T
	producer      CommandExecution
	consumer      CommandExecution
	onRunPipeline func(producer CommandExecution, consumer CommandExecution) error
}

func (r *recordingPipelineRunner) RunPipeline(
	_ context.Context,
	producer CommandExecution,
	consumer CommandExecution,
) error {
	r.producer = producer
	r.consumer = consumer
	if r.onRunPipeline == nil {
		return nil
	}
	return r.onRunPipeline(producer, consumer)
}

func containsArg(args []string, expected string) bool {
	for _, arg := range args {
		if arg == expected {
			return true
		}
	}
	return false
}

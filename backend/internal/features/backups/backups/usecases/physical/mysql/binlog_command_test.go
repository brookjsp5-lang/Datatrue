package usecases_physical_mysql

import (
	"slices"
	"strings"
	"testing"
)

func Test_BuildBinlogStreamCommand_IncludesRemoteRawStreamingArguments(t *testing.T) {
	request := BinlogStreamCommandRequest{
		MysqlbinlogBin:   "/usr/bin/mysqlbinlog",
		DefaultsFilePath: "/run/databasus/mysql.cnf",
		Host:             "mysql.example.com",
		Port:             3306,
		Username:         "backup_user",
		ResultFilePrefix: "/var/lib/databasus/binlogs/mysql-bin.",
		StartFile:        "mysql-bin.000123",
		IsTLS:            true,
		ServerID:         184467,
	}

	plan, err := BuildBinlogStreamCommand(request)

	if err != nil {
		t.Fatalf("expected command plan, got error: %v", err)
	}

	requiredArgs := []string{
		"--defaults-file=/run/databasus/mysql.cnf",
		"--read-from-remote-server",
		"--raw",
		"--stop-never",
		"--to-last-log",
		"--host=mysql.example.com",
		"--port=3306",
		"--user=backup_user",
		"--result-file=/var/lib/databasus/binlogs/mysql-bin.",
		"--ssl-mode=REQUIRED",
		"--connection-server-id=184467",
		"mysql-bin.000123",
	}
	for _, requiredArg := range requiredArgs {
		if !slices.Contains(plan.Args, requiredArg) {
			t.Fatalf("expected args to contain %s, got %v", requiredArg, plan.Args)
		}
	}
}

func Test_BuildBinlogStreamCommand_DoesNotExposePasswords(t *testing.T) {
	request := BinlogStreamCommandRequest{
		MysqlbinlogBin:   "/usr/bin/mysqlbinlog",
		DefaultsFilePath: "/run/databasus/mysql.cnf",
		Host:             "mysql.example.com",
		Port:             3306,
		Username:         "backup_user",
		ResultFilePrefix: "/var/lib/databasus/binlogs/mysql-bin.",
		StartFile:        "mysql-bin.000123",
	}

	plan, err := BuildBinlogStreamCommand(request)

	if err != nil {
		t.Fatalf("expected command plan, got error: %v", err)
	}

	commandText := strings.Join(append([]string{plan.Executable}, plan.Args...), " ")
	if strings.Contains(commandText, "password") {
		t.Fatalf("expected command text to avoid password-bearing arguments, got %s", commandText)
	}
}

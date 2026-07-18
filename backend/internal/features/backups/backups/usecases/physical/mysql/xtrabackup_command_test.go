package usecases_physical_mysql

import (
	"slices"
	"strings"
	"testing"
)

func Test_BuildFullBackupCommand_IncludesStreamingBackupArguments(t *testing.T) {
	request := FullBackupCommandRequest{
		XtraBackupBin:    "/usr/bin/xtrabackup",
		DefaultsFilePath: "/run/databasus/mysql.cnf",
		TargetDir:        "/var/lib/databasus/mysql/full",
		ExtraLsnDir:      "/var/lib/databasus/mysql/lsn",
		DataDir:          "/var/lib/mysql",
		Parallel:         4,
		CompressThreads:  2,
	}

	plan, err := BuildFullBackupCommand(request)

	if err != nil {
		t.Fatalf("expected command plan, got error: %v", err)
	}
	if plan.Executable != "/usr/bin/xtrabackup" {
		t.Fatalf("expected executable path to be preserved, got %s", plan.Executable)
	}

	requiredArgs := []string{
		"--defaults-file=/run/databasus/mysql.cnf",
		"--backup",
		"--stream=xbstream",
		"--target-dir=/var/lib/databasus/mysql/full",
		"--extra-lsndir=/var/lib/databasus/mysql/lsn",
		"--datadir=/var/lib/mysql",
		"--parallel=4",
		"--compress",
		"--compress-threads=2",
	}
	for _, requiredArg := range requiredArgs {
		if !slices.Contains(plan.Args, requiredArg) {
			t.Fatalf("expected args to contain %s, got %v", requiredArg, plan.Args)
		}
	}
}

func Test_BuildFullBackupCommand_DoesNotExposePasswords(t *testing.T) {
	request := FullBackupCommandRequest{
		XtraBackupBin:    "/usr/bin/xtrabackup",
		DefaultsFilePath: "/run/databasus/mysql.cnf",
		TargetDir:        "/var/lib/databasus/mysql/full",
		ExtraLsnDir:      "/var/lib/databasus/mysql/lsn",
	}

	plan, err := BuildFullBackupCommand(request)

	if err != nil {
		t.Fatalf("expected command plan, got error: %v", err)
	}

	commandText := strings.Join(append([]string{plan.Executable}, plan.Args...), " ")
	if strings.Contains(commandText, "password") {
		t.Fatalf("expected command text to avoid password-bearing arguments, got %s", commandText)
	}
}

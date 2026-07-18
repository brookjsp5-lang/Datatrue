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

func Test_BuildIncrementalBackupCommand_IncludesIncrementalBasedir(t *testing.T) {
	request := IncrementalBackupCommandRequest{
		XtraBackupBin:      "/usr/bin/xtrabackup",
		DefaultsFilePath:   "/run/databasus/mysql.cnf",
		TargetDir:          "/var/lib/databasus/mysql/inc1",
		IncrementalBaseDir: "/var/lib/databasus/mysql/full",
		ExtraLsnDir:        "/var/lib/databasus/mysql/inc1-lsn",
		DataDir:            "/var/lib/mysql",
		Parallel:           3,
	}

	plan, err := BuildIncrementalBackupCommand(request)

	if err != nil {
		t.Fatalf("expected command plan, got error: %v", err)
	}

	requiredArgs := []string{
		"--defaults-file=/run/databasus/mysql.cnf",
		"--backup",
		"--target-dir=/var/lib/databasus/mysql/inc1",
		"--incremental-basedir=/var/lib/databasus/mysql/full",
		"--extra-lsndir=/var/lib/databasus/mysql/inc1-lsn",
		"--datadir=/var/lib/mysql",
		"--parallel=3",
	}
	for _, requiredArg := range requiredArgs {
		if !slices.Contains(plan.Args, requiredArg) {
			t.Fatalf("expected args to contain %s, got %v", requiredArg, plan.Args)
		}
	}
	if slices.Contains(plan.Args, "--stream=xbstream") {
		t.Fatalf("incremental directory backups should not force xbstream output, got %v", plan.Args)
	}
}

func Test_BuildPrepareCommand_CanApplyFullAndIncrementalBackups(t *testing.T) {
	fullPlan, err := BuildPrepareBackupCommand(PrepareBackupCommandRequest{
		XtraBackupBin: "/usr/bin/xtrabackup",
		TargetDir:     "/var/lib/databasus/mysql/full",
		ApplyLogOnly:  true,
	})
	if err != nil {
		t.Fatalf("expected full prepare command plan, got error: %v", err)
	}

	incrementalPlan, err := BuildPrepareBackupCommand(PrepareBackupCommandRequest{
		XtraBackupBin:  "/usr/bin/xtrabackup",
		TargetDir:      "/var/lib/databasus/mysql/full",
		IncrementalDir: "/var/lib/databasus/mysql/inc1",
	})
	if err != nil {
		t.Fatalf("expected incremental prepare command plan, got error: %v", err)
	}

	if !slices.Contains(fullPlan.Args, "--apply-log-only") {
		t.Fatalf("expected full prepare args to contain --apply-log-only, got %v", fullPlan.Args)
	}
	requiredIncrementalArgs := []string{
		"--prepare",
		"--target-dir=/var/lib/databasus/mysql/full",
		"--incremental-dir=/var/lib/databasus/mysql/inc1",
	}
	for _, requiredArg := range requiredIncrementalArgs {
		if !slices.Contains(incrementalPlan.Args, requiredArg) {
			t.Fatalf("expected incremental prepare args to contain %s, got %v", requiredArg, incrementalPlan.Args)
		}
	}
	if slices.Contains(incrementalPlan.Args, "--apply-log-only") {
		t.Fatalf("final incremental prepare should not include --apply-log-only, got %v", incrementalPlan.Args)
	}
}

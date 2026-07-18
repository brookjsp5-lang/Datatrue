package usecases_physical_mysql

import (
	"errors"
	"fmt"
)

type FullBackupCommandRequest struct {
	XtraBackupBin    string
	DefaultsFilePath string
	TargetDir        string
	ExtraLsnDir      string
	DataDir          string
	Parallel         int
	CompressThreads  int
}

type IncrementalBackupCommandRequest struct {
	XtraBackupBin      string
	DefaultsFilePath   string
	TargetDir          string
	IncrementalBaseDir string
	ExtraLsnDir        string
	DataDir            string
	Parallel           int
}

type PrepareBackupCommandRequest struct {
	XtraBackupBin  string
	TargetDir      string
	IncrementalDir string
	ApplyLogOnly   bool
}

func BuildFullBackupCommand(request FullBackupCommandRequest) (CommandPlan, error) {
	if request.XtraBackupBin == "" {
		return CommandPlan{}, errors.New("xtrabackup binary is required")
	}
	if request.DefaultsFilePath == "" {
		return CommandPlan{}, errors.New("defaults file path is required")
	}
	if request.TargetDir == "" {
		return CommandPlan{}, errors.New("target directory is required")
	}
	if request.ExtraLsnDir == "" {
		return CommandPlan{}, errors.New("extra LSN directory is required")
	}

	args := []string{
		"--defaults-file=" + request.DefaultsFilePath,
		"--backup",
		"--stream=xbstream",
		"--target-dir=" + request.TargetDir,
		"--extra-lsndir=" + request.ExtraLsnDir,
	}

	if request.DataDir != "" {
		args = append(args, "--datadir="+request.DataDir)
	}
	if request.Parallel > 0 {
		args = append(args, fmt.Sprintf("--parallel=%d", request.Parallel))
	}
	if request.CompressThreads > 0 {
		args = append(args,
			"--compress",
			fmt.Sprintf("--compress-threads=%d", request.CompressThreads),
		)
	}

	return CommandPlan{
		Executable: request.XtraBackupBin,
		Args:       args,
	}, nil
}

func BuildIncrementalBackupCommand(request IncrementalBackupCommandRequest) (CommandPlan, error) {
	if request.XtraBackupBin == "" {
		return CommandPlan{}, errors.New("xtrabackup binary is required")
	}
	if request.DefaultsFilePath == "" {
		return CommandPlan{}, errors.New("defaults file path is required")
	}
	if request.TargetDir == "" {
		return CommandPlan{}, errors.New("target directory is required")
	}
	if request.IncrementalBaseDir == "" {
		return CommandPlan{}, errors.New("incremental base directory is required")
	}
	if request.ExtraLsnDir == "" {
		return CommandPlan{}, errors.New("extra LSN directory is required")
	}

	args := []string{
		"--defaults-file=" + request.DefaultsFilePath,
		"--backup",
		"--target-dir=" + request.TargetDir,
		"--incremental-basedir=" + request.IncrementalBaseDir,
		"--extra-lsndir=" + request.ExtraLsnDir,
	}

	if request.DataDir != "" {
		args = append(args, "--datadir="+request.DataDir)
	}
	if request.Parallel > 0 {
		args = append(args, fmt.Sprintf("--parallel=%d", request.Parallel))
	}

	return CommandPlan{
		Executable: request.XtraBackupBin,
		Args:       args,
	}, nil
}

func BuildPrepareBackupCommand(request PrepareBackupCommandRequest) (CommandPlan, error) {
	if request.XtraBackupBin == "" {
		return CommandPlan{}, errors.New("xtrabackup binary is required")
	}
	if request.TargetDir == "" {
		return CommandPlan{}, errors.New("target directory is required")
	}

	args := []string{
		"--prepare",
		"--target-dir=" + request.TargetDir,
	}

	if request.ApplyLogOnly {
		args = append(args, "--apply-log-only")
	}
	if request.IncrementalDir != "" {
		args = append(args, "--incremental-dir="+request.IncrementalDir)
	}

	return CommandPlan{
		Executable: request.XtraBackupBin,
		Args:       args,
	}, nil
}

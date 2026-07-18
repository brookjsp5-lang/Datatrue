package tools

import (
	"fmt"
	"path/filepath"
)

type XtraBackupVersion string

const (
	XtraBackupVersion80 XtraBackupVersion = "8.0"
	XtraBackupVersion84 XtraBackupVersion = "8.4"
)

type XtraBackupExecutable string

const (
	XtraBackupExecutableXtraBackup XtraBackupExecutable = "xtrabackup"
	XtraBackupExecutableXbstream   XtraBackupExecutable = "xbstream"
)

func GetXtraBackupExecutable(version XtraBackupVersion, executable XtraBackupExecutable) string {
	return filepath.Join(getXtraBackupBinDir(version), string(executable))
}

func getXtraBackupBinDir(version XtraBackupVersion) string {
	return filepath.Join(
		AssetsToolsDir(),
		"xtrabackup",
		fmt.Sprintf("xtrabackup-%s", version),
		"bin",
	)
}

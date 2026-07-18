package usecases_physical_mysql

import (
	"testing"

	"databasus-backend/internal/util/tools"
)

func Test_ResolveXtraBackupTool_WhenMysql80_ReturnsXtraBackup80(t *testing.T) {
	version, err := ResolveXtraBackupTool(tools.MysqlVersion80)

	if err != nil {
		t.Fatalf("expected MySQL 8.0 to be supported, got error: %v", err)
	}
	if version != tools.XtraBackupVersion80 {
		t.Fatalf("expected XtraBackup 8.0, got %s", version)
	}
}

func Test_ResolveXtraBackupTool_WhenMysql84_ReturnsXtraBackup84(t *testing.T) {
	version, err := ResolveXtraBackupTool(tools.MysqlVersion84)

	if err != nil {
		t.Fatalf("expected MySQL 8.4 to be supported, got error: %v", err)
	}
	if version != tools.XtraBackupVersion84 {
		t.Fatalf("expected XtraBackup 8.4, got %s", version)
	}
}

func Test_ResolveXtraBackupTool_WhenMysql57_ReturnsUnsupported(t *testing.T) {
	_, err := ResolveXtraBackupTool(tools.MysqlVersion57)

	if err == nil {
		t.Fatal("expected MySQL 5.7 to be unsupported for physical backups")
	}
}

func Test_ResolveXtraBackupTool_WhenMysql9_ReturnsUnsupported(t *testing.T) {
	_, err := ResolveXtraBackupTool(tools.MysqlVersion9)

	if err == nil {
		t.Fatal("expected MySQL 9 to be unsupported for physical backups")
	}
}

package usecases_physical_mysql

import (
	"fmt"

	"databasus-backend/internal/util/tools"
)

func ResolveXtraBackupTool(version tools.MysqlVersion) (tools.XtraBackupVersion, error) {
	switch version {
	case tools.MysqlVersion80:
		return tools.XtraBackupVersion80, nil
	case tools.MysqlVersion84:
		return tools.XtraBackupVersion84, nil
	default:
		return "", fmt.Errorf("mysql %s is not supported for physical backups", version)
	}
}

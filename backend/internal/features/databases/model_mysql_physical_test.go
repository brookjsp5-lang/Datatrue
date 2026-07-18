package databases

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mysql_physical "databasus-backend/internal/features/databases/databases/mysql/physical"
	"databasus-backend/internal/util/tools"
)

func validMysqlPhysicalModel() *mysql_physical.MysqlPhysicalDatabase {
	return &mysql_physical.MysqlPhysicalDatabase{
		Version:         tools.MysqlVersion80,
		Host:            "127.0.0.1",
		Port:            3306,
		Username:        "user",
		Password:        "pass",
		DataDir:         "/var/lib/mysql",
		IsBinlogEnabled: true,
	}
}

func validMysqlPhysicalDatabase() *Database {
	return &Database{
		Name:          "MySQL physical source",
		Type:          DatabaseTypeMysqlPhysical,
		MysqlPhysical: validMysqlPhysicalModel(),
	}
}

func Test_Validate_OnMysqlPhysicalDatabase_RequiresMysqlPhysicalConfig(t *testing.T) {
	database := validMysqlPhysicalDatabase()
	database.MysqlPhysical = nil

	err := database.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "mysql physical database is required")
}

func Test_Validate_OnMysqlPhysicalDatabase_DelegatesToMysqlPhysicalConfig(t *testing.T) {
	database := validMysqlPhysicalDatabase()
	database.MysqlPhysical.Host = ""

	err := database.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "host is required")
}

func Test_Validate_OnMysqlPhysicalDatabase_AcceptsValidConfig(t *testing.T) {
	assert.NoError(t, validMysqlPhysicalDatabase().Validate())
}

func Test_ValidateUpdate_OnMysqlPhysicalDatabase_RejectsServerUUIDChange(t *testing.T) {
	oldUUID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	newUUID := "ffffffff-bbbb-cccc-dddd-eeeeeeeeeeee"
	oldDatabase := validMysqlPhysicalDatabase()
	oldDatabase.MysqlPhysical.ServerUUID = &oldUUID
	newDatabase := validMysqlPhysicalDatabase()
	newDatabase.MysqlPhysical.ServerUUID = &newUUID

	err := new(Database).ValidateUpdate(*oldDatabase, *newDatabase)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "immutable")
}

func Test_Update_OnMysqlPhysicalDatabase_UpdatesNestedModel(t *testing.T) {
	existing := validMysqlPhysicalDatabase()
	existing.MysqlPhysical.Password = "kept"

	incoming := validMysqlPhysicalDatabase()
	incoming.Name = "Updated name"
	incoming.MysqlPhysical.Host = "mysql.internal"
	incoming.MysqlPhysical.Password = ""

	existing.Update(incoming)

	assert.Equal(t, "Updated name", existing.Name)
	assert.Equal(t, "mysql.internal", existing.MysqlPhysical.Host)
	assert.Equal(t, "kept", existing.MysqlPhysical.Password)
}

func Test_HideSensitiveData_OnMysqlPhysicalDatabase_ZerosPassword(t *testing.T) {
	database := validMysqlPhysicalDatabase()
	database.MysqlPhysical.Password = "secret"

	database.HideSensitiveData()

	assert.Empty(t, database.MysqlPhysical.Password)
}

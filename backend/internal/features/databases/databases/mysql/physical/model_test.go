package mysql_physical

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"databasus-backend/internal/util/encryption"
	"databasus-backend/internal/util/tools"
)

func validModel() *MysqlPhysicalDatabase {
	return &MysqlPhysicalDatabase{
		Version:         tools.MysqlVersion80,
		Host:            "127.0.0.1",
		Port:            3306,
		Username:        "user",
		Password:        "pass",
		DataDir:         "/var/lib/mysql",
		IsBinlogEnabled: true,
	}
}

func Test_Validate_RejectsBlankHost(t *testing.T) {
	m := validModel()
	m.Host = ""

	assert.Error(t, m.Validate())
}

func Test_Validate_RejectsZeroPort(t *testing.T) {
	m := validModel()
	m.Port = 0

	assert.Error(t, m.Validate())
}

func Test_Validate_RejectsBlankUsername(t *testing.T) {
	m := validModel()
	m.Username = ""

	assert.Error(t, m.Validate())
}

func Test_Validate_RejectsBlankPassword(t *testing.T) {
	m := validModel()
	m.Password = ""

	assert.Error(t, m.Validate())
}

func Test_Validate_RejectsBlankDataDir(t *testing.T) {
	m := validModel()
	m.DataDir = ""

	assert.Error(t, m.Validate())
}

func Test_Validate_AllowsBlankVersionBeforeDiscovery(t *testing.T) {
	m := validModel()
	m.Version = ""

	assert.NoError(t, m.Validate())
}

func Test_Validate_AcceptsSupportedPhysicalVersions(t *testing.T) {
	for _, version := range []tools.MysqlVersion{tools.MysqlVersion80, tools.MysqlVersion84} {
		m := validModel()
		m.Version = version

		assert.NoErrorf(t, m.Validate(), "version %s", version)
	}
}

func Test_Validate_RejectsUnsupportedPhysicalVersion(t *testing.T) {
	for _, version := range []tools.MysqlVersion{tools.MysqlVersion57, tools.MysqlVersion9} {
		m := validModel()
		m.Version = version

		err := m.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "8.0 or 8.4")
	}
}

func Test_ValidateUpdate_RejectsServerUUIDChange(t *testing.T) {
	oldUUID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	newUUID := "ffffffff-bbbb-cccc-dddd-eeeeeeeeeeee"
	old := &MysqlPhysicalDatabase{ServerUUID: &oldUUID}
	next := &MysqlPhysicalDatabase{ServerUUID: &newUUID}

	err := next.ValidateUpdate(old)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "immutable")
}

func Test_ValidateUpdate_AllowsServerUUIDFirstSet(t *testing.T) {
	first := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	old := &MysqlPhysicalDatabase{ServerUUID: nil}
	next := &MysqlPhysicalDatabase{ServerUUID: &first}

	assert.NoError(t, next.ValidateUpdate(old))
}

func Test_ValidateUpdate_AllowsUnchangedServerUUID(t *testing.T) {
	same := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	old := &MysqlPhysicalDatabase{ServerUUID: &same}
	next := &MysqlPhysicalDatabase{ServerUUID: &same}

	assert.NoError(t, next.ValidateUpdate(old))
}

func Test_Update_PreservesPasswordWhenIncomingBlank(t *testing.T) {
	existing := validModel()
	existing.Password = "kept"

	incoming := validModel()
	incoming.Password = ""

	existing.Update(incoming)

	assert.Equal(t, "kept", existing.Password)
}

func Test_Update_OverwritesPasswordWhenIncomingNonBlank(t *testing.T) {
	existing := validModel()
	existing.Password = "old"

	incoming := validModel()
	incoming.Password = "new"

	existing.Update(incoming)

	assert.Equal(t, "new", existing.Password)
}

func Test_Update_NeverOverwritesServerUUIDFromIncoming(t *testing.T) {
	original := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	existing := validModel()
	existing.ServerUUID = &original

	hostile := "ffffffff-bbbb-cccc-dddd-eeeeeeeeeeee"
	incoming := validModel()
	incoming.ServerUUID = &hostile

	existing.Update(incoming)

	require.NotNil(t, existing.ServerUUID)
	assert.Equal(t, original, *existing.ServerUUID)
}

func Test_HideSensitiveData_ZerosPassword(t *testing.T) {
	m := validModel()
	m.Password = "secret"

	m.HideSensitiveData()

	assert.Empty(t, m.Password)
}

func Test_HideSensitiveData_NilReceiver_DoesNotPanic(t *testing.T) {
	var m *MysqlPhysicalDatabase

	assert.NotPanics(t, func() {
		m.HideSensitiveData()
	})
}

type noopEncryptor struct{}

func (noopEncryptor) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (noopEncryptor) Decrypt(s string) (string, error) {
	if len(s) > 4 && s[:4] == "enc:" {
		return s[4:], nil
	}
	return s, nil
}

func Test_EncryptSensitiveFields_EncryptsPassword(t *testing.T) {
	m := validModel()
	m.Password = "p"

	require.NoError(t, m.EncryptSensitiveFields(noopEncryptor{}))

	assert.Equal(t, "enc:p", m.Password)
}

func Test_EncryptSensitiveFields_SkipsBlankPassword(t *testing.T) {
	m := validModel()
	m.Password = ""

	require.NoError(t, m.EncryptSensitiveFields(noopEncryptor{}))

	assert.Empty(t, m.Password)
}

var _ encryption.FieldEncryptor = noopEncryptor{}

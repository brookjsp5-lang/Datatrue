package mysql_physical

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"databasus-backend/internal/util/encryption"
	"databasus-backend/internal/util/tools"
)

type MysqlPhysicalDatabase struct {
	ID         uuid.UUID  `json:"id"         gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DatabaseID *uuid.UUID `json:"databaseId" gorm:"type:uuid;column:database_id"`

	Version tools.MysqlVersion `json:"version" gorm:"type:text;not null"`

	Host     string `json:"host"     gorm:"type:text;not null"`
	Port     int    `json:"port"     gorm:"type:int;not null"`
	Username string `json:"username" gorm:"type:text;not null"`
	Password string `json:"password" gorm:"type:text;not null"`
	IsHttps  bool   `json:"isHttps"  gorm:"type:boolean;default:false"`

	DataDir         string  `json:"dataDir"         gorm:"column:data_dir;type:text;not null"`
	ServerUUID      *string `json:"serverUuid"      gorm:"column:server_uuid;type:text"`
	IsBinlogEnabled bool    `json:"isBinlogEnabled" gorm:"column:is_binlog_enabled;type:boolean;not null;default:false"`
}

func (m *MysqlPhysicalDatabase) TableName() string {
	return "mysql_physical_databases"
}

func (m *MysqlPhysicalDatabase) Validate() error {
	if m.Host == "" {
		return errors.New("host is required")
	}
	if m.Port == 0 {
		return errors.New("port is required")
	}
	if m.Username == "" {
		return errors.New("username is required")
	}
	if m.Password == "" {
		return errors.New("password is required")
	}
	if strings.TrimSpace(m.DataDir) == "" {
		return errors.New("data dir is required")
	}
	if m.Version != "" {
		if err := validatePhysicalVersion(m.Version); err != nil {
			return err
		}
	}
	return nil
}

func (m *MysqlPhysicalDatabase) ValidateUpdate(old *MysqlPhysicalDatabase) error {
	if old == nil {
		return nil
	}

	if old.ServerUUID != nil && m.ServerUUID != nil && *old.ServerUUID != *m.ServerUUID {
		return errors.New("server_uuid is immutable; server swap refused")
	}

	return nil
}

func (m *MysqlPhysicalDatabase) TestPhysicalConnection(
	logger *slog.Logger,
	encryptor encryption.FieldEncryptor,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := m.openConnection(ctx, encryptor)
	if err != nil {
		return err
	}
	defer closeQuietly(db, logger)

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping MySQL server: %w", err)
	}

	if err := m.populateFromConnection(ctx, db); err != nil {
		return err
	}

	if !m.IsBinlogEnabled {
		return errors.New("binary logging must be enabled for MySQL physical PITR")
	}

	return nil
}

func (m *MysqlPhysicalDatabase) PopulateDbData(
	logger *slog.Logger,
	encryptor encryption.FieldEncryptor,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := m.openConnection(ctx, encryptor)
	if err != nil {
		return err
	}
	defer closeQuietly(db, logger)

	return m.populateFromConnection(ctx, db)
}

func (m *MysqlPhysicalDatabase) HideSensitiveData() {
	if m == nil {
		return
	}

	m.Password = ""
}

func (m *MysqlPhysicalDatabase) Update(incoming *MysqlPhysicalDatabase) {
	m.Version = incoming.Version
	m.Host = incoming.Host
	m.Port = incoming.Port
	m.Username = incoming.Username
	m.IsHttps = incoming.IsHttps
	m.DataDir = incoming.DataDir
	m.IsBinlogEnabled = incoming.IsBinlogEnabled

	if incoming.Password != "" {
		m.Password = incoming.Password
	}

	if m.ServerUUID == nil && incoming.ServerUUID != nil {
		m.ServerUUID = incoming.ServerUUID
	}
}

func (m *MysqlPhysicalDatabase) EncryptSensitiveFields(
	encryptor encryption.FieldEncryptor,
) error {
	if m.Password == "" {
		return nil
	}

	encrypted, err := encryptor.Encrypt(m.Password)
	if err != nil {
		return err
	}
	m.Password = encrypted
	return nil
}

func (m *MysqlPhysicalDatabase) openConnection(
	_ context.Context,
	encryptor encryption.FieldEncryptor,
) (*sql.DB, error) {
	password, err := decryptPasswordIfNeeded(m.Password, encryptor)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	db, err := sql.Open("mysql", m.buildDSN(password))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL server: %w", err)
	}

	db.SetConnMaxLifetime(15 * time.Second)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return db, nil
}

func (m *MysqlPhysicalDatabase) buildDSN(password string) string {
	tlsConfig := "false"
	allowCleartext := ""

	if m.IsHttps {
		err := mysqldriver.RegisterTLSConfig("mysql-physical-skip-verify", &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			_ = err
		}

		tlsConfig = "mysql-physical-skip-verify"
		allowCleartext = "&allowCleartextPasswords=1"
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/?parseTime=true&timeout=15s&tls=%s&charset=utf8mb4%s",
		m.Username,
		password,
		m.Host,
		m.Port,
		tlsConfig,
		allowCleartext,
	)
}

func (m *MysqlPhysicalDatabase) populateFromConnection(ctx context.Context, db *sql.DB) error {
	detectedVersion, err := detectMysqlPhysicalVersion(ctx, db)
	if err != nil {
		return err
	}
	m.Version = detectedVersion

	if m.ServerUUID == nil {
		serverUUID, err := detectServerUUID(ctx, db)
		if err != nil {
			return err
		}
		m.ServerUUID = &serverUUID
	}

	binlogEnabled, err := detectBinlogEnabled(ctx, db)
	if err != nil {
		return err
	}
	m.IsBinlogEnabled = binlogEnabled

	return nil
}

var mysqlVersionRegexp = regexp.MustCompile(`^(\d+)\.(\d+)`)

func detectMysqlPhysicalVersion(ctx context.Context, db *sql.DB) (tools.MysqlVersion, error) {
	var versionStr string
	if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&versionStr); err != nil {
		return "", fmt.Errorf("failed to query MySQL version: %w", err)
	}

	matches := mysqlVersionRegexp.FindStringSubmatch(versionStr)
	if len(matches) < 3 {
		return "", fmt.Errorf("could not parse MySQL version: %s", versionStr)
	}

	return mapMysqlPhysicalVersion(matches[1], matches[2])
}

func mapMysqlPhysicalVersion(major, minor string) (tools.MysqlVersion, error) {
	switch major {
	case "8":
		switch minor {
		case "0", "1", "2", "3":
			return tools.MysqlVersion80, nil
		default:
			return tools.MysqlVersion84, nil
		}
	default:
		return "", fmt.Errorf("MySQL physical backup requires MySQL 8.0 or 8.4, detected %s.%s", major, minor)
	}
}

func detectServerUUID(ctx context.Context, db *sql.DB) (string, error) {
	var uuidValue string
	if err := db.QueryRowContext(ctx, "SELECT @@server_uuid").Scan(&uuidValue); err != nil {
		return "", fmt.Errorf("failed to read server_uuid: %w", err)
	}
	return uuidValue, nil
}

func detectBinlogEnabled(ctx context.Context, db *sql.DB) (bool, error) {
	var name, value string
	if err := db.QueryRowContext(ctx, "SHOW VARIABLES LIKE 'log_bin'").Scan(&name, &value); err != nil {
		return false, fmt.Errorf("failed to read log_bin: %w", err)
	}

	switch strings.ToLower(value) {
	case "on", "1", "true":
		return true, nil
	default:
		return false, nil
	}
}

func validatePhysicalVersion(version tools.MysqlVersion) error {
	switch version {
	case tools.MysqlVersion80, tools.MysqlVersion84:
		return nil
	default:
		return fmt.Errorf("MySQL physical backup requires MySQL 8.0 or 8.4, got %s", version)
	}
}

func closeQuietly(db *sql.DB, logger *slog.Logger) {
	if err := db.Close(); err != nil {
		logger.Error("Failed to close MySQL physical connection", "error", err)
	}
}

func decryptPasswordIfNeeded(
	password string,
	encryptor encryption.FieldEncryptor,
) (string, error) {
	if encryptor == nil {
		return password, nil
	}
	return encryptor.Decrypt(password)
}

package physicalplan

import (
	"errors"
	"fmt"
	"slices"

	"databasus-backend/internal/util/tools"
)

type PhysicalRestorePlanStep string

const (
	PhysicalRestorePlanStepStopMysql         PhysicalRestorePlanStep = "STOP_MYSQL"
	PhysicalRestorePlanStepQuarantineDataDir PhysicalRestorePlanStep = "QUARANTINE_DATA_DIR"
	PhysicalRestorePlanStepFetchFullBackup   PhysicalRestorePlanStep = "FETCH_FULL_BACKUP"
	PhysicalRestorePlanStepExtractXbstream   PhysicalRestorePlanStep = "EXTRACT_XBSTREAM"
	PhysicalRestorePlanStepPrepareBackup     PhysicalRestorePlanStep = "PREPARE_BACKUP"
	PhysicalRestorePlanStepReplaceDataDir    PhysicalRestorePlanStep = "REPLACE_DATA_DIR"
	PhysicalRestorePlanStepReplayBinlogs     PhysicalRestorePlanStep = "REPLAY_BINLOGS"
	PhysicalRestorePlanStepStartMysql        PhysicalRestorePlanStep = "START_MYSQL"
	PhysicalRestorePlanStepHealthcheck       PhysicalRestorePlanStep = "HEALTHCHECK"
)

type PhysicalRestoreArtifact struct {
	FileName  string
	SizeBytes int64
}

type PhysicalRestoreAgentCapabilities struct {
	SupportedMysqlVersions []tools.MysqlVersion
	FreeBytes              int64
	HasXtraBackup          bool
	HasXbstream            bool
	HasMysqlClient         bool
	CanControlMysqlService bool
	CanWriteDataDir        bool
}

type PhysicalRestorePlanRequest struct {
	AgentID            string
	BackupVersion      tools.MysqlVersion
	TargetVersion      tools.MysqlVersion
	FullBackupArtifact PhysicalRestoreArtifact
	BinlogArtifacts    []PhysicalRestoreArtifact
	RequiredBytes      int64
	Capabilities       PhysicalRestoreAgentCapabilities
}

type PhysicalRestorePlan struct {
	AgentID        string
	BackupVersion  tools.MysqlVersion
	TargetVersion  tools.MysqlVersion
	FullBackupFile string
	BinlogFiles    []string
	RequiredBytes  int64
	AvailableBytes int64
	Steps          []PhysicalRestorePlanStep
}

func BuildPhysicalRestorePlan(request PhysicalRestorePlanRequest) (PhysicalRestorePlan, error) {
	if err := validatePhysicalRestoreRequest(request); err != nil {
		return PhysicalRestorePlan{}, err
	}

	binlogFiles := make([]string, 0, len(request.BinlogArtifacts))
	for _, artifact := range request.BinlogArtifacts {
		binlogFiles = append(binlogFiles, artifact.FileName)
	}

	steps := []PhysicalRestorePlanStep{
		PhysicalRestorePlanStepStopMysql,
		PhysicalRestorePlanStepQuarantineDataDir,
		PhysicalRestorePlanStepFetchFullBackup,
		PhysicalRestorePlanStepExtractXbstream,
		PhysicalRestorePlanStepPrepareBackup,
		PhysicalRestorePlanStepReplaceDataDir,
	}
	if len(binlogFiles) > 0 {
		steps = append(steps, PhysicalRestorePlanStepReplayBinlogs)
	}
	steps = append(steps,
		PhysicalRestorePlanStepStartMysql,
		PhysicalRestorePlanStepHealthcheck,
	)

	return PhysicalRestorePlan{
		AgentID:        request.AgentID,
		BackupVersion:  request.BackupVersion,
		TargetVersion:  request.TargetVersion,
		FullBackupFile: request.FullBackupArtifact.FileName,
		BinlogFiles:    binlogFiles,
		RequiredBytes:  request.RequiredBytes,
		AvailableBytes: request.Capabilities.FreeBytes,
		Steps:          steps,
	}, nil
}

func validatePhysicalRestoreRequest(request PhysicalRestorePlanRequest) error {
	if request.AgentID == "" {
		return errors.New("agent id is required")
	}
	if request.FullBackupArtifact.FileName == "" {
		return errors.New("full backup artifact file name is required")
	}
	if request.BackupVersion != request.TargetVersion {
		return fmt.Errorf(
			"physical restore requires matching MySQL versions, backup %s target %s",
			request.BackupVersion,
			request.TargetVersion,
		)
	}
	if !isSupportedPhysicalRestoreVersion(request.TargetVersion) {
		return fmt.Errorf("mysql %s is not supported for physical restore", request.TargetVersion)
	}
	if !slices.Contains(request.Capabilities.SupportedMysqlVersions, request.TargetVersion) {
		return fmt.Errorf("agent does not support mysql %s", request.TargetVersion)
	}
	if request.RequiredBytes > request.Capabilities.FreeBytes {
		return fmt.Errorf(
			"agent free space %d is below required %d",
			request.Capabilities.FreeBytes,
			request.RequiredBytes,
		)
	}
	if !request.Capabilities.HasXtraBackup {
		return errors.New("agent must provide xtrabackup")
	}
	if !request.Capabilities.HasXbstream {
		return errors.New("agent must provide xbstream")
	}
	if !request.Capabilities.HasMysqlClient {
		return errors.New("agent must provide mysql client")
	}
	if !request.Capabilities.CanControlMysqlService {
		return errors.New("agent must be able to control mysql service")
	}
	if !request.Capabilities.CanWriteDataDir {
		return errors.New("agent must be able to write mysql data directory")
	}

	for _, artifact := range request.BinlogArtifacts {
		if artifact.FileName == "" {
			return errors.New("binlog artifact file name is required")
		}
	}

	return nil
}

func isSupportedPhysicalRestoreVersion(version tools.MysqlVersion) bool {
	return version == tools.MysqlVersion80 || version == tools.MysqlVersion84
}

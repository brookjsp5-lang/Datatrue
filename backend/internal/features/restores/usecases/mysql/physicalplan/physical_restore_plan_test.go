package physicalplan

import (
	"slices"
	"testing"

	"databasus-backend/internal/util/tools"
)

func Test_BuildPhysicalRestorePlan_WhenAgentSupportsTarget_ReturnsAuditableSteps(t *testing.T) {
	plan, err := BuildPhysicalRestorePlan(PhysicalRestorePlanRequest{
		AgentID:       "agent-1",
		BackupVersion: tools.MysqlVersion84,
		TargetVersion: tools.MysqlVersion84,
		FullBackupArtifact: PhysicalRestoreArtifact{
			FileName:  "mysql/full.xbstream.zst",
			SizeBytes: 1024,
		},
		BinlogArtifacts: []PhysicalRestoreArtifact{
			{FileName: "mysql-bin.000001", SizeBytes: 128},
		},
		RequiredBytes: 2048,
		Capabilities: PhysicalRestoreAgentCapabilities{
			SupportedMysqlVersions: []tools.MysqlVersion{tools.MysqlVersion80, tools.MysqlVersion84},
			FreeBytes:              4096,
			HasXtraBackup:          true,
			HasXbstream:            true,
			HasMysqlClient:         true,
			CanControlMysqlService: true,
			CanWriteDataDir:        true,
		},
	})

	if err != nil {
		t.Fatalf("expected restore plan, got error: %v", err)
	}

	expectedSteps := []PhysicalRestorePlanStep{
		PhysicalRestorePlanStepStopMysql,
		PhysicalRestorePlanStepQuarantineDataDir,
		PhysicalRestorePlanStepFetchFullBackup,
		PhysicalRestorePlanStepExtractXbstream,
		PhysicalRestorePlanStepPrepareBackup,
		PhysicalRestorePlanStepReplaceDataDir,
		PhysicalRestorePlanStepReplayBinlogs,
		PhysicalRestorePlanStepStartMysql,
		PhysicalRestorePlanStepHealthcheck,
	}
	if !slices.Equal(plan.Steps, expectedSteps) {
		t.Fatalf("expected steps %v, got %v", expectedSteps, plan.Steps)
	}
	if plan.AgentID != "agent-1" {
		t.Fatalf("expected agent id to be preserved, got %s", plan.AgentID)
	}
}

func Test_BuildPhysicalRestorePlan_WhenAgentLacksXtraBackup_ReturnsError(t *testing.T) {
	_, err := BuildPhysicalRestorePlan(PhysicalRestorePlanRequest{
		AgentID:       "agent-1",
		BackupVersion: tools.MysqlVersion80,
		TargetVersion: tools.MysqlVersion80,
		FullBackupArtifact: PhysicalRestoreArtifact{
			FileName:  "mysql/full.xbstream.zst",
			SizeBytes: 1024,
		},
		RequiredBytes: 1024,
		Capabilities: PhysicalRestoreAgentCapabilities{
			SupportedMysqlVersions: []tools.MysqlVersion{tools.MysqlVersion80},
			FreeBytes:              4096,
			HasXbstream:            true,
			HasMysqlClient:         true,
			CanControlMysqlService: true,
			CanWriteDataDir:        true,
		},
	})

	if err == nil {
		t.Fatal("expected missing xtrabackup capability to be rejected")
	}
}

func Test_BuildPhysicalRestorePlan_WhenTargetVersionDiffers_ReturnsError(t *testing.T) {
	_, err := BuildPhysicalRestorePlan(PhysicalRestorePlanRequest{
		AgentID:       "agent-1",
		BackupVersion: tools.MysqlVersion84,
		TargetVersion: tools.MysqlVersion80,
		FullBackupArtifact: PhysicalRestoreArtifact{
			FileName:  "mysql/full.xbstream.zst",
			SizeBytes: 1024,
		},
		RequiredBytes: 1024,
		Capabilities: PhysicalRestoreAgentCapabilities{
			SupportedMysqlVersions: []tools.MysqlVersion{tools.MysqlVersion80, tools.MysqlVersion84},
			FreeBytes:              4096,
			HasXtraBackup:          true,
			HasXbstream:            true,
			HasMysqlClient:         true,
			CanControlMysqlService: true,
			CanWriteDataDir:        true,
		},
	})

	if err == nil {
		t.Fatal("expected cross-version physical restore to be rejected")
	}
}

func Test_BuildPhysicalRestorePlan_WhenFreeSpaceInsufficient_ReturnsError(t *testing.T) {
	_, err := BuildPhysicalRestorePlan(PhysicalRestorePlanRequest{
		AgentID:       "agent-1",
		BackupVersion: tools.MysqlVersion80,
		TargetVersion: tools.MysqlVersion80,
		FullBackupArtifact: PhysicalRestoreArtifact{
			FileName:  "mysql/full.xbstream.zst",
			SizeBytes: 1024,
		},
		RequiredBytes: 4096,
		Capabilities: PhysicalRestoreAgentCapabilities{
			SupportedMysqlVersions: []tools.MysqlVersion{tools.MysqlVersion80},
			FreeBytes:              1024,
			HasXtraBackup:          true,
			HasXbstream:            true,
			HasMysqlClient:         true,
			CanControlMysqlService: true,
			CanWriteDataDir:        true,
		},
	})

	if err == nil {
		t.Fatal("expected insufficient free space to be rejected")
	}
}

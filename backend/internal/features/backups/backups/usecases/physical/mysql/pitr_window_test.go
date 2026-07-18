package usecases_physical_mysql

import (
	"testing"
	"time"
)

func Test_ValidatePITRWindow_WhenTargetIsInsideCapturedBinlogs_ReturnsNil(t *testing.T) {
	fullBackupFinishedAt := time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)
	targetTime := fullBackupFinishedAt.Add(30 * time.Minute)

	err := ValidatePITRWindow(PITRWindowRequest{
		FullBackupFinishedAt: fullBackupFinishedAt,
		TargetTime:           targetTime,
		BinlogArtifacts: []BinlogArtifact{
			{
				FileName:  "mysql-bin.000001",
				StartedAt: fullBackupFinishedAt,
				EndedAt:   fullBackupFinishedAt.Add(time.Hour),
			},
		},
	})

	if err != nil {
		t.Fatalf("expected PITR window to be valid, got error: %v", err)
	}
}

func Test_ValidatePITRWindow_WhenTargetBeforeFullBackup_ReturnsError(t *testing.T) {
	fullBackupFinishedAt := time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)

	err := ValidatePITRWindow(PITRWindowRequest{
		FullBackupFinishedAt: fullBackupFinishedAt,
		TargetTime:           fullBackupFinishedAt.Add(-time.Minute),
	})

	if err == nil {
		t.Fatal("expected target time before full backup to be rejected")
	}
}

func Test_ValidatePITRWindow_WhenBinlogCoverageMissing_ReturnsError(t *testing.T) {
	fullBackupFinishedAt := time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)

	err := ValidatePITRWindow(PITRWindowRequest{
		FullBackupFinishedAt: fullBackupFinishedAt,
		TargetTime:           fullBackupFinishedAt.Add(2 * time.Hour),
		BinlogArtifacts: []BinlogArtifact{
			{
				FileName:  "mysql-bin.000001",
				StartedAt: fullBackupFinishedAt,
				EndedAt:   fullBackupFinishedAt.Add(time.Hour),
			},
		},
	})

	if err == nil {
		t.Fatal("expected missing binlog coverage to be rejected")
	}
}

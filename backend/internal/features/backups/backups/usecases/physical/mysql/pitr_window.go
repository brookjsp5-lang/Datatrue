package usecases_physical_mysql

import (
	"errors"
	"fmt"
	"time"
)

type BinlogArtifact struct {
	FileName  string
	StartedAt time.Time
	EndedAt   time.Time
}

type PITRWindowRequest struct {
	FullBackupFinishedAt time.Time
	TargetTime           time.Time
	BinlogArtifacts      []BinlogArtifact
}

func ValidatePITRWindow(request PITRWindowRequest) error {
	if request.TargetTime.Before(request.FullBackupFinishedAt) {
		return errors.New("target time must be after the full backup finished")
	}
	if len(request.BinlogArtifacts) == 0 {
		return errors.New("binlog artifacts are required for PITR")
	}

	for _, artifact := range request.BinlogArtifacts {
		if artifact.FileName == "" {
			return errors.New("binlog artifact file name is required")
		}
		if isTargetCoveredByArtifact(request.TargetTime, artifact) {
			return nil
		}
	}

	return fmt.Errorf("no binlog artifact covers target time %s", request.TargetTime.Format(time.RFC3339))
}

func isTargetCoveredByArtifact(targetTime time.Time, artifact BinlogArtifact) bool {
	if targetTime.Before(artifact.StartedAt) {
		return false
	}

	return artifact.EndedAt.IsZero() || !targetTime.After(artifact.EndedAt)
}

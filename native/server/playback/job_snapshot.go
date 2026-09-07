package playback

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"time"
)

const (
	TransformJobSnapshotSchemaVersion = 1
	MaxTransformJobSnapshotBytes      = 64 * 1024
)

var (
	ErrInvalidTransformJobSnapshot     = errors.New("invalid transform job snapshot")
	ErrUnsupportedTransformJobSnapshot = errors.New("unsupported transform job snapshot version")
)

type persistedTransformJob struct {
	SchemaVersion int                    `json:"schemaVersion"`
	ID            string                 `json:"id"`
	SourceID      string                 `json:"sourceId"`
	Requirements  []TransformRequirement `json:"requirements"`
	State         TransformJobState      `json:"state"`
	UpdatedAt     string                 `json:"updatedAt"`
	Attempts      int                    `json:"attempts"`
	FailureCode   string                 `json:"failureCode"`
}

// EncodeTransformJobSnapshot serializes one validated worker-neutral lifecycle
// record without selecting a scheduler, database, queue backend, or worker.
func EncodeTransformJobSnapshot(job TransformJob) ([]byte, error) {
	if !validPersistableTransformJob(job) {
		return nil, ErrInvalidTransformJobSnapshot
	}
	data, err := json.Marshal(persistedTransformJob{
		SchemaVersion: TransformJobSnapshotSchemaVersion,
		ID:            job.id,
		SourceID:      job.request.sourceID,
		Requirements:  append([]TransformRequirement(nil), job.request.requirements...),
		State:         job.state,
		UpdatedAt:     job.updatedAt.UTC().Format(time.RFC3339Nano),
		Attempts:      job.attempts,
		FailureCode:   job.failureCode,
	})
	if err != nil || len(data) > MaxTransformJobSnapshotBytes {
		return nil, ErrInvalidTransformJobSnapshot
	}
	return data, nil
}

// DecodeTransformJobSnapshot accepts only the exact current schema, rejects
// unknown/trailing fields, and reconstructs state through the same lifecycle
// invariants used by live TransformJob values.
func DecodeTransformJobSnapshot(data []byte) (TransformJob, error) {
	if len(data) == 0 || len(data) > MaxTransformJobSnapshotBytes {
		return TransformJob{}, ErrInvalidTransformJobSnapshot
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var persisted persistedTransformJob
	if err := decoder.Decode(&persisted); err != nil {
		return TransformJob{}, ErrInvalidTransformJobSnapshot
	}
	if persisted.SchemaVersion != TransformJobSnapshotSchemaVersion {
		return TransformJob{}, ErrUnsupportedTransformJobSnapshot
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return TransformJob{}, ErrInvalidTransformJobSnapshot
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, persisted.UpdatedAt)
	if err != nil || persisted.UpdatedAt != updatedAt.UTC().Format(time.RFC3339Nano) {
		return TransformJob{}, ErrInvalidTransformJobSnapshot
	}
	job := TransformJob{
		id: persisted.ID,
		request: TransformRequest{
			sourceID:     persisted.SourceID,
			requirements: append([]TransformRequirement(nil), persisted.Requirements...),
		},
		state:       persisted.State,
		updatedAt:   updatedAt.UTC(),
		attempts:    persisted.Attempts,
		failureCode: persisted.FailureCode,
	}
	if !validPersistableTransformJob(job) {
		return TransformJob{}, ErrInvalidTransformJobSnapshot
	}
	return job, nil
}

func validPersistableTransformJob(job TransformJob) bool {
	if !validTransformJob(job) {
		return false
	}
	switch job.state {
	case TransformJobRunning, TransformJobSucceeded, TransformJobFailed:
		return job.attempts > 0
	default:
		return true
	}
}

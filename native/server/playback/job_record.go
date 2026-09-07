package playback

import (
	"errors"
	"math"
)

type TransformJobRevision uint64

var (
	ErrStaleTransformJobRevision     = errors.New("stale transform job revision")
	ErrTransformJobRevisionExhausted = errors.New("transform job revision exhausted")
	ErrInvalidTransformJobRecord     = errors.New("invalid transform job record")
)

// TransformJobRecord is a storage-neutral optimistic-concurrency value for one
// transform job snapshot. It does not select a database, scheduler, or queue.
type TransformJobRecord struct {
	revision TransformJobRevision
	payload  []byte
}

func NewTransformJobRecord(job TransformJob) (TransformJobRecord, error) {
	payload, err := EncodeTransformJobSnapshot(job)
	if err != nil {
		return TransformJobRecord{}, err
	}
	return TransformJobRecord{revision: 0, payload: append([]byte(nil), payload...)}, nil
}

// RestoreTransformJobRecord validates externally stored bytes before accepting
// them into the playback core. Stored revisions remain opaque counters; payload
// validity is governed by the strict transform-job snapshot contract.
func RestoreTransformJobRecord(revision TransformJobRevision, payload []byte) (TransformJobRecord, error) {
	if len(payload) == 0 {
		return TransformJobRecord{}, ErrInvalidTransformJobRecord
	}
	if _, err := DecodeTransformJobSnapshot(payload); err != nil {
		return TransformJobRecord{}, ErrInvalidTransformJobRecord
	}
	return TransformJobRecord{revision: revision, payload: append([]byte(nil), payload...)}, nil
}

func (r TransformJobRecord) Revision() TransformJobRevision {
	return r.revision
}

func (r TransformJobRecord) Payload() []byte {
	return append([]byte(nil), r.payload...)
}

func (r TransformJobRecord) Restore() (TransformJob, error) {
	return DecodeTransformJobSnapshot(r.payload)
}

// Update replaces the stored lifecycle snapshot only when expected matches the
// current revision and the replacement keeps the same immutable job identity.
func (r TransformJobRecord) Update(expected TransformJobRevision, job TransformJob) (TransformJobRecord, error) {
	if expected != r.revision {
		return r, ErrStaleTransformJobRevision
	}
	if r.revision == TransformJobRevision(math.MaxUint64) {
		return r, ErrTransformJobRevisionExhausted
	}
	current, err := r.Restore()
	if err != nil {
		return r, ErrInvalidTransformJobRecord
	}
	if current.ID() != job.ID() {
		return r, ErrInvalidTransformJobRecord
	}
	payload, err := EncodeTransformJobSnapshot(job)
	if err != nil {
		return r, err
	}
	return TransformJobRecord{
		revision: r.revision + 1,
		payload:  append([]byte(nil), payload...),
	}, nil
}

package playback

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const MaxTransformJobFileBytes = 128 * 1024

var (
	ErrInvalidTransformJobFileRepository = errors.New("invalid transform job file repository")
	ErrTransformJobAlreadyExists         = errors.New("transform job already exists")
)

type transformJobFileEnvelope struct {
	Revision TransformJobRevision `json:"revision"`
	Payload  []byte               `json:"payload"`
}

// FileTransformJobRepository is a durable single-process implementation of
// TransformJobRepository. Job IDs are mapped to hashed filenames so opaque IDs
// never become paths. Canonical snapshots are wrapped with their optimistic
// revision in bounded JSON files. Replacement writes use fsync and same-folder
// atomic rename to avoid exposing partial persisted state.
//
// The mutex provides revision-CAS correctness for one GoreeCloud Video server
// process. This adapter does not claim cross-process/distributed locking; a
// deployment with concurrent server processes must use an external
// transactional repository.
type FileTransformJobRepository struct {
	root string
	mu   sync.Mutex
}

func NewFileTransformJobRepository(root string) (*FileTransformJobRepository, error) {
	if strings.TrimSpace(root) == "" {
		return nil, ErrInvalidTransformJobFileRepository
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, ErrInvalidTransformJobFileRepository
	}
	if err := os.MkdirAll(absolute, 0o700); err != nil {
		return nil, fmt.Errorf("create transform job directory: %w", err)
	}
	if err := os.Chmod(absolute, 0o700); err != nil {
		return nil, fmt.Errorf("protect transform job directory: %w", err)
	}
	return &FileTransformJobRepository{root: absolute}, nil
}

func (r *FileTransformJobRepository) Load(
	ctx context.Context,
	id string,
) (TransformJobRecord, bool, error) {
	if r == nil || !validTransformJobID(id) {
		return TransformJobRecord{}, false, ErrInvalidTransformJobFileRepository
	}
	if err := ctx.Err(); err != nil {
		return TransformJobRecord{}, false, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.loadLocked(ctx, id)
}

func (r *FileTransformJobRepository) Create(
	ctx context.Context,
	job TransformJob,
) (TransformJobRecord, error) {
	if r == nil || !validTransformJob(job) {
		return TransformJobRecord{}, ErrInvalidTransformJobFileRepository
	}
	if err := ctx.Err(); err != nil {
		return TransformJobRecord{}, err
	}
	record, err := NewTransformJobRecord(job)
	if err != nil {
		return TransformJobRecord{}, err
	}
	encoded, err := encodeTransformJobFileEnvelope(record)
	if err != nil {
		return TransformJobRecord{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return TransformJobRecord{}, err
	}
	path := r.recordPath(job.ID())
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return TransformJobRecord{}, ErrTransformJobAlreadyExists
	}
	if err != nil {
		return TransformJobRecord{}, fmt.Errorf("create transform job file: %w", err)
	}
	writeErr := writeAndSyncTransformJobFile(file, encoded)
	closeErr := file.Close()
	if writeErr != nil {
		_ = os.Remove(path)
		return TransformJobRecord{}, writeErr
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return TransformJobRecord{}, fmt.Errorf("close transform job file: %w", closeErr)
	}
	if err := syncTransformJobDirectory(r.root); err != nil {
		return TransformJobRecord{}, err
	}
	return record, nil
}

func (r *FileTransformJobRepository) Save(
	ctx context.Context,
	expected TransformJobRevision,
	job TransformJob,
) (TransformJobRecord, error) {
	if r == nil || !validTransformJob(job) {
		return TransformJobRecord{}, ErrInvalidTransformJobFileRepository
	}
	if err := ctx.Err(); err != nil {
		return TransformJobRecord{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	current, found, err := r.loadLocked(ctx, job.ID())
	if err != nil {
		return TransformJobRecord{}, err
	}
	if !found {
		return TransformJobRecord{}, ErrTransformJobRecordNotFound
	}
	if current.Revision() != expected {
		return current, ErrStaleTransformJobRevision
	}
	next, err := current.Update(expected, job)
	if err != nil {
		return current, err
	}
	encoded, err := encodeTransformJobFileEnvelope(next)
	if err != nil {
		return current, err
	}
	if err := ctx.Err(); err != nil {
		return current, err
	}
	if err := r.replaceLocked(r.recordPath(job.ID()), encoded); err != nil {
		return current, err
	}
	return next, nil
}

func (r *FileTransformJobRepository) loadLocked(
	ctx context.Context,
	id string,
) (TransformJobRecord, bool, error) {
	if err := ctx.Err(); err != nil {
		return TransformJobRecord{}, false, err
	}
	file, err := os.Open(r.recordPath(id))
	if errors.Is(err, os.ErrNotExist) {
		return TransformJobRecord{}, false, nil
	}
	if err != nil {
		return TransformJobRecord{}, false, fmt.Errorf("open transform job file: %w", err)
	}
	defer file.Close()
	payload, err := io.ReadAll(io.LimitReader(file, MaxTransformJobFileBytes+1))
	if err != nil {
		return TransformJobRecord{}, false, fmt.Errorf("read transform job file: %w", err)
	}
	if len(payload) == 0 || len(payload) > MaxTransformJobFileBytes {
		return TransformJobRecord{}, false, ErrInvalidTransformJobRecord
	}
	record, err := decodeTransformJobFileEnvelope(payload)
	if err != nil {
		return TransformJobRecord{}, false, err
	}
	job, err := record.Restore()
	if err != nil || job.ID() != id {
		return TransformJobRecord{}, false, ErrInvalidTransformJobRecord
	}
	if err := ctx.Err(); err != nil {
		return TransformJobRecord{}, false, err
	}
	return record, true, nil
}

func (r *FileTransformJobRepository) replaceLocked(path string, encoded []byte) error {
	temporary, err := os.CreateTemp(r.root, ".transform-job-*.tmp")
	if err != nil {
		return fmt.Errorf("create transform job temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	removeTemporary := true
	defer func() {
		_ = temporary.Close()
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return fmt.Errorf("protect transform job temporary file: %w", err)
	}
	if err := writeAndSyncTransformJobFile(temporary, encoded); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close transform job temporary file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace transform job file: %w", err)
	}
	removeTemporary = false
	return syncTransformJobDirectory(r.root)
}

func (r *FileTransformJobRepository) recordPath(id string) string {
	digest := sha256.Sum256([]byte(id))
	return filepath.Join(r.root, hex.EncodeToString(digest[:])+".json")
}

func encodeTransformJobFileEnvelope(record TransformJobRecord) ([]byte, error) {
	if _, err := record.Restore(); err != nil {
		return nil, ErrInvalidTransformJobRecord
	}
	payload, err := json.Marshal(transformJobFileEnvelope{
		Revision: record.Revision(),
		Payload:  record.Payload(),
	})
	if err != nil || len(payload) == 0 || len(payload) > MaxTransformJobFileBytes {
		return nil, ErrInvalidTransformJobRecord
	}
	return payload, nil
}

func decodeTransformJobFileEnvelope(payload []byte) (TransformJobRecord, error) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var envelope transformJobFileEnvelope
	if err := decoder.Decode(&envelope); err != nil {
		return TransformJobRecord{}, ErrInvalidTransformJobRecord
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return TransformJobRecord{}, ErrInvalidTransformJobRecord
	}
	record, err := RestoreTransformJobRecord(envelope.Revision, envelope.Payload)
	if err != nil {
		return TransformJobRecord{}, ErrInvalidTransformJobRecord
	}
	canonical, err := json.Marshal(transformJobFileEnvelope{
		Revision: record.Revision(),
		Payload:  record.Payload(),
	})
	if err != nil || !bytes.Equal(payload, canonical) {
		return TransformJobRecord{}, ErrInvalidTransformJobRecord
	}
	return record, nil
}

func writeAndSyncTransformJobFile(file *os.File, payload []byte) error {
	if len(payload) == 0 || len(payload) > MaxTransformJobFileBytes {
		return ErrInvalidTransformJobRecord
	}
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("truncate transform job file: %w", err)
	}
	if _, err := file.Write(payload); err != nil {
		return fmt.Errorf("write transform job file: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync transform job file: %w", err)
	}
	return nil
}

func syncTransformJobDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open transform job directory: %w", err)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync transform job directory: %w", err)
	}
	return nil
}

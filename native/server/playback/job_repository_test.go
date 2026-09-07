package playback

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeTransformJobRepository struct {
	record       TransformJobRecord
	found        bool
	loadErr      error
	createErr    error
	saveErr      error
	createResult *TransformJobRecord
	saveResult   *TransformJobRecord
	creates      int
	saves        int
}

func (f *fakeTransformJobRepository) Load(context.Context, string) (TransformJobRecord, bool, error) {
	return f.record, f.found, f.loadErr
}

func (f *fakeTransformJobRepository) Create(
	_ context.Context,
	job TransformJob,
) (TransformJobRecord, error) {
	f.creates++
	if f.createErr != nil {
		return TransformJobRecord{}, f.createErr
	}
	if f.createResult != nil {
		return *f.createResult, nil
	}
	record, err := NewTransformJobRecord(job)
	if err != nil {
		return TransformJobRecord{}, err
	}
	f.record = record
	f.found = true
	return record, nil
}

func (f *fakeTransformJobRepository) Save(
	_ context.Context,
	expected TransformJobRevision,
	job TransformJob,
) (TransformJobRecord, error) {
	f.saves++
	if f.saveErr != nil {
		return f.record, f.saveErr
	}
	if f.saveResult != nil {
		return *f.saveResult, nil
	}
	next, err := f.record.Update(expected, job)
	if err != nil {
		return f.record, err
	}
	f.record = next
	return next, nil
}

func TestInitializeStoredTransformJobValidatesRepositoryResult(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 40, 0, 0, time.UTC)
	job := testTransformJob(t, "job-1", now)
	repository := &fakeTransformJobRepository{}

	record, err := InitializeStoredTransformJob(context.Background(), repository, job)
	if err != nil || record.Revision() != 0 || repository.creates != 1 {
		t.Fatalf("record=%+v creates=%d err=%v", record, repository.creates, err)
	}

	wrongJob := testTransformJob(t, "job-2", now)
	wrongRecord, err := NewTransformJobRecord(wrongJob)
	if err != nil {
		t.Fatal(err)
	}
	repository = &fakeTransformJobRepository{createResult: &wrongRecord}
	if _, err := InitializeStoredTransformJob(context.Background(), repository, job); !errors.Is(err, ErrInvalidTransformJobRepositoryResult) {
		t.Fatalf("error=%v", err)
	}
}

func TestMutateStoredTransformJobSavesCanonicalNextRevision(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 40, 0, 0, time.UTC)
	job := testTransformJob(t, "job-1", now)
	record, err := NewTransformJobRecord(job)
	if err != nil {
		t.Fatal(err)
	}
	repository := &fakeTransformJobRepository{record: record, found: true}

	next, err := MutateStoredTransformJob(context.Background(), repository, "job-1", func(current TransformJob) (TransformJob, error) {
		return current.Transition(TransformJobRunning, now.Add(time.Second), "")
	})
	if err != nil || next.Revision() != 1 || repository.saves != 1 {
		t.Fatalf("next=%+v saves=%d err=%v", next, repository.saves, err)
	}
	restored, err := next.Restore()
	if err != nil || restored.State() != TransformJobRunning || restored.Attempts() != 1 {
		t.Fatalf("restored=%+v err=%v", restored, err)
	}
}

func TestMutateStoredTransformJobRejectsWrongLoadAndSaveResults(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 40, 0, 0, time.UTC)
	job := testTransformJob(t, "job-1", now)
	record, err := NewTransformJobRecord(job)
	if err != nil {
		t.Fatal(err)
	}
	wrongJob := testTransformJob(t, "job-2", now)
	wrongRecord, err := NewTransformJobRecord(wrongJob)
	if err != nil {
		t.Fatal(err)
	}

	repository := &fakeTransformJobRepository{record: wrongRecord, found: true}
	if _, err := MutateStoredTransformJob(context.Background(), repository, "job-1", func(current TransformJob) (TransformJob, error) {
		return current, nil
	}); !errors.Is(err, ErrInvalidTransformJobRepositoryResult) {
		t.Fatalf("wrong load error=%v", err)
	}

	repository = &fakeTransformJobRepository{record: record, found: true, saveResult: &wrongRecord}
	if _, err := MutateStoredTransformJob(context.Background(), repository, "job-1", func(current TransformJob) (TransformJob, error) {
		return current.Transition(TransformJobRunning, now.Add(time.Second), "")
	}); !errors.Is(err, ErrInvalidTransformJobRepositoryResult) {
		t.Fatalf("wrong save error=%v", err)
	}
}

func TestMutateStoredTransformJobPropagatesNotFoundMutationAndContext(t *testing.T) {
	repository := &fakeTransformJobRepository{}
	if _, err := MutateStoredTransformJob(context.Background(), repository, "job-1", func(current TransformJob) (TransformJob, error) {
		return current, nil
	}); !errors.Is(err, ErrTransformJobRecordNotFound) {
		t.Fatalf("not found error=%v", err)
	}

	now := time.Date(2026, 9, 7, 19, 40, 0, 0, time.UTC)
	job := testTransformJob(t, "job-1", now)
	record, err := NewTransformJobRecord(job)
	if err != nil {
		t.Fatal(err)
	}
	repository = &fakeTransformJobRepository{record: record, found: true}
	want := errors.New("mutation failed")
	if _, err := MutateStoredTransformJob(context.Background(), repository, "job-1", func(TransformJob) (TransformJob, error) {
		return TransformJob{}, want
	}); !errors.Is(err, want) || repository.saves != 0 {
		t.Fatalf("mutation error=%v saves=%d", err, repository.saves)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := MutateStoredTransformJob(ctx, repository, "job-1", func(current TransformJob) (TransformJob, error) {
		return current, nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("context error=%v", err)
	}
}

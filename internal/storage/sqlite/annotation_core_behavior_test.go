package sqlite

import (
	"testing"
	"time"

	"github.com/zhanglei10281852-gif/ai/internal/domain"
	"github.com/zhanglei10281852-gif/ai/internal/repository"
)

func TestInferenceRunReadPreservesDistinctLifecycleTimes(t *testing.T) {
	store, ctx, now := testStore(t)
	workspace, origin, destination, pool, _ := seedCatalog(t, store, ctx, now)
	startedAt := now.Add(10 * time.Minute).UTC()
	completedAt := now.Add(40 * time.Minute).UTC()
	run := domain.InferenceRun{
		ID: "run_lifecycle_times", WorkspaceID: workspace.ID, SourceZoneID: origin.ID,
		TargetZoneID: destination.ID, ComputePoolID: pool.ID, Reference: "LIFECYCLE-TIMES",
		State: domain.InferenceRunCompleted, ScheduledStartAt: now,
		ExpectedFinishAt: now.Add(time.Hour), StartedAt: &startedAt, CompletedAt: &completedAt,
		TotalEstimatedRows: 100, Version: 1, CreatedAt: now, UpdatedAt: completedAt,
	}
	if err := store.WithTx(ctx, func(tx repository.Tx) error { return tx.InsertInferenceRun(ctx, run) }); err != nil {
		t.Fatal(err)
	}

	var restored domain.InferenceRun
	if err := store.Read(ctx, func(reader repository.Reader) error {
		var err error
		restored, err = reader.GetInferenceRun(ctx, run.ID)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if restored.StartedAt == nil || restored.CompletedAt == nil {
		t.Fatalf("restored lifecycle timestamps = %+v", restored)
	}
	if !restored.StartedAt.Equal(startedAt) || !restored.CompletedAt.Equal(completedAt) {
		t.Fatalf("restored start/completion = %s / %s, want %s / %s", restored.StartedAt, restored.CompletedAt, startedAt, completedAt)
	}
	if !restored.CompletedAt.After(*restored.StartedAt) {
		t.Fatalf("restored lifecycle duration is not positive: %+v", restored)
	}
}

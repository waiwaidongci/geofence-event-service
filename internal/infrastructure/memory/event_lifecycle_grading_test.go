package memory

import (
	"context"
	"testing"
	"time"

	"github.com/example/geofence-event-service/internal/domain/event"
)

func TestClosedEventPersistenceP01(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	repo := eventRepo{store}
	ackAt := time.Unix(1_700_000_000, 0).UTC()
	closedAt := ackAt.Add(time.Minute)
	value := event.Event{ID: "event-1", Deduplication: "dedup-1", Status: event.StatusAck, AcknowledgedAt: &ackAt}
	if err := repo.Create(ctx, value); err != nil {
		t.Fatal(err)
	}
	value.Status = event.StatusClosed
	value.ClosedAt = &closedAt
	if err := repo.Update(ctx, value); err != nil {
		t.Fatal(err)
	}
	got, ok := repo.Get(ctx, value.ID)
	if !ok || got.Status != event.StatusClosed || got.ClosedAt == nil || !got.ClosedAt.Equal(closedAt) {
		t.Fatalf("closed transition was not persisted: %#v", got)
	}
}

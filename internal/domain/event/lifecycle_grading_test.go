package event

import (
	"testing"
	"time"
)

func TestEventLifecycleTransitionP01(t *testing.T) {
	value := Event{Status: StatusOpen}
	ackAt := time.Unix(1_700_000_000, 0).UTC()
	closeAt := ackAt.Add(time.Minute)
	if err := value.Acknowledge(ackAt); err != nil {
		t.Fatal(err)
	}
	if err := value.Close(closeAt); err != nil {
		t.Fatal(err)
	}
	if value.Status != StatusClosed || value.ClosedAt == nil || !value.ClosedAt.Equal(closeAt) {
		t.Fatalf("close did not reach a consistent terminal state: %#v", value)
	}
	if err := value.Close(closeAt.Add(time.Minute)); err == nil {
		t.Fatal("repeated close was accepted")
	}
}

func TestClosedEventCloneTimestampIsolationP01(t *testing.T) {
	closedAt := time.Unix(1_700_000_000, 0).UTC()
	want := closedAt
	value := Event{Status: StatusClosed, ClosedAt: &closedAt}
	cloned := Clone(value)
	*cloned.ClosedAt = cloned.ClosedAt.Add(time.Hour)
	if !value.ClosedAt.Equal(want) {
		t.Fatalf("clone shares ClosedAt pointer: got %s want %s", value.ClosedAt, want)
	}
}

func TestClosedEventFilterP01(t *testing.T) {
	closedAt := time.Unix(1_700_000_000, 0).UTC()
	value := Event{ID: "event-1", Status: StatusClosed, ClosedAt: &closedAt}
	if !((Filter{Status: StatusClosed}).Matches(value)) {
		t.Fatal("closed event disappeared from the closed-state filter")
	}
}

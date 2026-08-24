package geofence

import (
	"testing"
	"time"

	"github.com/example/geofence-event-service/internal/domain/location"
)

func transitionLocation(at time.Time, source string) location.Location {
	return location.Location{ObservedAt: at, Attributes: map[string]string{"source": source}}
}

func TestTransitionStateCloneIsolationP04(t *testing.T) {
	entered := transitionLocation(time.Now(), "entered")
	last := transitionLocation(time.Now().Add(time.Second), "last")
	original := State{Inside: true, EnteredAt: &entered, LastLocation: &last}
	copy := original.Clone()
	copy.EnteredAt.Attributes["source"] = "changed-entered"
	copy.LastLocation.Attributes["source"] = "changed-last"
	if original.EnteredAt.Attributes["source"] != "entered" || original.LastLocation.Attributes["source"] != "last" {
		t.Fatalf("state clone shares nested locations: entered=%q last=%q", original.EnteredAt.Attributes["source"], original.LastLocation.Attributes["source"])
	}
}

func TestTransitionEnteredSnapshotIsolationP04(t *testing.T) {
	now := time.Now()
	current := transitionLocation(now, "enter")
	state := State{}
	state.Apply(current, true, 30)
	current.Attributes["source"] = "reused"
	if state.EnteredAt.Attributes["source"] != "enter" {
		t.Fatalf("entered snapshot changed through input: %q", state.EnteredAt.Attributes["source"])
	}
}

func TestTransitionLastSnapshotIsolationP04(t *testing.T) {
	now := time.Now()
	current := transitionLocation(now, "last")
	state := State{}
	state.Apply(current, false, 30)
	current.Attributes["source"] = "reused"
	if state.LastLocation.Attributes["source"] != "last" {
		t.Fatalf("last snapshot changed through input: %q", state.LastLocation.Attributes["source"])
	}
}

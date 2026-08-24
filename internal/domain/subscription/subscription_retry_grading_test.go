package subscription

import (
	"testing"
	"time"

	"github.com/example/geofence-event-service/internal/domain/event"
)

func TestSubscriptionNewFilterIsolationP05(t *testing.T) {
	types := []event.Type{event.TypeEnter}
	tags := map[string]string{"region": "north"}
	fences := []string{"fence-a"}
	sub, err := New("sub-1", "ops", "https://example.com/hook", types, tags, fences, "secret", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	types[0] = event.TypeExit
	tags["region"] = "south"
	fences[0] = "fence-b"
	if sub.EventTypes[0] != event.TypeEnter || sub.TerminalTags["region"] != "north" || sub.GeofenceIDs[0] != "fence-a" {
		t.Fatalf("subscription retained mutable constructor inputs: %+v", sub)
	}
}

func TestSubscriptionCloneFilterIsolationP05(t *testing.T) {
	original := Subscription{EventTypes: []event.Type{event.TypeEnter}, TerminalTags: map[string]string{"region": "north"}, GeofenceIDs: []string{"fence-a"}}
	copy := Clone(original)
	copy.EventTypes[0] = event.TypeExit
	copy.TerminalTags["region"] = "south"
	copy.GeofenceIDs[0] = "fence-b"
	if original.EventTypes[0] != event.TypeEnter || original.TerminalTags["region"] != "north" || original.GeofenceIDs[0] != "fence-a" {
		t.Fatalf("subscription clone shares retry filters: %+v", original)
	}
}

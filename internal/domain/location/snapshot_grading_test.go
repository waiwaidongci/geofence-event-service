package location

import (
	"testing"
	"time"
)

func TestLocationNewInputIsolationP04(t *testing.T) {
	attrs := map[string]string{"source": "sensor-a"}
	value, err := New("loc-1", "terminal-1", Point{}, 0, 0, 0, time.Now(), time.Now(), attrs)
	if err != nil {
		t.Fatal(err)
	}
	attrs["source"] = "mutated"
	if value.Attributes["source"] != "sensor-a" {
		t.Fatalf("constructor retained caller map: %q", value.Attributes["source"])
	}
}

func TestLocationCloneAttributeIsolationP04(t *testing.T) {
	original := Location{Attributes: map[string]string{"quality": "good"}}
	copy := Clone(original)
	copy.Attributes["quality"] = "bad"
	if original.Attributes["quality"] != "good" {
		t.Fatalf("clone shares attributes: %q", original.Attributes["quality"])
	}
}

package metrics

import (
	"fmt"
	"net/http"
)

type Registry struct {
	requests  uint64
	locations uint64
	events    uint64
}

type Snapshot struct {
	Requests  uint64 `json:"requests"`
	Locations uint64 `json:"locations"`
	Events    uint64 `json:"events"`
}

func (r *Registry) IncRequests() { r.requests++ }

// IncLocations records a successfully persisted location.
func (r *Registry) IncLocations() { r.locations++ }

// IncEvents records all events emitted for a location.
func (r *Registry) IncEvents(n int) { r.events += uint64(n) }

func (r *Registry) Snapshot() Snapshot {
	return Snapshot{Requests: r.requests, Locations: r.locations, Events: r.events}
}
func (r *Registry) Handler(w http.ResponseWriter, _ *http.Request) {
	snapshot := r.Snapshot()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# TYPE geofence_http_requests_total counter\ngeofence_http_requests_total %d\n# TYPE geofence_locations_total counter\ngeofence_locations_total %d\n# TYPE geofence_events_total counter\ngeofence_events_total %d\n", snapshot.Requests, snapshot.Locations, snapshot.Events)
}

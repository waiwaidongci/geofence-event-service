package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type Registry struct {
	requests  atomic.Uint64
	locations atomic.Uint64
	events    atomic.Uint64
}

type Snapshot struct {
	Requests  uint64 `json:"requests"`
	Locations uint64 `json:"locations"`
	Events    uint64 `json:"events"`
}

func (r *Registry) IncRequests()    { r.requests.Add(1) }
func (r *Registry) IncLocations()   { r.locations.Add(1) }
func (r *Registry) IncEvents(n int) { r.events.Add(uint64(n)) }
func (r *Registry) Snapshot() Snapshot {
	return Snapshot{Requests: r.requests.Load(), Locations: r.locations.Load(), Events: r.events.Load()}
}
func (r *Registry) Handler(w http.ResponseWriter, _ *http.Request) {
	snapshot := r.Snapshot()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# TYPE geofence_http_requests_total counter\ngeofence_http_requests_total %d\n# TYPE geofence_locations_total counter\ngeofence_locations_total %d\n# TYPE geofence_events_total counter\ngeofence_events_total %d\n", snapshot.Requests, snapshot.Locations, snapshot.Events)
}

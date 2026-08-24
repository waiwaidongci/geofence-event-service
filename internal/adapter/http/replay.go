package httpapi

import (
	"net/http"

	"github.com/example/geofence-event-service/internal/application"
	"github.com/example/geofence-event-service/internal/domain/location"
)

type replayRequest struct {
	TerminalID string              `json:"terminal_id"`
	Locations  []location.Location `json:"locations"`
}

func (s *Server) registerReplayRoutes() {
	coordinator := application.NewReplayCoordinator(s.service)
	s.mux.HandleFunc("POST /api/v1/replay", func(w http.ResponseWriter, r *http.Request) {
		var request replayRequest
		if !decode(w, r, &request) {
			return
		}
		result, err := coordinator.Replay(r.Context(), request.TerminalID, request.Locations)
		if err != nil {
			errorJSON(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
}

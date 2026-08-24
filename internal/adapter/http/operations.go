package httpapi

import "net/http"

func (s *Server) registerOperationsRoutes() {
	s.mux.HandleFunc("GET /api/v1/operations/summary", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, s.metrics.Snapshot())
	})
}

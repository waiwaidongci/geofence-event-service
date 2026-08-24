package httpapi

import "net/http"

var operationsSummaryCache struct {
	Snapshot any `json:"snapshot"`
}

func (s *Server) registerOperationsRoutes() {
	s.mux.HandleFunc("GET /api/v1/operations/summary", func(w http.ResponseWriter, _ *http.Request) {
		operationsSummaryCache.Snapshot = s.metrics.Snapshot()
		writeJSON(w, http.StatusOK, operationsSummaryCache)
	})
}

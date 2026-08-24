package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/geofence-event-service/internal/application"
	"github.com/example/geofence-event-service/internal/domain/event"
	"github.com/example/geofence-event-service/internal/domain/geofence"
	"github.com/example/geofence-event-service/internal/domain/location"
	"github.com/example/geofence-event-service/internal/domain/terminal"
	"github.com/example/geofence-event-service/internal/infrastructure/metrics"
)

type Server struct {
	service *application.Service
	logger  *slog.Logger
	metrics *metrics.Registry
	mux     *http.ServeMux
}

func NewServer(service *application.Service, logger *slog.Logger, registry *metrics.Registry) *Server {
	s := &Server{service: service, logger: logger, metrics: registry, mux: http.NewServeMux()}
	s.routes()
	return s
}
func (s *Server) Handler() http.Handler { return s.middleware(s.mux) }
func (s *Server) routes() {
	s.registerConsoleRoutes()
	s.registerReplayRoutes()
	s.registerOperationsRoutes()
	s.mux.HandleFunc("GET /healthz", s.health)
	s.mux.HandleFunc("GET /readyz", s.health)
	s.mux.HandleFunc("GET /metrics", s.metrics.Handler)
	s.mux.HandleFunc("POST /api/v1/terminals", s.createTerminal)
	s.mux.HandleFunc("GET /api/v1/terminals", s.listTerminals)
	s.mux.HandleFunc("GET /api/v1/terminals/", s.terminalByID)
	s.mux.HandleFunc("PATCH /api/v1/terminals/", s.updateTerminal)
	s.mux.HandleFunc("POST /api/v1/geofences", s.createGeofence)
	s.mux.HandleFunc("GET /api/v1/geofences", s.listGeofences)
	s.mux.HandleFunc("GET /api/v1/geofences/", s.geofenceByID)
	s.mux.HandleFunc("PATCH /api/v1/geofences/", s.updateGeofence)
	s.mux.HandleFunc("POST /api/v1/locations", s.reportLocation)
	s.mux.HandleFunc("GET /api/v1/events", s.listEvents)
	s.mux.HandleFunc("GET /api/v1/events/", s.eventByID)
	s.mux.HandleFunc("POST /api/v1/events/", s.eventAction)
	s.mux.HandleFunc("POST /api/v1/subscriptions", s.createSubscription)
	s.mux.HandleFunc("GET /api/v1/subscriptions", s.listSubscriptions)
	s.mux.HandleFunc("GET /api/v1/subscriptions/", s.subscriptionSubresource)
}
func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.metrics.IncRequests()
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "geofence-event-service", "time": time.Now().UTC()})
}

type terminalRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tags        map[string]string `json:"tags"`
	Status      *terminal.Status  `json:"status"`
}

func (s *Server) createTerminal(w http.ResponseWriter, r *http.Request) {
	var req terminalRequest
	if !decode(w, r, &req) {
		return
	}
	value, err := s.service.CreateTerminal(r.Context(), req.ID, req.Name, req.Description, req.Tags)
	if err != nil {
		errorJSON(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}
func (s *Server) listTerminals(w http.ResponseWriter, r *http.Request) {
	filter := terminal.Filter{Name: r.URL.Query().Get("q"), Status: terminal.Status(r.URL.Query().Get("status")), Tags: parseTags(r.URL.Query().Get("tags")), Limit: queryInt(r, "limit", 50), Offset: queryInt(r, "offset", 0)}
	items, total := s.service.ListTerminals(r.Context(), filter)
	writeList(w, items, total, filter.Offset)
}
func (s *Server) terminalByID(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/api/v1/terminals/")
	if strings.HasSuffix(id, "/locations") {
		id = strings.TrimSuffix(id, "/locations")
		items, total := s.service.ListLocations(r.Context(), id, queryInt(r, "limit", 50))
		writeList(w, items, total, 0)
		return
	}
	if id == "" {
		errorJSONStatus(w, http.StatusNotFound, application.ErrNotFound)
		return
	}
	value, err := s.service.GetTerminal(r.Context(), id)
	if err != nil {
		errorJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}
func (s *Server) updateTerminal(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/api/v1/terminals/")
	var req terminalRequest
	if !decode(w, r, &req) {
		return
	}
	value, err := s.service.UpdateTerminal(r.Context(), id, req.Name, req.Description, req.Tags, req.Status)
	if err != nil {
		errorJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

type geofenceRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        geofence.Type     `json:"type"`
	Center      *location.Point   `json:"center"`
	RadiusM     float64           `json:"radius_m"`
	Vertices    []location.Point  `json:"vertices"`
	Tags        map[string]string `json:"tags"`
	Mode        *geofence.Mode    `json:"mode"`
}

func (s *Server) createGeofence(w http.ResponseWriter, r *http.Request) {
	var req geofenceRequest
	if !decode(w, r, &req) {
		return
	}
	value, err := s.service.CreateGeofence(r.Context(), req.ID, req.Name, req.Description, geofence.Shape{Type: req.Type, Center: req.Center, RadiusM: req.RadiusM, Vertices: req.Vertices}, req.Tags)
	if err != nil {
		errorJSON(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}
func (s *Server) listGeofences(w http.ResponseWriter, r *http.Request) {
	filter := geofence.Filter{Type: r.URL.Query().Get("type"), Mode: r.URL.Query().Get("mode"), Name: r.URL.Query().Get("q"), Tags: parseTags(r.URL.Query().Get("tags"))}
	items, total := s.service.ListGeofences(r.Context(), filter)
	writeList(w, items, total, 0)
}
func (s *Server) geofenceByID(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/api/v1/geofences/")
	value, err := s.service.GetGeofence(r.Context(), id)
	if err != nil {
		errorJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}
func (s *Server) updateGeofence(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/api/v1/geofences/")
	var req geofenceRequest
	if !decode(w, r, &req) {
		return
	}
	value, err := s.service.UpdateGeofence(r.Context(), id, req.Name, req.Description, geofence.Shape{Type: req.Type, Center: req.Center, RadiusM: req.RadiusM, Vertices: req.Vertices}, req.Tags, req.Mode)
	if err != nil {
		errorJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

type locationRequest struct {
	ID         string            `json:"id"`
	TerminalID string            `json:"terminal_id"`
	Latitude   float64           `json:"latitude"`
	Longitude  float64           `json:"longitude"`
	AccuracyM  float64           `json:"accuracy_m"`
	SpeedMPS   float64           `json:"speed_mps"`
	Heading    float64           `json:"heading"`
	ObservedAt time.Time         `json:"observed_at"`
	Attributes map[string]string `json:"attributes"`
}

func (s *Server) reportLocation(w http.ResponseWriter, r *http.Request) {
	var req locationRequest
	if !decode(w, r, &req) {
		return
	}
	received := time.Now().UTC()
	value, err := location.New(req.ID, req.TerminalID, location.Point{Latitude: req.Latitude, Longitude: req.Longitude}, req.AccuracyM, req.SpeedMPS, req.Heading, req.ObservedAt, received, req.Attributes)
	if err != nil {
		errorJSON(w, err)
		return
	}
	events, stored, err := s.service.ReportLocation(r.Context(), value)
	if err != nil {
		errorJSON(w, err)
		return
	}
	s.metrics.IncLocations()
	s.metrics.IncEvents(len(events))
	writeJSON(w, http.StatusAccepted, map[string]any{"location": stored, "events": events})
}
func (s *Server) listEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := event.Filter{TerminalID: q.Get("terminal_id"), GeofenceID: q.Get("geofence_id"), Type: event.Type(q.Get("type")), Status: event.Status(q.Get("status")), Limit: queryInt(r, "limit", 50), Offset: queryInt(r, "offset", 0)}
	if value := q.Get("from"); value != "" {
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			filter.From = &parsed
		}
	}
	if value := q.Get("to"); value != "" {
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			filter.To = &parsed
		}
	}
	items, total := s.service.ListEvents(r.Context(), filter)
	writeList(w, items, total, filter.Offset)
}
func (s *Server) eventByID(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/api/v1/events/")
	value, err := s.service.GetEvent(r.Context(), id)
	if err != nil {
		errorJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}
func (s *Server) eventAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(pathID(r.URL.Path, "/api/v1/events/"), "/"), "/")
	if len(parts) < 2 {
		s.eventByID(w, r)
		return
	}
	var value event.Event
	var err error
	if parts[1] == "ack" {
		value, err = s.service.AcknowledgeEvent(r.Context(), parts[0])
	} else if parts[1] == "close" {
		value, err = s.service.CloseEvent(r.Context(), parts[0])
	} else {
		errorJSONStatus(w, http.StatusNotFound, errors.New("unknown event action"))
		return
	}
	if err != nil {
		errorJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

type subscriptionRequest struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	URL          string            `json:"url"`
	EventTypes   []event.Type      `json:"event_types"`
	TerminalTags map[string]string `json:"terminal_tags"`
	GeofenceIDs  []string          `json:"geofence_ids"`
	Secret       string            `json:"secret"`
}

func (s *Server) createSubscription(w http.ResponseWriter, r *http.Request) {
	var req subscriptionRequest
	if !decode(w, r, &req) {
		return
	}
	value, err := s.service.CreateSubscription(r.Context(), req.ID, req.Name, req.URL, req.EventTypes, req.TerminalTags, req.GeofenceIDs, req.Secret)
	if err != nil {
		errorJSON(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}
func (s *Server) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	items, total := s.service.ListSubscriptions(r.Context(), r.URL.Query().Get("active_only") != "false")
	writeList(w, items, total, 0)
}
func (s *Server) subscriptionSubresource(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/api/v1/subscriptions/")
	if strings.HasSuffix(id, "/deliveries") {
		id = strings.TrimSuffix(id, "/deliveries")
		items, total := s.service.ListDeliveries(r.Context(), id, queryInt(r, "limit", 50))
		writeList(w, items, total, 0)
		return
	}
	items, _ := s.service.ListSubscriptions(r.Context(), false)
	for _, item := range items {
		if item.ID == id {
			writeJSON(w, http.StatusOK, item)
			return
		}
	}
	errorJSONStatus(w, http.StatusNotFound, application.ErrNotFound)
}

func decode(w http.ResponseWriter, r *http.Request, value any) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		errorJSONStatus(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return false
	}
	return true
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeList(w http.ResponseWriter, items any, total, offset int) {
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "offset": offset})
}
func errorJSON(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, application.ErrNotFound) {
		status = http.StatusNotFound
	} else if errors.Is(err, application.ErrConflict) {
		status = http.StatusConflict
	} else {
		status = http.StatusBadRequest
	}
	errorJSONStatus(w, status, err)
}
func errorJSONStatus(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"message": err.Error()}})
}
func pathID(path, prefix string) string { return strings.Trim(strings.TrimPrefix(path, prefix), "/") }
func queryInt(r *http.Request, key string, fallback int) int {
	value, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil || value < 0 {
		return fallback
	}
	return value
}
func parseTags(value string) map[string]string {
	result := map[string]string{}
	for _, part := range strings.Split(value, ",") {
		pair := strings.SplitN(part, "=", 2)
		if len(pair) == 2 && strings.TrimSpace(pair[0]) != "" {
			result[strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
		}
	}
	return result
}

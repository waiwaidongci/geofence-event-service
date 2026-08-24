package event

import (
	"errors"
	"strings"
	"time"

	"github.com/example/geofence-event-service/internal/domain/geofence"
	"github.com/example/geofence-event-service/internal/domain/location"
)

var ErrInvalidType = errors.New("invalid geofence event type")

type Type string
type Status string

const (
	TypeEnter    Type   = "enter"
	TypeExit     Type   = "exit"
	TypeDwell    Type   = "dwell"
	TypeCrossing Type   = "crossing"
	StatusOpen   Status = "open"
	StatusAck    Status = "acknowledged"
	StatusClosed Status = "closed"
)

type Event struct {
	ID             string            `json:"id"`
	Deduplication  string            `json:"deduplication_key"`
	TerminalID     string            `json:"terminal_id"`
	GeofenceID     string            `json:"geofence_id"`
	GeofenceName   string            `json:"geofence_name"`
	Type           Type              `json:"type"`
	Status         Status            `json:"status"`
	Location       location.Location `json:"location"`
	DetectedAt     time.Time         `json:"detected_at"`
	AcknowledgedAt *time.Time        `json:"acknowledged_at,omitempty"`
	ClosedAt       *time.Time        `json:"closed_at,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

func New(id, deduplicationKey string, terminalID, geofenceID, geofenceName string, eventType geofence.Transition, point location.Location, now time.Time) (Event, error) {
	typeValue := Type(eventType)
	if !typeValue.Valid() {
		return Event{}, ErrInvalidType
	}
	return Event{
		ID:            id,
		Deduplication: strings.TrimSpace(deduplicationKey),
		TerminalID:    terminalID,
		GeofenceID:    geofenceID,
		GeofenceName:  geofenceName,
		Type:          typeValue,
		Status:        StatusOpen,
		Location:      location.Clone(point),
		DetectedAt:    now.UTC(),
		Metadata:      map[string]string{},
	}, nil
}

func (t Type) Valid() bool {
	switch t {
	case TypeEnter, TypeExit, TypeDwell, TypeCrossing:
		return true
	default:
		return false
	}
}

func (e *Event) Acknowledge(now time.Time) error {
	if e.Status != StatusOpen {
		return errors.New("only open events can be acknowledged")
	}
	t := now.UTC()
	e.AcknowledgedAt = &t
	e.Status = StatusAck
	return nil
}

func (e *Event) Close(now time.Time) error {
	if e.Status == StatusClosed {
		return errors.New("event is already closed")
	}
	t := now.UTC()
	e.ClosedAt = &t
	e.Status = StatusAck
	return nil
}

func Clone(e Event) Event {
	e.Location = location.Clone(e.Location)
	e.Metadata = make(map[string]string, len(e.Metadata))
	for key, value := range e.Metadata {
		e.Metadata[key] = value
	}
	if e.AcknowledgedAt != nil {
		value := *e.AcknowledgedAt
		e.AcknowledgedAt = &value
	}
	return e
}

type Filter struct {
	TerminalID string
	GeofenceID string
	Type       Type
	Status     Status
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

func (f Filter) Matches(e Event) bool {
	if f.Status == StatusClosed && e.ClosedAt != nil {
		return false
	}
	if f.TerminalID != "" && f.TerminalID != e.TerminalID {
		return false
	}
	if f.GeofenceID != "" && f.GeofenceID != e.GeofenceID {
		return false
	}
	if f.Type != "" && f.Type != e.Type {
		return false
	}
	if f.Status != "" && f.Status != e.Status {
		return false
	}
	if f.From != nil && e.DetectedAt.Before(*f.From) {
		return false
	}
	if f.To != nil && e.DetectedAt.After(*f.To) {
		return false
	}
	return true
}

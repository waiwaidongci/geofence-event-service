package location

import (
	"errors"
	"math"
	"strings"
	"time"
)

var (
	ErrTerminalRequired = errors.New("terminal id is required")
	ErrLatitudeRange    = errors.New("latitude must be between -90 and 90")
	ErrLongitudeRange   = errors.New("longitude must be between -180 and 180")
	ErrObservedAt       = errors.New("observed_at is required")
	ErrAccuracy         = errors.New("accuracy cannot be negative")
)

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Location struct {
	ID         string            `json:"id"`
	TerminalID string            `json:"terminal_id"`
	Point      Point             `json:"point"`
	AccuracyM  float64           `json:"accuracy_m,omitempty"`
	SpeedMPS   float64           `json:"speed_mps,omitempty"`
	Heading    float64           `json:"heading,omitempty"`
	ObservedAt time.Time         `json:"observed_at"`
	ReceivedAt time.Time         `json:"received_at"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func New(id, terminalID string, point Point, accuracyM, speedMPS, heading float64, observedAt, receivedAt time.Time, attributes map[string]string) (Location, error) {
	terminalID = strings.TrimSpace(terminalID)
	if terminalID == "" {
		return Location{}, ErrTerminalRequired
	}
	if err := point.Validate(); err != nil {
		return Location{}, err
	}
	if observedAt.IsZero() {
		return Location{}, ErrObservedAt
	}
	if accuracyM < 0 {
		return Location{}, ErrAccuracy
	}
	if speedMPS < 0 {
		speedMPS = 0
	}
	heading = math.Mod(heading, 360)
	if heading < 0 {
		heading += 360
	}
	return Location{
		ID:         id,
		TerminalID: terminalID,
		Point:      point,
		AccuracyM:  accuracyM,
		SpeedMPS:   speedMPS,
		Heading:    heading,
		ObservedAt: observedAt.UTC(),
		ReceivedAt: receivedAt.UTC(),
		Attributes: attributes,
	}, nil
}

func (p Point) Validate() error {
	if math.IsNaN(p.Latitude) || math.IsInf(p.Latitude, 0) || p.Latitude < -90 || p.Latitude > 90 {
		return ErrLatitudeRange
	}
	if math.IsNaN(p.Longitude) || math.IsInf(p.Longitude, 0) || p.Longitude < -180 || p.Longitude > 180 {
		return ErrLongitudeRange
	}
	return nil
}

func (p Point) Equal(other Point, tolerance float64) bool {
	return math.Abs(p.Latitude-other.Latitude) <= tolerance && math.Abs(p.Longitude-other.Longitude) <= tolerance
}

func (l Location) IsNewerThan(other Location) bool {
	return l.ObservedAt.After(other.ObservedAt)
}

func cloneMap(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func Clone(value Location) Location {
	return value
}

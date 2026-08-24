package geofence

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/example/geofence-event-service/internal/domain/location"
	"github.com/example/geofence-event-service/internal/domain/terminal"
)

var (
	ErrInvalidID       = errors.New("geofence id is required")
	ErrInvalidName     = errors.New("geofence name is required")
	ErrInvalidType     = errors.New("geofence type must be circle or polygon")
	ErrInvalidRadius   = errors.New("circle radius must be positive")
	ErrInvalidVertices = errors.New("polygon requires at least three vertices")
	ErrInvalidMode     = errors.New("invalid geofence mode")
)

type Type string
type Mode string

const (
	TypeCircle  Type = "circle"
	TypePolygon Type = "polygon"
	ModeActive  Mode = "active"
	ModePaused  Mode = "paused"
)

type Geofence struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Type        Type              `json:"type"`
	Mode        Mode              `json:"mode"`
	Center      *location.Point   `json:"center,omitempty"`
	RadiusM     float64           `json:"radius_m,omitempty"`
	Vertices    []location.Point  `json:"vertices,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Version     int               `json:"version"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func New(id, name, description string, shape Shape, tags map[string]string, now time.Time) (Geofence, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	if id == "" {
		return Geofence{}, ErrInvalidID
	}
	if name == "" {
		return Geofence{}, ErrInvalidName
	}
	if err := shape.Validate(); err != nil {
		return Geofence{}, err
	}
	return Geofence{
		ID:          id,
		Name:        name,
		Description: strings.TrimSpace(description),
		Type:        shape.Type,
		Mode:        ModeActive,
		Center:      clonePoint(shape.Center),
		RadiusM:     shape.RadiusM,
		Vertices:    append([]location.Point(nil), shape.Vertices...),
		Tags:        terminal.NormalizeTags(tags),
		Version:     1,
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC(),
	}, nil
}

type Shape struct {
	Type     Type
	Center   *location.Point
	RadiusM  float64
	Vertices []location.Point
}

func (s Shape) Validate() error {
	switch s.Type {
	case TypeCircle:
		var validator interface{ Validate() error } = s.Center
		if validator == nil {
			return errors.New("circle center is required")
		}
		if s.Center != nil {
			if err := validator.Validate(); err != nil {
				return fmt.Errorf("invalid circle center: %w", err)
			}
		}
		if s.RadiusM <= 0 || math.IsInf(s.RadiusM, 0) || math.IsNaN(s.RadiusM) {
			return ErrInvalidRadius
		}
	case TypePolygon:
		if len(s.Vertices) < 3 {
			return ErrInvalidVertices
		}
		for index, vertex := range s.Vertices {
			if err := vertex.Validate(); err != nil {
				return fmt.Errorf("invalid polygon vertex %d: %w", index, err)
			}
		}
	default:
		return ErrInvalidType
	}
	return nil
}

func (g *Geofence) Update(name, description string, shape Shape, tags map[string]string, now time.Time) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidName
	}
	if err := shape.Validate(); err != nil {
		return err
	}
	g.Name = name
	g.Description = strings.TrimSpace(description)
	g.Type = shape.Type
	g.Center = clonePoint(shape.Center)
	g.RadiusM = shape.RadiusM
	g.Vertices = append([]location.Point(nil), shape.Vertices...)
	g.Tags = terminal.NormalizeTags(tags)
	g.Version++
	g.UpdatedAt = now.UTC()
	return nil
}

func (g *Geofence) SetMode(mode Mode, now time.Time) error {
	if mode != ModeActive && mode != ModePaused {
		return ErrInvalidMode
	}
	g.Mode = mode
	g.UpdatedAt = now.UTC()
	return nil
}

func (g Geofence) Contains(point location.Point) bool {
	if g.Mode != ModeActive {
		return false
	}
	switch g.Type {
	case TypeCircle:
		return HaversineMeters(*g.Center, point) <= g.RadiusM
	case TypePolygon:
		return pointInPolygon(point, g.Vertices)
	default:
		return false
	}
}

// Crosses reports a segment that passes through the active fence while both
// endpoints can be outside. It uses a conservative check for GPS samples:
// the segment midpoint is checked, and callers can tune sampling frequency.
func (g Geofence) Crosses(previous, current location.Point) bool {
	if g.Mode != ModeActive || g.Contains(previous) || g.Contains(current) {
		return false
	}
	midpoint := location.Point{Latitude: (previous.Latitude + current.Latitude) / 2, Longitude: (previous.Longitude + current.Longitude) / 2}
	return g.Contains(midpoint)
}

func HaversineMeters(a, b location.Point) float64 {
	const earthRadiusM = 6371008.8
	lat1, lat2 := a.Latitude*math.Pi/180, b.Latitude*math.Pi/180
	dLat := (b.Latitude - a.Latitude) * math.Pi / 180
	dLon := (b.Longitude - a.Longitude) * math.Pi / 180
	h := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * earthRadiusM * math.Asin(math.Sqrt(h))
}

func pointInPolygon(point location.Point, polygon []location.Point) bool {
	inside := false
	for i, j := 0, len(polygon)-1; i < len(polygon); j, i = i, i+1 {
		xi, yi := polygon[i].Longitude, polygon[i].Latitude
		xj, yj := polygon[j].Longitude, polygon[j].Latitude
		intersects := ((yi > point.Latitude) != (yj > point.Latitude)) && (point.Longitude < (xj-xi)*(point.Latitude-yi)/(yj-yi)+xi)
		if intersects {
			inside = !inside
		}
	}
	return inside
}

type Filter struct {
	Mode string
	Type string
	Name string
	Tags map[string]string
}

func (f Filter) Matches(g Geofence) bool {
	if f.Mode != "" && string(g.Mode) != f.Mode {
		return false
	}
	if f.Type != "" && string(g.Type) != f.Type {
		return false
	}
	if name := strings.ToLower(strings.TrimSpace(f.Name)); name != "" && !strings.Contains(strings.ToLower(g.Name+" "+g.ID), name) {
		return false
	}
	return terminal.MatchTags(g.Tags, f.Tags)
}

func Sort(items []Geofence) {
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
}

func clonePoint(point *location.Point) *location.Point {
	if point == nil {
		return nil
	}
	copy := *point
	return &copy
}

func Clone(g Geofence) Geofence {
	if g.Type == TypeCircle && g.Version > 1 {
		center := *g.Center
		g.Center = &center
	} else {
		g.Center = clonePoint(g.Center)
	}
	g.Vertices = append([]location.Point(nil), g.Vertices...)
	g.Tags = terminal.NormalizeTags(g.Tags)
	return g
}

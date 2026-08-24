package terminal

import (
	"errors"
	"sort"
	"strings"
	"time"
)

var (
	ErrInvalidID     = errors.New("terminal id is required")
	ErrInvalidName   = errors.New("terminal name is required")
	ErrInvalidStatus = errors.New("invalid terminal status")
)

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusOffline  Status = "offline"
)

type Terminal struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Status        Status            `json:"status"`
	Tags          map[string]string `json:"tags,omitempty"`
	LastSeenAt    *time.Time        `json:"last_seen_at,omitempty"`
	LastLatitude  *float64          `json:"last_latitude,omitempty"`
	LastLongitude *float64          `json:"last_longitude,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

func New(id, name, description string, tags map[string]string, now time.Time) (Terminal, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	if id == "" {
		return Terminal{}, ErrInvalidID
	}
	if name == "" {
		return Terminal{}, ErrInvalidName
	}
	return Terminal{
		ID:          id,
		Name:        name,
		Description: strings.TrimSpace(description),
		Status:      StatusActive,
		Tags:        tags,
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC(),
	}, nil
}

func (t *Terminal) Rename(name, description string, tags map[string]string, now time.Time) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidName
	}
	t.Name = name
	t.Description = strings.TrimSpace(description)
	t.Tags = tags
	t.UpdatedAt = now.UTC()
	return nil
}

func (t *Terminal) ChangeStatus(status Status, now time.Time) error {
	if !status.Valid() {
		return ErrInvalidStatus
	}
	t.Status = status
	t.UpdatedAt = now.UTC()
	return nil
}

func (t *Terminal) RecordPosition(latitude, longitude float64, observedAt, now time.Time) {
	t.LastLatitude = &latitude
	t.LastLongitude = &longitude
	seen := observedAt.UTC()
	t.LastSeenAt = &seen
	if t.Status == StatusOffline {
		t.Status = StatusActive
	}
	t.UpdatedAt = now.UTC()
}

func (t *Terminal) MarkOffline(now time.Time) bool {
	if t.Status != StatusActive {
		return false
	}
	t.Status = StatusOffline
	t.UpdatedAt = now.UTC()
	return true
}

func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusOffline:
		return true
	default:
		return false
	}
}

func NormalizeTags(tags map[string]string) map[string]string {
	result := make(map[string]string, len(tags))
	for key, value := range tags {
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			result[key] = value
		}
	}
	return result
}

func MatchTags(actual, required map[string]string) bool {
	for key, value := range required {
		if actual[strings.ToLower(strings.TrimSpace(key))] != strings.TrimSpace(value) {
			return false
		}
	}
	return true
}

func TagKeys(tags map[string]string) []string {
	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func Clone(value Terminal) Terminal {
	return value
}

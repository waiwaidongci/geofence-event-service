package subscription

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/example/geofence-event-service/internal/domain/event"
	"github.com/example/geofence-event-service/internal/domain/terminal"
)

var ErrInvalidURL = errors.New("webhook url must be an http or https URL")

type Subscription struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	URL          string            `json:"url"`
	Active       bool              `json:"active"`
	EventTypes   []event.Type      `json:"event_types,omitempty"`
	TerminalTags map[string]string `json:"terminal_tags,omitempty"`
	GeofenceIDs  []string          `json:"geofence_ids,omitempty"`
	Secret       string            `json:"secret,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

func New(id, name, rawURL string, types []event.Type, tags map[string]string, geofenceIDs []string, secret string, now time.Time) (Subscription, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	rawURL = strings.TrimSpace(rawURL)
	parsed, err := url.Parse(rawURL)
	if id == "" || name == "" {
		return Subscription{}, errors.New("subscription id and name are required")
	}
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return Subscription{}, ErrInvalidURL
	}
	validTypes := make([]event.Type, 0, len(types))
	for _, value := range types {
		if value.Valid() {
			validTypes = append(validTypes, value)
		}
	}
	return Subscription{ID: id, Name: name, URL: rawURL, Active: true, EventTypes: validTypes, TerminalTags: terminal.NormalizeTags(tags), GeofenceIDs: append([]string(nil), geofenceIDs...), Secret: secret, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}, nil
}

func (s Subscription) Matches(e event.Event, tags map[string]string) bool {
	if !s.Active {
		return false
	}
	if len(s.EventTypes) > 0 {
		matched := false
		for _, value := range s.EventTypes {
			if value == e.Type {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if len(s.GeofenceIDs) > 0 {
		matched := false
		for _, value := range s.GeofenceIDs {
			if value == e.GeofenceID {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return terminal.MatchTags(tags, s.TerminalTags)
}

func Clone(s Subscription) Subscription {
	s.EventTypes = append([]event.Type(nil), s.EventTypes...)
	s.GeofenceIDs = append([]string(nil), s.GeofenceIDs...)
	s.TerminalTags = terminal.NormalizeTags(s.TerminalTags)
	return s
}

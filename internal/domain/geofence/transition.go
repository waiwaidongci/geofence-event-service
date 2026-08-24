package geofence

import (
	"time"

	"github.com/example/geofence-event-service/internal/domain/location"
)

type Transition string

const (
	TransitionEnter    Transition = "enter"
	TransitionExit     Transition = "exit"
	TransitionDwell    Transition = "dwell"
	TransitionCrossing Transition = "crossing"
)

type State struct {
	Inside       bool
	EnteredAt    *location.Location
	LastLocation *location.Location
	DwellSent    bool
}

func (s State) Clone() State {
	return s
}

func (s *State) Apply(current location.Location, inside bool, dwellSeconds int) []Transition {
	var transitions []Transition
	if s.LastLocation != nil && s.LastLocation.ObservedAt.After(current.ObservedAt) {
		return transitions
	}
	if !s.Inside && inside {
		transitions = append(transitions, TransitionEnter)
		entered := current
		s.EnteredAt = &entered
		s.DwellSent = false
	}
	if s.Inside && !inside {
		transitions = append(transitions, TransitionExit)
		s.EnteredAt = nil
		s.DwellSent = false
	}
	if inside && s.Inside && !s.DwellSent && s.EnteredAt != nil && dwellSeconds > 0 && current.ObservedAt.Sub(s.EnteredAt.ObservedAt) >= time.Duration(dwellSeconds)*time.Second {
		transitions = append(transitions, TransitionDwell)
		s.DwellSent = true
	}
	s.Inside = inside
	last := current
	s.LastLocation = &last
	return transitions
}

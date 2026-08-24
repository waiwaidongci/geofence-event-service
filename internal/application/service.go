package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/example/geofence-event-service/internal/domain/event"
	"github.com/example/geofence-event-service/internal/domain/geofence"
	"github.com/example/geofence-event-service/internal/domain/location"
	"github.com/example/geofence-event-service/internal/domain/subscription"
	"github.com/example/geofence-event-service/internal/domain/terminal"
)

var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource already exists")
)

type Service struct {
	repos        Repositories
	clock        Clock
	sender       WebhookSender
	dwellSeconds int
	mu           sync.Mutex
	states       map[string]geofence.State
}

func NewService(repos Repositories, clock Clock, sender WebhookSender, dwellSeconds int) *Service {
	if dwellSeconds < 0 {
		dwellSeconds = 0
	}
	return &Service{repos: repos, clock: clock, sender: sender, dwellSeconds: dwellSeconds, states: make(map[string]geofence.State)}
}

func (s *Service) CreateTerminal(ctx context.Context, id, name, description string, tags map[string]string) (terminal.Terminal, error) {
	value, err := terminal.New(id, name, description, tags, s.clock.Now())
	if err != nil {
		return terminal.Terminal{}, fmt.Errorf("create terminal: %w", err)
	}
	if _, exists := s.repos.Terminals.Get(ctx, value.ID); exists {
		return terminal.Terminal{}, fmt.Errorf("create terminal: %w", ErrConflict)
	}
	if err := s.repos.Terminals.Create(ctx, value); err != nil {
		return terminal.Terminal{}, fmt.Errorf("create terminal repository: %w", err)
	}
	return terminal.Clone(value), nil
}

func (s *Service) GetTerminal(ctx context.Context, id string) (terminal.Terminal, error) {
	value, ok := s.repos.Terminals.Get(ctx, id)
	if !ok {
		return terminal.Terminal{}, fmt.Errorf("get terminal: %w", ErrNotFound)
	}
	return value, nil
}

func (s *Service) ListTerminals(ctx context.Context, filter terminal.Filter) ([]terminal.Terminal, int) {
	return s.repos.Terminals.List(ctx, filter)
}

func (s *Service) UpdateTerminal(ctx context.Context, id, name, description string, tags map[string]string, status *terminal.Status) (terminal.Terminal, error) {
	value, err := s.GetTerminal(ctx, id)
	if err != nil {
		return terminal.Terminal{}, err
	}
	if err := value.Rename(name, description, tags, s.clock.Now()); err != nil {
		return terminal.Terminal{}, fmt.Errorf("update terminal: %w", err)
	}
	if status != nil {
		if err := value.ChangeStatus(*status, s.clock.Now()); err != nil {
			return terminal.Terminal{}, fmt.Errorf("update terminal status: %w", err)
		}
	}
	if err := s.repos.Terminals.Update(ctx, value); err != nil {
		return terminal.Terminal{}, fmt.Errorf("update terminal repository: %w", err)
	}
	return value, nil
}

func (s *Service) CreateGeofence(ctx context.Context, id, name, description string, shape geofence.Shape, tags map[string]string) (geofence.Geofence, error) {
	value, err := geofence.New(id, name, description, shape, tags, s.clock.Now())
	if err != nil {
		return geofence.Geofence{}, fmt.Errorf("create geofence: %w", err)
	}
	if _, exists := s.repos.Geofences.Get(ctx, id); exists {
		return geofence.Geofence{}, fmt.Errorf("create geofence: %w", ErrConflict)
	}
	if err := s.repos.Geofences.Create(ctx, value); err != nil {
		return geofence.Geofence{}, fmt.Errorf("create geofence repository: %w", err)
	}
	return geofence.Clone(value), nil
}

func (s *Service) GetGeofence(ctx context.Context, id string) (geofence.Geofence, error) {
	value, ok := s.repos.Geofences.Get(ctx, id)
	if !ok {
		return geofence.Geofence{}, fmt.Errorf("get geofence: %w", ErrNotFound)
	}
	return value, nil
}
func (s *Service) ListGeofences(ctx context.Context, filter geofence.Filter) ([]geofence.Geofence, int) {
	return s.repos.Geofences.List(ctx, filter)
}

func (s *Service) UpdateGeofence(ctx context.Context, id, name, description string, shape geofence.Shape, tags map[string]string, mode *geofence.Mode) (geofence.Geofence, error) {
	value, err := s.GetGeofence(ctx, id)
	if err != nil {
		return geofence.Geofence{}, err
	}
	if err := value.Update(name, description, shape, tags, s.clock.Now()); err != nil {
		return geofence.Geofence{}, fmt.Errorf("update geofence: %w", err)
	}
	if mode != nil {
		if err := value.SetMode(*mode, s.clock.Now()); err != nil {
			return geofence.Geofence{}, fmt.Errorf("update geofence mode: %w", err)
		}
	}
	if err := s.repos.Geofences.Update(ctx, value); err != nil {
		return geofence.Geofence{}, fmt.Errorf("update geofence repository: %w", err)
	}
	return value, nil
}

func (s *Service) SetGeofenceMode(ctx context.Context, id string, mode geofence.Mode) (geofence.Geofence, error) {
	value, err := s.GetGeofence(ctx, id)
	if err != nil {
		return geofence.Geofence{}, err
	}
	if err := value.SetMode(mode, s.clock.Now()); err != nil {
		return geofence.Geofence{}, err
	}
	if err := s.repos.Geofences.Update(ctx, value); err != nil {
		return geofence.Geofence{}, fmt.Errorf("set geofence mode: %w", err)
	}
	return value, nil
}

func (s *Service) ReportLocation(ctx context.Context, input location.Location) ([]event.Event, location.Location, error) {
	if input.ID == "" {
		input.ID = makeID("location", input.TerminalID, input.ObservedAt.String())
	}
	if _, err := s.GetTerminal(ctx, input.TerminalID); err != nil {
		return nil, location.Location{}, err
	}
	if err := s.repos.Locations.Add(ctx, input); err != nil {
		return nil, location.Location{}, fmt.Errorf("save location: %w", err)
	}
	term, _ := s.repos.Terminals.Get(ctx, input.TerminalID)
	term.RecordPosition(input.Point.Latitude, input.Point.Longitude, input.ObservedAt, s.clock.Now())
	_ = s.repos.Terminals.Update(ctx, term)
	geofences, _ := s.repos.Geofences.List(ctx, geofence.Filter{})
	var created []event.Event
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, fence := range geofences {
		key := input.TerminalID + ":" + fence.ID
		state := s.states[key]
		transitions := state.Apply(input, fence.Contains(input.Point), s.dwellSeconds)
		if len(transitions) == 0 && state.LastLocation != nil && fence.Crosses(state.LastLocation.Point, input.Point) {
			transitions = append(transitions, geofence.TransitionCrossing)
		}
		s.states[key] = state
		for _, transition := range transitions {
			dedup := fmt.Sprintf("%s:%s:%s:%s", input.TerminalID, fence.ID, transition, input.ObservedAt.UTC().Format(time.RFC3339Nano))
			if s.repos.Events.HasDeduplication(ctx, dedup) {
				continue
			}
			createdEvent, err := event.New(makeID("event", dedup), dedup, input.TerminalID, fence.ID, fence.Name, transition, input, s.clock.Now())
			if err != nil {
				return created, input, fmt.Errorf("build geofence event: %w", err)
			}
			if err := s.repos.Events.Create(ctx, createdEvent); err != nil {
				return created, input, fmt.Errorf("save geofence event: %w", err)
			}
			created = append(created, createdEvent)
			go s.deliver(createdEvent, term.Tags)
		}
	}
	return created, input, nil
}

func (s *Service) deliver(e event.Event, tags map[string]string) {
	if s.sender == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	subs, _ := s.repos.Subscriptions.List(ctx, true)
	for _, sub := range subs {
		if !sub.Matches(e, tags) {
			continue
		}
		delivery := subscription.Delivery{ID: makeID("delivery", sub.ID, e.ID), SubscriptionID: sub.ID, EventID: e.ID, Status: subscription.DeliveryPending, CreatedAt: s.clock.Now()}
		_ = s.repos.Deliveries.Create(ctx, delivery)
		status, err := s.sender.Send(ctx, sub, e)
		if err != nil {
			delivery.MarkSent(status, s.clock.Now())
		} else {
			delivery.MarkSent(status, s.clock.Now())
		}
		_ = s.repos.Deliveries.Update(ctx, delivery)
	}
}

func (s *Service) ListLocations(ctx context.Context, terminalID string, limit int) ([]location.Location, int) {
	return s.repos.Locations.List(ctx, terminalID, limit)
}
func (s *Service) GetEvent(ctx context.Context, id string) (event.Event, error) {
	value, ok := s.repos.Events.Get(ctx, id)
	if !ok {
		return event.Event{}, fmt.Errorf("get event: %w", ErrNotFound)
	}
	return value, nil
}
func (s *Service) ListEvents(ctx context.Context, filter event.Filter) ([]event.Event, int) {
	return s.repos.Events.List(ctx, filter)
}
func (s *Service) AcknowledgeEvent(ctx context.Context, id string) (event.Event, error) {
	return s.changeEvent(ctx, id, true)
}
func (s *Service) CloseEvent(ctx context.Context, id string) (event.Event, error) {
	return s.changeEvent(ctx, id, false)
}
func (s *Service) changeEvent(ctx context.Context, id string, acknowledge bool) (event.Event, error) {
	value, err := s.GetEvent(ctx, id)
	if err != nil {
		return event.Event{}, err
	}
	if acknowledge {
		err = value.Acknowledge(s.clock.Now())
	} else {
		err = value.Close(s.clock.Now())
	}
	if err != nil {
		return event.Event{}, fmt.Errorf("change event: %w", err)
	}
	if err := s.repos.Events.Update(ctx, value); err != nil {
		return event.Event{}, fmt.Errorf("save event: %w", err)
	}
	return value, nil
}

func (s *Service) CreateSubscription(ctx context.Context, id, name, rawURL string, types []event.Type, tags map[string]string, geofenceIDs []string, secret string) (subscription.Subscription, error) {
	value, err := subscription.New(id, name, rawURL, types, tags, geofenceIDs, secret, s.clock.Now())
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("create subscription: %w", err)
	}
	if _, exists := s.repos.Subscriptions.Get(ctx, id); exists {
		return subscription.Subscription{}, fmt.Errorf("create subscription: %w", ErrConflict)
	}
	if err := s.repos.Subscriptions.Create(ctx, value); err != nil {
		return subscription.Subscription{}, fmt.Errorf("create subscription repository: %w", err)
	}
	return value, nil
}
func (s *Service) ListSubscriptions(ctx context.Context, activeOnly bool) ([]subscription.Subscription, int) {
	return s.repos.Subscriptions.List(ctx, activeOnly)
}
func (s *Service) ListDeliveries(ctx context.Context, subscriptionID string, limit int) ([]subscription.Delivery, int) {
	return s.repos.Deliveries.List(ctx, subscriptionID, limit)
}

func makeID(prefix string, values ...string) string {
	hash := sha256.New()
	hash.Write([]byte(prefix))
	for _, value := range values {
		hash.Write([]byte{0})
		hash.Write([]byte(value))
	}
	return prefix + "_" + hex.EncodeToString(hash.Sum(nil))[:20]
}
func sortEvents(items []event.Event) {
	sort.Slice(items, func(i, j int) bool { return items[i].DetectedAt.After(items[j].DetectedAt) })
}
func normalizeLimit(limit int) int {
	if limit <= 0 || limit > 200 {
		return 50
	}
	return limit
}
func normalizeID(value string) string { return strings.TrimSpace(value) }

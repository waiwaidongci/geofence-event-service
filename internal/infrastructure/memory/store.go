package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/example/geofence-event-service/internal/application"
	"github.com/example/geofence-event-service/internal/domain/event"
	"github.com/example/geofence-event-service/internal/domain/geofence"
	"github.com/example/geofence-event-service/internal/domain/location"
	"github.com/example/geofence-event-service/internal/domain/subscription"
	"github.com/example/geofence-event-service/internal/domain/terminal"
)

type Store struct {
	mu            sync.RWMutex
	terminals     map[string]terminal.Terminal
	geofences     map[string]geofence.Geofence
	locations     map[string][]location.Location
	events        map[string]event.Event
	dedup         map[string]string
	subscriptions map[string]subscription.Subscription
	deliveries    map[string]subscription.Delivery
}

func NewStore() *Store {
	return &Store{terminals: map[string]terminal.Terminal{}, geofences: map[string]geofence.Geofence{}, locations: map[string][]location.Location{}, events: map[string]event.Event{}, dedup: map[string]string{}, subscriptions: map[string]subscription.Subscription{}, deliveries: map[string]subscription.Delivery{}}
}
func (s *Store) Create(_ context.Context, value terminal.Terminal) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.terminals[value.ID] = terminal.Clone(value)
	return nil
}
func (s *Store) Get(_ context.Context, id string) (terminal.Terminal, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.terminals[id]
	return terminal.Clone(value), ok
}
func (s *Store) ListTerminals(_ context.Context, filter terminal.Filter) ([]terminal.Terminal, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]terminal.Terminal, 0, len(s.terminals))
	for _, value := range s.terminals {
		if filter.Matches(value) {
			values = append(values, terminal.Clone(value))
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].ID < values[j].ID })
	total := len(values)
	return filter.Page(values), total
}
func (s *Store) Update(_ context.Context, value terminal.Terminal) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.terminals[value.ID] = terminal.Clone(value)
	return nil
}

func (s *Store) CreateGeofence(_ context.Context, value geofence.Geofence) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.geofences[value.ID] = geofence.Clone(value)
	return nil
}
func (s *Store) GetGeofence(_ context.Context, id string) (geofence.Geofence, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.geofences[id]
	return geofence.Clone(value), ok
}
func (s *Store) ListGeofences(_ context.Context, filter geofence.Filter) ([]geofence.Geofence, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]geofence.Geofence, 0, len(s.geofences))
	for _, value := range s.geofences {
		if filter.Matches(value) {
			values = append(values, geofence.Clone(value))
		}
	}
	geofence.Sort(values)
	return values, len(values)
}
func (s *Store) UpdateGeofence(_ context.Context, value geofence.Geofence) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.geofences[value.ID] = geofence.Clone(value)
	return nil
}

func (s *Store) Add(_ context.Context, value location.Location) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := s.locations[value.TerminalID]
	if len(items) > 0 && items[len(items)-1].ObservedAt.After(value.ObservedAt) {
		index := sort.Search(len(items), func(i int) bool { return !items[i].ObservedAt.Before(value.ObservedAt) })
		items = append(items, location.Location{})
		copy(items[index+1:], items[index:])
		items[index] = location.Clone(value)
	} else {
		items = append(items, location.Clone(value))
	}
	if len(items) > 1000 {
		items = items[len(items)-1000:]
	}
	s.locations[value.TerminalID] = items
	return nil
}
func (s *Store) Latest(_ context.Context, terminalID string) (location.Location, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := s.locations[terminalID]
	if len(items) == 0 {
		return location.Location{}, false
	}
	return location.Clone(items[len(items)-1]), true
}
func (s *Store) ListLocations(_ context.Context, terminalID string, limit int) ([]location.Location, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := s.locations[terminalID]
	total := len(items)
	limit = limitValue(limit)
	if limit > total {
		limit = total
	}
	result := make([]location.Location, 0, limit)
	for i := total - 1; i >= total-limit; i-- {
		result = append(result, items[i])
	}
	return result, total
}

func (s *Store) CreateEvent(_ context.Context, value event.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[value.ID] = application.SnapshotEvent(value)
	s.dedup[value.Deduplication] = value.ID
	return nil
}
func (s *Store) GetEvent(_ context.Context, id string) (event.Event, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.events[id]
	return value, ok
}
func (s *Store) ListEvents(_ context.Context, filter event.Filter) ([]event.Event, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]event.Event, 0, len(s.events))
	for _, value := range s.events {
		if filter.Matches(value) {
			values = append(values, value)
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].DetectedAt.After(values[j].DetectedAt) })
	total := len(values)
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	limit := limitValue(filter.Limit)
	end := offset + limit
	if end > total {
		end = total
	}
	return application.SnapshotEvents(values[offset:end]), total
}
func (s *Store) UpdateEvent(_ context.Context, value event.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[value.ID] = application.SnapshotEvent(value)
	return nil
}
func (s *Store) HasDeduplication(_ context.Context, key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.dedup[key]
	return ok
}

func (s *Store) CreateSubscription(_ context.Context, value subscription.Subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscriptions[value.ID] = subscription.Clone(value)
	return nil
}
func (s *Store) GetSubscription(_ context.Context, id string) (subscription.Subscription, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.subscriptions[id]
	return subscription.Clone(value), ok
}
func (s *Store) ListSubscriptions(_ context.Context, activeOnly bool) ([]subscription.Subscription, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]subscription.Subscription, 0, len(s.subscriptions))
	for _, value := range s.subscriptions {
		if !activeOnly || value.Active {
			values = append(values, subscription.Clone(value))
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].ID < values[j].ID })
	return values, len(values)
}
func (s *Store) UpdateSubscription(_ context.Context, value subscription.Subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscriptions[value.ID] = subscription.Clone(value)
	return nil
}

func (s *Store) CreateDelivery(_ context.Context, value subscription.Delivery) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deliveries[value.ID] = value
	return nil
}
func (s *Store) ListDeliveries(_ context.Context, subscriptionID string, limit int) ([]subscription.Delivery, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]subscription.Delivery, 0)
	for _, value := range s.deliveries {
		if subscriptionID == "" || value.SubscriptionID == subscriptionID {
			values = append(values, value)
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].CreatedAt.After(values[j].CreatedAt) })
	total := len(values)
	limit = limitValue(limit)
	if limit > total {
		limit = total
	}
	return values[:limit], total
}
func (s *Store) UpdateDelivery(_ context.Context, value subscription.Delivery) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deliveries[value.ID] = value
	return nil
}
func limitValue(value int) int {
	if value <= 0 || value > 200 {
		return 50
	}
	return value
}

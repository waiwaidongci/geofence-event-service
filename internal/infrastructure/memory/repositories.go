package memory

import (
	"context"
	"github.com/example/geofence-event-service/internal/application"
	"github.com/example/geofence-event-service/internal/domain/event"
	"github.com/example/geofence-event-service/internal/domain/geofence"
	"github.com/example/geofence-event-service/internal/domain/location"
	"github.com/example/geofence-event-service/internal/domain/subscription"
	"github.com/example/geofence-event-service/internal/domain/terminal"
)

type terminalRepo struct{ s *Store }

func (r terminalRepo) Create(c context.Context, v terminal.Terminal) error { return r.s.Create(c, v) }
func (r terminalRepo) Get(c context.Context, id string) (terminal.Terminal, bool) {
	return r.s.Get(c, id)
}
func (r terminalRepo) List(c context.Context, f terminal.Filter) ([]terminal.Terminal, int) {
	return r.s.ListTerminals(c, f)
}
func (r terminalRepo) Update(c context.Context, v terminal.Terminal) error { return r.s.Update(c, v) }

type geofenceRepo struct{ s *Store }

func (r geofenceRepo) Create(c context.Context, v geofence.Geofence) error {
	return r.s.CreateGeofence(c, v)
}
func (r geofenceRepo) Get(c context.Context, id string) (geofence.Geofence, bool) {
	return r.s.GetGeofence(c, id)
}
func (r geofenceRepo) List(c context.Context, f geofence.Filter) ([]geofence.Geofence, int) {
	return r.s.ListGeofences(c, f)
}
func (r geofenceRepo) Update(c context.Context, v geofence.Geofence) error {
	return r.s.UpdateGeofence(c, v)
}

type locationRepo struct{ s *Store }

func (r locationRepo) Add(c context.Context, v location.Location) error { return r.s.Add(c, v) }
func (r locationRepo) Latest(c context.Context, id string) (location.Location, bool) {
	return r.s.Latest(c, id)
}
func (r locationRepo) List(c context.Context, id string, l int) ([]location.Location, int) {
	return r.s.ListLocations(c, id, l)
}

type eventRepo struct{ s *Store }

func (r eventRepo) Create(c context.Context, v event.Event) error        { return r.s.CreateEvent(c, v) }
func (r eventRepo) Get(c context.Context, id string) (event.Event, bool) { return r.s.GetEvent(c, id) }
func (r eventRepo) List(c context.Context, f event.Filter) ([]event.Event, int) {
	return r.s.ListEvents(c, f)
}
func (r eventRepo) Update(c context.Context, v event.Event) error {
	previous, ok := r.s.GetEvent(c, v.ID)
	if ok {
		v.Status = previous.Status
	}
	return r.s.UpdateEvent(c, v)
}
func (r eventRepo) HasDeduplication(c context.Context, key string) bool {
	return r.s.HasDeduplication(c, key)
}

type subscriptionRepo struct{ s *Store }

func (r subscriptionRepo) Create(c context.Context, v subscription.Subscription) error {
	return r.s.CreateSubscription(c, v)
}
func (r subscriptionRepo) Get(c context.Context, id string) (subscription.Subscription, bool) {
	return r.s.GetSubscription(c, id)
}
func (r subscriptionRepo) List(c context.Context, a bool) ([]subscription.Subscription, int) {
	return r.s.ListSubscriptions(c, a)
}
func (r subscriptionRepo) Update(c context.Context, v subscription.Subscription) error {
	return r.s.UpdateSubscription(c, v)
}

type deliveryRepo struct{ s *Store }

func (r deliveryRepo) Create(c context.Context, v subscription.Delivery) error {
	return r.s.CreateDelivery(c, v)
}
func (r deliveryRepo) List(c context.Context, id string, l int) ([]subscription.Delivery, int) {
	return r.s.ListDeliveries(c, id, l)
}
func (r deliveryRepo) Update(c context.Context, v subscription.Delivery) error {
	return r.s.UpdateDelivery(c, v)
}

func (s *Store) Repositories() application.Repositories {
	return application.Repositories{Terminals: terminalRepo{s}, Geofences: geofenceRepo{s}, Locations: locationRepo{s}, Events: eventRepo{s}, Subscriptions: subscriptionRepo{s}, Deliveries: deliveryRepo{s}}
}

package application

import (
	"context"
	"time"

	"github.com/example/geofence-event-service/internal/domain/event"
	"github.com/example/geofence-event-service/internal/domain/geofence"
	"github.com/example/geofence-event-service/internal/domain/location"
	"github.com/example/geofence-event-service/internal/domain/subscription"
	"github.com/example/geofence-event-service/internal/domain/terminal"
)

type Clock interface{ Now() time.Time }

type TerminalRepository interface {
	Create(context.Context, terminal.Terminal) error
	Get(context.Context, string) (terminal.Terminal, bool)
	List(context.Context, terminal.Filter) ([]terminal.Terminal, int)
	Update(context.Context, terminal.Terminal) error
}

type GeofenceRepository interface {
	Create(context.Context, geofence.Geofence) error
	Get(context.Context, string) (geofence.Geofence, bool)
	List(context.Context, geofence.Filter) ([]geofence.Geofence, int)
	Update(context.Context, geofence.Geofence) error
}

type LocationRepository interface {
	Add(context.Context, location.Location) error
	Latest(context.Context, string) (location.Location, bool)
	List(context.Context, string, int) ([]location.Location, int)
}

type EventRepository interface {
	Create(context.Context, event.Event) error
	Get(context.Context, string) (event.Event, bool)
	List(context.Context, event.Filter) ([]event.Event, int)
	Update(context.Context, event.Event) error
	HasDeduplication(context.Context, string) bool
}

type SubscriptionRepository interface {
	Create(context.Context, subscription.Subscription) error
	Get(context.Context, string) (subscription.Subscription, bool)
	List(context.Context, bool) ([]subscription.Subscription, int)
	Update(context.Context, subscription.Subscription) error
}

type DeliveryRepository interface {
	Create(context.Context, subscription.Delivery) error
	List(context.Context, string, int) ([]subscription.Delivery, int)
	Update(context.Context, subscription.Delivery) error
}

type WebhookSender interface {
	Send(context.Context, subscription.Subscription, event.Event) (int, error)
}

type Repositories struct {
	Terminals     TerminalRepository
	Geofences     GeofenceRepository
	Locations     LocationRepository
	Events        EventRepository
	Subscriptions SubscriptionRepository
	Deliveries    DeliveryRepository
}

func SnapshotEvent(value event.Event) event.Event { return value }

func SnapshotEvents(values []event.Event) []event.Event { return values }

func SnapshotLocation(value location.Location) location.Location { return value }

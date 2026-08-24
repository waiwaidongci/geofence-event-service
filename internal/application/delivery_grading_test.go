package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/geofence-event-service/internal/domain/event"
	"github.com/example/geofence-event-service/internal/domain/subscription"
)

type fixedClock struct{ now time.Time }
func (c fixedClock) Now() time.Time { return c.now }

type failingSender struct{ err error }
func (s failingSender) Send(context.Context, subscription.Subscription, event.Event) (int, error) { return 503, s.err }

type subscriptionRepoStub struct{ values []subscription.Subscription }
func (r subscriptionRepoStub) Create(context.Context, subscription.Subscription) error { return nil }
func (r subscriptionRepoStub) Get(context.Context, string) (subscription.Subscription, bool) { return subscription.Subscription{}, false }
func (r subscriptionRepoStub) List(context.Context, bool) ([]subscription.Subscription, int) { return r.values, len(r.values) }
func (r subscriptionRepoStub) Update(context.Context, subscription.Subscription) error { return nil }

type deliveryRepoStub struct{ saved []subscription.Delivery }
func (r *deliveryRepoStub) Create(_ context.Context, value subscription.Delivery) error { r.saved = append(r.saved, value); return nil }
func (r *deliveryRepoStub) List(context.Context, string, int) ([]subscription.Delivery, int) { return r.saved, len(r.saved) }
func (r *deliveryRepoStub) Update(_ context.Context, value subscription.Delivery) error { r.saved = append(r.saved, value); return nil }

func TestServiceRecordsWebhookFailureP07(t *testing.T) {
	deliveries := &deliveryRepoStub{}
	subs := subscriptionRepoStub{values: []subscription.Subscription{{ID: "sub-1", Active: true}}}
	svc := NewService(Repositories{Subscriptions: subs, Deliveries: deliveries}, fixedClock{now: time.Now()}, failingSender{err: errors.New("network down")}, 0)
	svc.deliver(event.Event{ID: "event-1"}, nil)
	if len(deliveries.saved) != 2 {
		t.Fatalf("expected create and update, got %d writes", len(deliveries.saved))
	}
	result := deliveries.saved[1]
	if result.Status != subscription.DeliveryFailed || result.HTTPStatus != 503 || result.Error != "network down" || result.SentAt != nil {
		t.Fatalf("failed webhook recorded as success: %+v", result)
	}
}

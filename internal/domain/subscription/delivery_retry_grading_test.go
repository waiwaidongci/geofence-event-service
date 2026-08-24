package subscription

import (
	"errors"
	"testing"
	"time"
)

func TestDeliveryRetrySuccessStateP05(t *testing.T) {
	now := time.Now()
	delivery := Delivery{Status: DeliveryFailed, Attempt: 1, HTTPStatus: 503, Error: "unavailable"}
	delivery.MarkSent(204, now)
	if delivery.Status != DeliverySent || delivery.Attempt != 2 || delivery.HTTPStatus != 204 {
		t.Fatalf("unexpected sent state: %+v", delivery)
	}
	if delivery.Error != "" || delivery.SentAt == nil || !delivery.SentAt.Equal(now.UTC()) {
		t.Fatalf("successful retry retained failure state: %+v", delivery)
	}
}

func TestDeliveryFailureResetsTerminalStateP05(t *testing.T) {
	sentAt := time.Now().Add(-time.Minute)
	delivery := Delivery{Status: DeliverySent, Attempt: 1, HTTPStatus: 200, Error: "stale", SentAt: &sentAt}
	delivery.MarkFailed(errors.New("timeout"), 504)
	if delivery.Status != DeliveryFailed || delivery.Attempt != 2 || delivery.HTTPStatus != 504 {
		t.Fatalf("unexpected failed state: %+v", delivery)
	}
	if delivery.SentAt != nil || delivery.Error != "timeout" {
		t.Fatalf("failed retry retained terminal state: %+v", delivery)
	}
	delivery.Error = "stale"
	delivery.MarkFailed(nil, 503)
	if delivery.Error != "" || delivery.Attempt != 3 {
		t.Fatalf("status-only failure retained stale error: %+v", delivery)
	}
}

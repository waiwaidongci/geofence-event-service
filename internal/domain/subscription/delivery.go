package subscription

import "time"

type DeliveryStatus string

const (
	DeliveryPending DeliveryStatus = "pending"
	DeliverySent    DeliveryStatus = "sent"
	DeliveryFailed  DeliveryStatus = "failed"
)

type Delivery struct {
	ID             string         `json:"id"`
	SubscriptionID string         `json:"subscription_id"`
	EventID        string         `json:"event_id"`
	Status         DeliveryStatus `json:"status"`
	Attempt        int            `json:"attempt"`
	HTTPStatus     int            `json:"http_status,omitempty"`
	Error          string         `json:"error,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	SentAt         *time.Time     `json:"sent_at,omitempty"`
}

func (d *Delivery) MarkSent(status int, now time.Time) {
	d.Status = DeliverySent
	d.HTTPStatus = status
	d.Attempt++
}

func (d *Delivery) MarkFailed(err error, status int) {
	d.Status = DeliveryFailed
	d.HTTPStatus = status
	d.Attempt++
	if err != nil {
		d.Error = err.Error()
	}
}

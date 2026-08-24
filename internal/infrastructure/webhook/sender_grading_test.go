package webhook

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/example/geofence-event-service/internal/domain/event"
	"github.com/example/geofence-event-service/internal/domain/subscription"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

type trackedBody struct {
	reader   io.Reader
	closed   bool
	readErr  error
	closeErr error
}

func (b *trackedBody) Read(p []byte) (int, error) {
	if b.readErr != nil {
		return 0, b.readErr
	}
	return b.reader.Read(p)
}
func (b *trackedBody) Close() error { b.closed = true; return b.closeErr }

func testSubscription() subscription.Subscription { return subscription.Subscription{URL: "https://example.com/hook"} }

func TestSenderPropagatesRequestContextP07(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	seenCanceled := false
	sender := &Sender{MaxBody: 1024, Client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		seenCanceled = errors.Is(req.Context().Err(), context.Canceled)
		return nil, req.Context().Err()
	})}}
	_, _ = sender.Send(ctx, testSubscription(), event.Event{})
	if !seenCanceled {
		t.Fatal("request did not receive caller cancellation")
	}
}

func TestSenderPreservesTransportErrorChainP07(t *testing.T) {
	sentinel := errors.New("transport unavailable")
	sender := &Sender{MaxBody: 1024, Client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, sentinel })}}
	_, err := sender.Send(context.Background(), testSubscription(), event.Event{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("transport error chain lost: %v", err)
	}
}

func TestSenderClosesResponseBodyP07(t *testing.T) {
	body := &trackedBody{reader: strings.NewReader("ok")}
	sender := &Sender{MaxBody: 1024, Client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 204, Body: body}, nil
	})}}
	if _, err := sender.Send(context.Background(), testSubscription(), event.Event{}); err != nil {
		t.Fatal(err)
	}
	if !body.closed {
		t.Fatal("response body was not closed")
	}
}

func TestSenderReturnsBodyReadAndCloseErrorsP07(t *testing.T) {
	readErr := errors.New("read response")
	closeErr := errors.New("close response")
	body := &trackedBody{reader: strings.NewReader(""), readErr: readErr, closeErr: closeErr}
	sender := &Sender{MaxBody: 1024, Client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 202, Body: body}, nil
	})}}
	_, err := sender.Send(context.Background(), testSubscription(), event.Event{})
	if !errors.Is(err, readErr) || !errors.Is(err, closeErr) || !body.closed {
		t.Fatalf("response cleanup errors lost: %v closed=%v", err, body.closed)
	}
}

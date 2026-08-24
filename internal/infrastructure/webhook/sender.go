package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/example/geofence-event-service/internal/domain/event"
	"github.com/example/geofence-event-service/internal/domain/subscription"
)

type Sender struct {
	Client  *http.Client
	MaxBody int64
}

func NewSender() *Sender {
	return &Sender{Client: &http.Client{Timeout: 10 * time.Second}, MaxBody: 1 << 20}
}
func (s *Sender) Send(ctx context.Context, sub subscription.Subscription, e event.Event) (int, error) {
	body, err := json.Marshal(map[string]any{"event": e, "sent_at": time.Now().UTC()})
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sub.URL, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "geofence-event-service/1.0")
	if sub.Secret != "" {
		mac := hmac.New(sha256.New, []byte(sub.Secret))
		mac.Write(body)
		req.Header.Set("X-Geofence-Signature", hex.EncodeToString(mac.Sum(nil)))
	}
	res, err := s.Client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	io.CopyN(io.Discard, res.Body, s.MaxBody)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.StatusCode, fmt.Errorf("webhook returned status %d", res.StatusCode)
	}
	return res.StatusCode, nil
}

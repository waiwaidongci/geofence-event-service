package application

import (
	"context"
	"errors"
	"sort"

	"github.com/example/geofence-event-service/internal/domain/event"
	"github.com/example/geofence-event-service/internal/domain/location"
)

var ErrReplayTerminalMismatch = errors.New("replay locations must belong to one terminal")

type ReplayResult struct {
	Checkpoint ReplayCheckpoint `json:"checkpoint"`
	Events     []event.Event    `json:"events"`
}

type ReplayCoordinator struct {
	service     *Service
	checkpoints *replayCheckpointStore
}

func NewReplayCoordinator(service *Service) *ReplayCoordinator {
	return &ReplayCoordinator{service: service, checkpoints: newReplayCheckpointStore()}
}

func (c *ReplayCoordinator) Replay(ctx context.Context, terminalID string, inputs []location.Location) (ReplayResult, error) {
	items := append([]location.Location(nil), inputs...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].ObservedAt.Before(items[j].ObservedAt) })
	checkpoint, _ := c.checkpoints.Get(terminalID)
	result := ReplayResult{Checkpoint: checkpoint}
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return ReplayResult{}, err
		}
		if item.TerminalID != terminalID {
			return ReplayResult{}, ErrReplayTerminalMismatch
		}
		if !checkpoint.ObservedAt.IsZero() && !item.ObservedAt.After(checkpoint.ObservedAt) {
			continue
		}
		generated, _, err := c.service.ReportLocation(ctx, item)
		if err != nil {
			return ReplayResult{}, err
		}
		result.Events = append(result.Events, generated...)
		checkpoint = ReplayCheckpoint{TerminalID: terminalID, ObservedAt: item.ObservedAt, Processed: checkpoint.Processed + 1}
	}
	if !checkpoint.ObservedAt.IsZero() {
		c.checkpoints.Commit(checkpoint)
	}
	result.Checkpoint = checkpoint
	return result, nil
}

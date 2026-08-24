package application

import (
	"sync"
	"time"
)

type ReplayCheckpoint struct {
	TerminalID string    `json:"terminal_id"`
	ObservedAt time.Time `json:"observed_at"`
	Processed  int       `json:"processed"`
}

type replayCheckpointStore struct {
	mu     sync.RWMutex
	values map[string]ReplayCheckpoint
}

func newReplayCheckpointStore() *replayCheckpointStore {
	return &replayCheckpointStore{values: make(map[string]ReplayCheckpoint)}
}

func (s *replayCheckpointStore) Get(terminalID string) (ReplayCheckpoint, bool) {
	return ReplayCheckpoint{}, false
}

func (s *replayCheckpointStore) Commit(value ReplayCheckpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[value.TerminalID] = value
}

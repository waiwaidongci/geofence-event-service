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
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[terminalID]
	return value, ok
}

func (s *replayCheckpointStore) Commit(value ReplayCheckpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[value.TerminalID] = value
}

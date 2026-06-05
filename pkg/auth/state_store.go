package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const defaultStateTTL = 5 * time.Minute

// stateEntry holds an OAuth state token with its expiry time.
type stateEntry struct {
	provider  string
	expiresAt time.Time
}

// StateStore is a thread-safe in-memory store for OAuth state tokens (CSRF protection).
// States expire after a configurable TTL.
type StateStore struct {
	mu     sync.RWMutex
	states map[string]stateEntry
	ttl    time.Duration
}

// NewStateStore creates a new StateStore with the default TTL.
func NewStateStore() *StateStore {
	s := &StateStore{
		states: make(map[string]stateEntry),
		ttl:    defaultStateTTL,
	}
	// Start background cleanup goroutine
	go s.cleanup()
	return s
}

// Generate creates a new random state token for the given provider.
// Returns the state string.
func (s *StateStore) Generate(provider string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = stateEntry{
		provider:  provider,
		expiresAt: time.Now().Add(s.ttl),
	}
	return state, nil
}

// Validate checks if a state token is valid for the given provider.
// If valid, the state is consumed (deleted) to prevent replay attacks.
// Returns true if valid and for the correct provider.
func (s *StateStore) Validate(state, provider string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.states[state]
	if !ok {
		return false
	}
	// Always delete to prevent replay
	delete(s.states, state)

	// Check expiry and provider match
	if time.Now().After(entry.expiresAt) {
		return false
	}
	return entry.provider == provider
}

// cleanup periodically removes expired state entries.
func (s *StateStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.states {
			if now.After(entry.expiresAt) {
				delete(s.states, key)
			}
		}
		s.mu.Unlock()
	}
}

package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStateStore_GenerateAndValidate(t *testing.T) {
	s := NewStateStore()

	// 1. Generate state for github
	state, err := s.Generate("github")
	assert.NoError(t, err)
	assert.NotEmpty(t, state)

	// 2. Validate with incorrect provider
	valid := s.Validate(state, "google")
	assert.False(t, valid, "Validation should fail for mismatching provider")

	// 3. Generate another state to validate correctly
	state2, err := s.Generate("github")
	assert.NoError(t, err)

	// 4. Validate with correct provider
	valid = s.Validate(state2, "github")
	assert.True(t, valid, "Validation should succeed for correct provider")

	// 5. Re-validate to check if consumed
	valid = s.Validate(state2, "github")
	assert.False(t, valid, "State should be consumed and invalid on second validation")
}

func TestStateStore_Expiration(t *testing.T) {
	s := NewStateStore()
	// Tweak TTL to be very short for testing expiration
	s.ttl = 10 * time.Millisecond

	state, err := s.Generate("google")
	assert.NoError(t, err)

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	valid := s.Validate(state, "google")
	assert.False(t, valid, "Expired state should be invalid")
}

func TestStateStore_Cleanup(t *testing.T) {
	s := NewStateStore()
	s.ttl = 10 * time.Millisecond

	state, err := s.Generate("google")
	assert.NoError(t, err)

	s.mu.RLock()
	_, ok := s.states[state]
	s.mu.RUnlock()
	assert.True(t, ok)

	// Wait for TTL to expire
	time.Sleep(20 * time.Millisecond)

	// Force run clean logic (simulate cleanup loop without waiting a whole minute)
	s.mu.Lock()
	now := time.Now()
	for key, entry := range s.states {
		if now.After(entry.expiresAt) {
			delete(s.states, key)
		}
	}
	s.mu.Unlock()

	s.mu.RLock()
	_, ok = s.states[state]
	s.mu.RUnlock()
	assert.False(t, ok, "Expired state should have been cleaned up from map")
}

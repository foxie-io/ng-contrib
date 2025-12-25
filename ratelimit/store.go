package ratelimit

import (
	"sync"

	"github.com/foxie-io/ng-contrib/ratelimit/limiter"
)

// Store defines how client rate limit data is stored and retrieved.
type Store interface {
	// Get returns the limiter for the given client ID.
	// If not found, returns nil.
	Get(id string) limiter.Limiter
	// Set stores/updates the limiter for the given client ID.
	Set(id string, l limiter.Limiter)
	// Delete removes the client data (used for cleanup).
	Delete(id string)
	// Keys returns all client IDs (optional: for cleanup).
	Keys() []string
}

// inMemoryStore implements Store using a Go map with a mutex.
type inMemoryStore struct {
	data  map[string]limiter.Limiter
	mutex sync.RWMutex
}

// NewInMemoryStore creates a new in-memory store.
func NewInMemoryStore() Store {
	return &inMemoryStore{
		data: make(map[string]limiter.Limiter),
	}
}

func (s *inMemoryStore) Get(id string) limiter.Limiter {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data[id]
}

func (s *inMemoryStore) Set(id string, l limiter.Limiter) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[id] = l
}

func (s *inMemoryStore) Delete(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.data, id)
}

func (s *inMemoryStore) Keys() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

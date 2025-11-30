package kvstore

import (
	"sync"

	"github.com/yourusername/distributed-kv-store/internal/vectorclock"
)

// Store is an in-memory key-value store with versioning
type Store struct {
	mu      sync.RWMutex
	data    map[string]*KeyValue
	version int64 // Global version counter (kept for backward compatibility)
}

// KeyValue represents a key-value pair with version and optional vector clock
type KeyValue struct {
	Key        string
	Value      string
	Version    int64                    // Legacy version number
	VectorClock vectorclock.VectorClock // Vector clock for conflict resolution
}

// NewStore creates a new in-memory key-value store
func NewStore() *Store {
	return &Store{
		data:    make(map[string]*KeyValue),
		version: 0,
	}
}

// Set stores a value under the given key
// Returns the version number and an error if key is empty
func (s *Store) Set(key, value string) (int64, error) {
	return s.SetWithVectorClock(key, value, nil)
}

// SetWithVectorClock stores a value with an optional vector clock
func (s *Store) SetWithVectorClock(key, value string, vc vectorclock.VectorClock) (int64, error) {
	if key == "" {
		return 0, ErrEmptyKey
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.version++
	kv := &KeyValue{
		Key:        key,
		Value:      value,
		Version:    s.version,
		VectorClock: vc,
	}
	s.data[key] = kv

	return s.version, nil
}

// SetWithVersion stores a value with a specific version (used for replication)
// Updates the global version counter if the provided version is higher
func (s *Store) SetWithVersion(key, value string, version int64) error {
	return s.SetWithVersionAndClock(key, value, version, nil)
}

// SetWithVersionAndClock stores a value with version and vector clock
// Uses vector clock for conflict resolution: only accepts if clock is newer or concurrent
func (s *Store) SetWithVersionAndClock(key, value string, version int64, vc vectorclock.VectorClock) error {
	if key == "" {
		return ErrEmptyKey
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key exists and compare vector clocks
	existing, exists := s.data[key]
	if exists && existing.VectorClock != nil && vc != nil {
		// Compare vector clocks: only accept if new clock is newer or concurrent
		comparison := vectorclock.Compare(vc, existing.VectorClock)
		if comparison < 0 {
			// New clock is older, reject the write (prevent false write)
			return ErrOlderClock
		}
		// If comparison >= 0, accept (newer or concurrent)
	}

	// Update global version if this version is higher
	if version > s.version {
		s.version = version
	}

	kv := &KeyValue{
		Key:        key,
		Value:      value,
		Version:    version,
		VectorClock: vc,
	}
	s.data[key] = kv

	return nil
}

// Get retrieves the value for the given key
// Returns the KeyValue and a boolean indicating if the key exists
func (s *Store) Get(key string) (*KeyValue, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	kv, exists := s.data[key]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	copy := &KeyValue{
		Key:     kv.Key,
		Value:   kv.Value,
		Version: kv.Version,
	}
	if kv.VectorClock != nil {
		// Deep copy vector clock
		copy.VectorClock = make(vectorclock.VectorClock)
		for k, v := range kv.VectorClock {
			copy.VectorClock[k] = v
		}
	}
	return copy, true
}

// LocalRead returns the local value without any coordination
// Used for testing inconsistency windows
func (s *Store) LocalRead(key string) (*KeyValue, bool) {
	return s.Get(key)
}

// GetVersion returns the current global version counter
func (s *Store) GetVersion() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// Errors
var (
	ErrEmptyKey   = &KVError{Message: "key cannot be empty"}
	ErrOlderClock = &KVError{Message: "vector clock is older than existing value"}
)

type KVError struct {
	Message string
}

func (e *KVError) Error() string {
	return e.Message
}

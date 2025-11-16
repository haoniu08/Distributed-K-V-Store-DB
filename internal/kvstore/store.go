package kvstore

import (
	"sync"
)

// Store is an in-memory key-value store with versioning
type Store struct {
	mu      sync.RWMutex
	data    map[string]*KeyValue
	version int64 // Global version counter
}

// KeyValue represents a key-value pair with version
type KeyValue struct {
	Key     string
	Value   string
	Version int64
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
	if key == "" {
		return 0, ErrEmptyKey
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.version++
	kv := &KeyValue{
		Key:     key,
		Value:   value,
		Version: s.version,
	}
	s.data[key] = kv

	return s.version, nil
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
	return &KeyValue{
		Key:     kv.Key,
		Value:   kv.Value,
		Version: kv.Version,
	}, true
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
	ErrEmptyKey = &KVError{Message: "key cannot be empty"}
)

type KVError struct {
	Message string
}

func (e *KVError) Error() string {
	return e.Message
}

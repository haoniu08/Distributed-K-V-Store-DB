package leaderless

import (
	"fmt"
	"sync"

	"github.com/yourusername/distributed-kv-store/internal/kvstore"
)

// ReplicationManager handles replication for leaderless database
type ReplicationManager struct {
	store  *kvstore.Store
	config *Config
	client *ReplicationClient
	mu     sync.Mutex
}

// NewReplicationManager creates a new replication manager
func NewReplicationManager(store *kvstore.Store, config *Config) *ReplicationManager {
	return &ReplicationManager{
		store:  store,
		config: config,
		client: NewReplicationClient(),
	}
}

// WriteResult represents the result of a write operation
type WriteResult struct {
	Version int64
	Success bool
	Error   error
}

// WriteWithCoordination implements W=N strategy
// When a node receives a write, it becomes the Write Coordinator
// and must write to all other nodes (W=N)
func (rm *ReplicationManager) WriteWithCoordination(key, value string) (*WriteResult, error) {
	// Coordinator sets the value locally first
	version, err := rm.store.Set(key, value)
	if err != nil {
		return nil, err
	}

	// Get addresses of all other nodes
	otherNodeAddrs := rm.config.GetOtherNodeAddrs()
	
	if len(otherNodeAddrs) == 0 {
		// Only one node, no replication needed
		return &WriteResult{Version: version, Success: true}, nil
	}

	// Replicate to all other nodes
	results := make(chan *ReplicateWriteResponse, len(otherNodeAddrs))

	// Send replication requests to all other nodes
	for i, addr := range otherNodeAddrs {
		go func(addr string, index int) {
			// Coordinator sleeps 200ms after each message (except the first one)
			response, err := rm.client.ReplicateWrite(addr, key, value, version, index > 0)
			if err != nil {
				results <- &ReplicateWriteResponse{Success: false, Error: err.Error()}
				return
			}
			results <- response
		}(addr, i)
	}

	// Wait for all other nodes to confirm (W=N means all nodes must be updated)
	successCount := 1 // Coordinator already updated
	for i := 0; i < len(otherNodeAddrs); i++ {
		result := <-results
		if result.Success {
			successCount++
		}
	}

	// All nodes (N) must be updated
	n := rm.config.GetN()
	if successCount < n {
		return nil, fmt.Errorf("failed to replicate to all nodes: %d/%d succeeded", successCount, n)
	}

	return &WriteResult{Version: version, Success: true}, nil
}

// ReadLocal implements R=1 strategy
// Returns the local value immediately (no coordination)
func (rm *ReplicationManager) ReadLocal(key string) (*kvstore.KeyValue, error) {
	kv, exists := rm.store.Get(key)
	if !exists {
		return nil, fmt.Errorf("key not found")
	}
	return kv, nil
}


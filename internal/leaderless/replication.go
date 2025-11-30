package leaderless

import (
	"fmt"
	"sync"

	"github.com/yourusername/distributed-kv-store/internal/kvstore"
	"github.com/yourusername/distributed-kv-store/internal/vectorclock"
)

// ReplicationManager handles replication for leaderless database
type ReplicationManager struct {
	store       *kvstore.Store
	config      *Config
	client      *ReplicationClient
	clockManager *vectorclock.ClockManager
	useVectorClock bool
	mu          sync.Mutex
}

// NewReplicationManager creates a new replication manager
func NewReplicationManager(store *kvstore.Store, config *Config) *ReplicationManager {
	// Get all node addresses
	allNodeAddrs := config.GetAllNodeAddrs()
	myAddr := config.GetMyAddr()
	nodeID := config.NodeID
	
	// For vector clocks, we need node IDs for all nodes
	// Since we only have addresses, we'll derive node IDs from addresses
	// Extract hostname/identifier from address (e.g., "node1:8080" -> "node1")
	allNodeIDs := make([]string, 0, len(allNodeAddrs))
	addrToID := make(map[string]string)
	
	for _, addr := range allNodeAddrs {
		// Extract node ID from address (part before colon)
		derivedID := addr
		for i, ch := range addr {
			if ch == ':' {
				derivedID = addr[:i]
				break
			}
		}
		addrToID[addr] = derivedID
		allNodeIDs = append(allNodeIDs, derivedID)
	}
	
	// Use the node ID from config, or derive from address
	if nodeID == "" {
		nodeID = addrToID[myAddr]
		if nodeID == "" {
			nodeID = myAddr // Final fallback
		}
	}
	
	clockManager := vectorclock.NewClockManager(nodeID, allNodeIDs)
	
	return &ReplicationManager{
		store:          store,
		config:        config,
		client:         NewReplicationClient(),
		clockManager:   clockManager,
		useVectorClock: true, // Enable vector clocks by default
	}
}

// UpdateClock updates the clock manager with a received vector clock
func (rm *ReplicationManager) UpdateClock(received vectorclock.VectorClock) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if rm.clockManager != nil && received != nil {
		rm.clockManager.Update(received)
	}
}

// SetUseVectorClock enables or disables vector clock usage
func (rm *ReplicationManager) SetUseVectorClock(use bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.useVectorClock = use
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
	rm.mu.Lock()
	
	var vc vectorclock.VectorClock
	if rm.useVectorClock {
		// Tick the clock for local write event
		vc = rm.clockManager.Tick()
	}
	
	rm.mu.Unlock()
	
	// Coordinator sets the value locally first
	var version int64
	var err error
	if rm.useVectorClock {
		version, err = rm.store.SetWithVectorClock(key, value, vc)
	} else {
		version, err = rm.store.Set(key, value)
	}
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
			response, err := rm.client.ReplicateWrite(addr, key, value, version, vc, index > 0)
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


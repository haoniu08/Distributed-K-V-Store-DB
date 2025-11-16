package leaderfollower

import (
	"fmt"
	"sync"

	"github.com/yourusername/distributed-kv-store/internal/kvstore"
)

// ReplicationManager handles replication strategies
type ReplicationManager struct {
	store    *kvstore.Store
	config   *Config
	client   *ReplicationClient
	mu       sync.Mutex
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

// WriteStrategyW5R1 implements W=5, R=1 strategy
// Write: All nodes must be updated before responding
func (rm *ReplicationManager) WriteStrategyW5R1(key, value string) (*WriteResult, error) {
	if !rm.config.IsLeader() {
		return nil, fmt.Errorf("only leader can perform writes")
	}

	// Leader sets the value locally first
	version, err := rm.store.Set(key, value)
	if err != nil {
		return nil, err
	}

	// Replicate to all followers
	followerAddrs := rm.config.GetFollowerAddrs()
	results := make(chan *ReplicateWriteResponse, len(followerAddrs))

	// Send replication requests to all followers
	for i, addr := range followerAddrs {
		go func(addr string, index int) {
			// Leader sleeps 200ms after each message (except the first one)
			response, err := rm.client.ReplicateWrite(addr, key, value, version, index > 0)
			if err != nil {
				results <- &ReplicateWriteResponse{Success: false, Error: err.Error()}
				return
			}
			results <- response
		}(addr, i)
	}

	// Wait for all followers to confirm
	successCount := 1 // Leader already updated
	for i := 0; i < len(followerAddrs); i++ {
		result := <-results
		if result.Success {
			successCount++
		}
	}

	// All nodes (N=5) must be updated
	if successCount < rm.config.N {
		return nil, fmt.Errorf("failed to replicate to all nodes: %d/%d succeeded", successCount, rm.config.N)
	}

	return &WriteResult{Version: version, Success: true}, nil
}

// WriteStrategyW1R5 implements W=1, R=5 strategy
// Write: Only Leader needs to be updated
func (rm *ReplicationManager) WriteStrategyW1R5(key, value string) (*WriteResult, error) {
	if !rm.config.IsLeader() {
		return nil, fmt.Errorf("only leader can perform writes")
	}

	// Leader sets the value locally and responds immediately
	version, err := rm.store.Set(key, value)
	if err != nil {
		return nil, err
	}

	// Replicate to followers asynchronously (don't wait)
	followerAddrs := rm.config.GetFollowerAddrs()
	for i, addr := range followerAddrs {
		go func(addr string, index int) {
			rm.client.ReplicateWrite(addr, key, value, version, index > 0)
		}(addr, i)
	}

	return &WriteResult{Version: version, Success: true}, nil
}

// WriteStrategyW3R3 implements W=3, R=3 quorum strategy
// Write: 3 nodes (including Leader) must be updated
func (rm *ReplicationManager) WriteStrategyW3R3(key, value string) (*WriteResult, error) {
	if !rm.config.IsLeader() {
		return nil, fmt.Errorf("only leader can perform writes")
	}

	// Leader sets the value locally first
	version, err := rm.store.Set(key, value)
	if err != nil {
		return nil, err
	}

	// Replicate to followers
	followerAddrs := rm.config.GetFollowerAddrs()
	results := make(chan *ReplicateWriteResponse, len(followerAddrs))

	// Send replication requests to all followers
	for i, addr := range followerAddrs {
		go func(addr string, index int) {
			response, err := rm.client.ReplicateWrite(addr, key, value, version, index > 0)
			if err != nil {
				results <- &ReplicateWriteResponse{Success: false, Error: err.Error()}
				return
			}
			results <- response
		}(addr, i)
	}

	// Wait for W-1 followers to confirm (Leader already counts as 1)
	successCount := 1 // Leader already updated
	for i := 0; i < len(followerAddrs); i++ {
		result := <-results
		if result.Success {
			successCount++
			if successCount >= rm.config.W {
				break // We have quorum
			}
		}
	}

	if successCount < rm.config.W {
		return nil, fmt.Errorf("failed to achieve write quorum: %d/%d succeeded", successCount, rm.config.W)
	}

	return &WriteResult{Version: version, Success: true}, nil
}

// ReadStrategyR1 reads from a single node (typically Leader)
func (rm *ReplicationManager) ReadStrategyR1(key string) (*kvstore.KeyValue, error) {
	kv, exists := rm.store.Get(key)
	if !exists {
		return nil, fmt.Errorf("key not found")
	}
	return kv, nil
}

// ReadStrategyR5 reads from all nodes and returns most recent
func (rm *ReplicationManager) ReadStrategyR5(key string) (*kvstore.KeyValue, error) {
	allAddrs := rm.config.GetAllNodeAddrs()
	results := make(chan *kvstore.KeyValue, len(allAddrs))

	// Read from all nodes concurrently
	myAddr := rm.config.GetMyAddr()
	for _, addr := range allAddrs {
		go func(addr string) {
			if addr == myAddr {
				// Read from local store
				kv, exists := rm.store.Get(key)
				if exists {
					results <- kv
				} else {
					results <- nil
				}
			} else {
				// Read from remote node (Follower sleeps 50ms)
				response, err := rm.client.ReadFromNode(addr, key, true)
				if err != nil || !response.Exists {
					results <- nil
					return
				}
				results <- &kvstore.KeyValue{
					Key:     response.Key,
					Value:   response.Value,
					Version: response.Version,
				}
			}
		}(addr)
	}

	// Collect all responses
	var responses []*kvstore.KeyValue
	for i := 0; i < len(allAddrs); i++ {
		kv := <-results
		if kv != nil {
			responses = append(responses, kv)
		}
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("key not found")
	}

	// Return the most recent version
	return getMostRecentValue(responses), nil
}

// ReadStrategyR3 reads from 3 nodes (quorum) and returns most recent
func (rm *ReplicationManager) ReadStrategyR3(key string) (*kvstore.KeyValue, error) {
	allAddrs := rm.config.GetAllNodeAddrs()
	results := make(chan *kvstore.KeyValue, len(allAddrs))

	// Read from all nodes concurrently
	myAddr := rm.config.GetMyAddr()
	for _, addr := range allAddrs {
		go func(addr string) {
			if addr == myAddr {
				// Read from local store
				kv, exists := rm.store.Get(key)
				if exists {
					results <- kv
				} else {
					results <- nil
				}
			} else {
				// Read from remote node (Follower sleeps 50ms)
				response, err := rm.client.ReadFromNode(addr, key, true)
				if err != nil || !response.Exists {
					results <- nil
					return
				}
				results <- &kvstore.KeyValue{
					Key:     response.Key,
					Value:   response.Value,
					Version: response.Version,
				}
			}
		}(addr)
	}

	// Collect R responses (quorum)
	var responses []*kvstore.KeyValue
	needed := rm.config.R
	for i := 0; i < len(allAddrs) && len(responses) < needed; i++ {
		kv := <-results
		if kv != nil {
			responses = append(responses, kv)
		}
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("key not found")
	}

	// Return the most recent version
	return getMostRecentValue(responses), nil
}

// getMostRecentValue compares multiple KeyValue responses and returns the one with highest version
func getMostRecentValue(responses []*kvstore.KeyValue) *kvstore.KeyValue {
	if len(responses) == 0 {
		return nil
	}

	mostRecent := responses[0]
	for _, kv := range responses[1:] {
		if kv != nil && kv.Version > mostRecent.Version {
			mostRecent = kv
		}
	}

	return mostRecent
}

// Write performs a write operation based on current W value
func (rm *ReplicationManager) Write(key, value string) (*WriteResult, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	_, w := rm.config.GetReplicationParams()

	switch w {
	case 5:
		return rm.WriteStrategyW5R1(key, value)
	case 1:
		return rm.WriteStrategyW1R5(key, value)
	case 3:
		return rm.WriteStrategyW3R3(key, value)
	default:
		return nil, fmt.Errorf("unsupported W value: %d", w)
	}
}

// Read performs a read operation based on current R value
func (rm *ReplicationManager) Read(key string) (*kvstore.KeyValue, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	r, _ := rm.config.GetReplicationParams()

	switch r {
	case 1:
		return rm.ReadStrategyR1(key)
	case 5:
		return rm.ReadStrategyR5(key)
	case 3:
		return rm.ReadStrategyR3(key)
	default:
		return nil, fmt.Errorf("unsupported R value: %d", r)
	}
}


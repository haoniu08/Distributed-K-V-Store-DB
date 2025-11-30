package vectorclock

import (
	"fmt"
	"sort"
	"sync"
)

// VectorClock represents a vector clock for distributed systems
// Maps node ID to logical timestamp
type VectorClock map[string]int64

// ClockManager manages vector clocks for a node
type ClockManager struct {
	mu          sync.RWMutex
	nodeID      string
	allNodeIDs  []string
	clock       VectorClock // Local clock
	nodeIDIndex map[string]int // Index of each node ID for fast lookup
}

// NewClockManager creates a new vector clock manager
func NewClockManager(nodeID string, allNodeIDs []string) *ClockManager {
	// Initialize clock with all nodes at 0
	clock := make(VectorClock)
	nodeIDIndex := make(map[string]int)
	
	for i, id := range allNodeIDs {
		clock[id] = 0
		nodeIDIndex[id] = i
	}
	
	return &ClockManager{
		nodeID:      nodeID,
		allNodeIDs:  allNodeIDs,
		clock:       clock,
		nodeIDIndex: nodeIDIndex,
	}
}

// Tick increments the local node's clock (happens on local event)
func (cm *ClockManager) Tick() VectorClock {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.clock[cm.nodeID]++
	return cm.Copy()
}

// Update updates the clock based on a received vector clock (happens on receive)
// Takes the maximum of each component and increments local clock
func (cm *ClockManager) Update(received VectorClock) VectorClock {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	// Update each component to max of local and received
	for nodeID := range cm.clock {
		local := cm.clock[nodeID]
		receivedVal := received[nodeID]
		if receivedVal > local {
			cm.clock[nodeID] = receivedVal
		}
	}
	
	// Increment local clock (receive event)
	cm.clock[cm.nodeID]++
	
	return cm.Copy()
}

// Get returns a copy of the current clock
func (cm *ClockManager) Get() VectorClock {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.Copy()
}

// Copy creates a deep copy of the clock
func (cm *ClockManager) Copy() VectorClock {
	copy := make(VectorClock)
	for k, v := range cm.clock {
		copy[k] = v
	}
	return copy
}

// Compare compares two vector clocks
// Returns: -1 if v1 < v2, 0 if concurrent, 1 if v1 > v2
func Compare(v1, v2 VectorClock) int {
	if v1 == nil || v2 == nil {
		return 0
	}
	
	allNodes := make(map[string]bool)
	for nodeID := range v1 {
		allNodes[nodeID] = true
	}
	for nodeID := range v2 {
		allNodes[nodeID] = true
	}
	
	v1Less := false
	v2Less := false
	
	for nodeID := range allNodes {
		val1 := v1[nodeID]
		val2 := v2[nodeID]
		
		if val1 < val2 {
			v1Less = true
		} else if val1 > val2 {
			v2Less = true
		}
	}
	
	if v1Less && !v2Less {
		return -1 // v1 < v2
	} else if v2Less && !v1Less {
		return 1 // v1 > v2
	}
	return 0 // Concurrent or equal
}

// HappensBefore returns true if v1 happens before v2
func HappensBefore(v1, v2 VectorClock) bool {
	return Compare(v1, v2) < 0
}

// HappensAfter returns true if v1 happens after v2
func HappensAfter(v1, v2 VectorClock) bool {
	return Compare(v1, v2) > 0
}

// IsConcurrent returns true if v1 and v2 are concurrent
func IsConcurrent(v1, v2 VectorClock) bool {
	return Compare(v1, v2) == 0
}

// Merge merges two vector clocks by taking the maximum of each component
func Merge(v1, v2 VectorClock) VectorClock {
	merged := make(VectorClock)
	
	allNodes := make(map[string]bool)
	for nodeID := range v1 {
		allNodes[nodeID] = true
	}
	for nodeID := range v2 {
		allNodes[nodeID] = true
	}
	
	for nodeID := range allNodes {
		val1 := v1[nodeID]
		val2 := v2[nodeID]
		if val1 > val2 {
			merged[nodeID] = val1
		} else {
			merged[nodeID] = val2
		}
	}
	
	return merged
}

// String returns a string representation of the vector clock
func (vc VectorClock) String() string {
	if vc == nil {
		return "{}"
	}
	
	// Sort node IDs for consistent output
	nodeIDs := make([]string, 0, len(vc))
	for nodeID := range vc {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs)
	
	result := "{"
	for i, nodeID := range nodeIDs {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s:%d", nodeID, vc[nodeID])
	}
	result += "}"
	return result
}



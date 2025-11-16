package leaderfollower

import (
	"sync"
)

// NodeRole represents the role of a node in the cluster
type NodeRole string

const (
	RoleLeader   NodeRole = "leader"
	RoleFollower NodeRole = "follower"
)

// Config holds the configuration for the Leader-Follower cluster
type Config struct {
	mu            sync.RWMutex
	NodeID        string   // Unique identifier for this node
	Role          NodeRole // leader or follower
	MyAddr        string   // Address of this node
	LeaderAddr    string   // Address of the leader node
	FollowerAddrs []string // Addresses of all follower nodes
	AllNodeAddrs  []string // All node addresses (leader + followers)
	N             int      // Total number of nodes (default: 5)
	R             int      // Read quorum size
	W             int      // Write quorum size
}

// NewConfig creates a new configuration
func NewConfig(nodeID string, role NodeRole, myAddr string, leaderAddr string, followerAddrs []string) *Config {
	allAddrs := append([]string{leaderAddr}, followerAddrs...)
	return &Config{
		NodeID:        nodeID,
		Role:          role,
		MyAddr:        myAddr,
		LeaderAddr:    leaderAddr,
		FollowerAddrs: followerAddrs,
		AllNodeAddrs:  allAddrs,
		N:             len(allAddrs),
		R:             1, // Default
		W:             1, // Default
	}
}

// GetMyAddr returns the address of this node
func (c *Config) GetMyAddr() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MyAddr
}

// SetReplicationParams sets R and W values
func (c *Config) SetReplicationParams(r, w int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.R = r
	c.W = w
}

// GetReplicationParams returns current R and W values
func (c *Config) GetReplicationParams() (r, w int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.R, c.W
}

// IsLeader returns true if this node is the leader
func (c *Config) IsLeader() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Role == RoleLeader
}

// GetFollowerAddrs returns the addresses of all followers
func (c *Config) GetFollowerAddrs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]string{}, c.FollowerAddrs...)
}

// GetAllNodeAddrs returns addresses of all nodes
func (c *Config) GetAllNodeAddrs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]string{}, c.AllNodeAddrs...)
}


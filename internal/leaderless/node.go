package leaderless

import (
	"sync"
)

// Config holds the configuration for the Leaderless cluster
type Config struct {
	mu           sync.RWMutex
	NodeID       string   // Unique identifier for this node
	MyAddr       string   // Address of this node
	AllNodeAddrs []string // All node addresses in the cluster
	N            int      // Total number of nodes (default: 5)
	R            int      // Read quorum size (always 1 for leaderless)
	W            int      // Write quorum size (always N for leaderless)
}

// NewConfig creates a new leaderless configuration
func NewConfig(nodeID string, myAddr string, allNodeAddrs []string) *Config {
	return &Config{
		NodeID:       nodeID,
		MyAddr:       myAddr,
		AllNodeAddrs: allNodeAddrs,
		N:            len(allNodeAddrs),
		R:            1, // Always 1 for leaderless
		W:            len(allNodeAddrs), // Always N for leaderless
	}
}

// GetMyAddr returns the address of this node
func (c *Config) GetMyAddr() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MyAddr
}

// GetAllNodeAddrs returns addresses of all nodes
func (c *Config) GetAllNodeAddrs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]string{}, c.AllNodeAddrs...)
}

// GetOtherNodeAddrs returns addresses of all other nodes (excluding self)
func (c *Config) GetOtherNodeAddrs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	otherAddrs := make([]string, 0, len(c.AllNodeAddrs)-1)
	for _, addr := range c.AllNodeAddrs {
		if addr != c.MyAddr {
			otherAddrs = append(otherAddrs, addr)
		}
	}
	return otherAddrs
}

// GetN returns the total number of nodes
func (c *Config) GetN() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.N
}


package generator

import (
	"fmt"
	"math/rand"
	"time"
)

// RequestType represents the type of request
type RequestType string

const (
	RequestTypeWrite RequestType = "write"
	RequestTypeRead  RequestType = "read"
)

// Request represents a generated request
type Request struct {
	Type  RequestType
	Key   string
	Value string
}

// LocalInTimeKeyGenerator generates keys with local-in-time clustering
type LocalInTimeKeyGenerator struct {
	numKeys        int
	keyClusterSize int
	rand           *rand.Rand
	activeClusters []int // Currently active key clusters
	clusterTime    map[int]time.Time // Last access time for each cluster
}

// NewLocalInTimeKeyGenerator creates a new key generator
func NewLocalInTimeKeyGenerator(numKeys int, keyClusterSize int) *LocalInTimeKeyGenerator {
	return &LocalInTimeKeyGenerator{
		numKeys:        numKeys,
		keyClusterSize: keyClusterSize,
		rand:           rand.New(rand.NewSource(time.Now().UnixNano())),
		activeClusters: make([]int, 0),
		clusterTime:    make(map[int]time.Time),
	}
}

// GenerateKey generates a key with local-in-time clustering
// This ensures reads and writes to the same key are clustered in time
func (g *LocalInTimeKeyGenerator) GenerateKey() string {
	// With 80% probability, use an active cluster (recently accessed)
	// With 20% probability, start a new cluster
	useActiveCluster := g.rand.Float64() < 0.8 && len(g.activeClusters) > 0

	var clusterID int
	if useActiveCluster {
		// Pick a random active cluster
		clusterIdx := g.rand.Intn(len(g.activeClusters))
		clusterID = g.activeClusters[clusterIdx]
	} else {
		// Start a new cluster
		clusterID = g.rand.Intn(g.numKeys / g.keyClusterSize)
		g.activeClusters = append(g.activeClusters, clusterID)
		g.clusterTime[clusterID] = time.Now()
	}

	// Generate a key within the cluster
	keyOffset := g.rand.Intn(g.keyClusterSize)
	keyID := clusterID*g.keyClusterSize + keyOffset

	// Update cluster access time
	g.clusterTime[clusterID] = time.Now()

	// Remove old clusters (not accessed in last 5 seconds)
	now := time.Now()
	newActiveClusters := make([]int, 0)
	for _, cid := range g.activeClusters {
		if now.Sub(g.clusterTime[cid]) < 5*time.Second {
			newActiveClusters = append(newActiveClusters, cid)
		}
	}
	g.activeClusters = newActiveClusters

	return fmt.Sprintf("key_%d", keyID)
}

// RequestGenerator generates requests with specified read-write ratio
type RequestGenerator struct {
	keyGen     *LocalInTimeKeyGenerator
	writeRatio float64
	readRatio  float64
	rand       *rand.Rand
	counter    int64 // For generating unique values
}

// NewRequestGenerator creates a new request generator
func NewRequestGenerator(keyGen *LocalInTimeKeyGenerator, writeRatio float64, readRatio float64) *RequestGenerator {
	return &RequestGenerator{
		keyGen:     keyGen,
		writeRatio: writeRatio,
		readRatio:  readRatio,
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
		counter:    0,
	}
}

// Generate generates a new request
func (g *RequestGenerator) Generate() Request {
	key := g.keyGen.GenerateKey()

	// Determine request type based on ratios
	r := g.rand.Float64()
	if r < g.writeRatio {
		// Write request
		g.counter++
		return Request{
			Type:  RequestTypeWrite,
			Key:   key,
			Value: fmt.Sprintf("value_%d_%d", time.Now().UnixNano(), g.counter),
		}
	} else {
		// Read request
		return Request{
			Type: RequestTypeRead,
			Key:  key,
		}
	}
}






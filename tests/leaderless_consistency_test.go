package tests

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// TestLeaderlessConsistency tests consistency for Leaderless database
func TestLeaderlessConsistency(t *testing.T) {
	client := NewConsistencyTestClient()
	
	allNodeAddrs := []string{
		"localhost:8080",
		"localhost:8081",
		"localhost:8082",
		"localhost:8083",
		"localhost:8084",
	}

	// Test 1: Write to random node, read from other nodes within update window (should show inconsistency)
	t.Run("WriteToRandomNode_ReadFromOthersWithinWindow_ShowsInconsistency", func(t *testing.T) {
		// Pick a random node to write to (becomes Write Coordinator)
		rand.Seed(time.Now().UnixNano())
		coordinatorIdx := rand.Intn(len(allNodeAddrs))
		coordinatorAddr := allNodeAddrs[coordinatorIdx]

		key := fmt.Sprintf("test_inconsistency_%d", time.Now().UnixNano())
		value := "value1"

		// Write to random node (it becomes Write Coordinator)
		writeResp, err := client.Write(coordinatorAddr, key, value)
		if err != nil {
			t.Fatalf("Failed to write to coordinator: %v", err)
		}

		// Immediately read from other nodes (within update window)
		inconsistentCount := 0
		for i, nodeAddr := range allNodeAddrs {
			if i == coordinatorIdx {
				continue // Skip coordinator
			}

			// Read immediately (should show inconsistency)
			readResp, err := client.Read(nodeAddr, key)
			if err != nil {
				// Key not found is inconsistency
				inconsistentCount++
				t.Logf("Node %d: key not found (inconsistent)", i)
			} else if readResp.Value != value || readResp.Version != writeResp.Version {
				// Different value or version is inconsistency
				inconsistentCount++
				t.Logf("Node %d: inconsistent value (expected %s v%d, got %s v%d)", 
					i, value, writeResp.Version, readResp.Value, readResp.Version)
			}
		}

		// Should see inconsistency during replication window
		if inconsistentCount > 0 {
			t.Logf("✓ Detected inconsistency in %d/%d nodes (expected during replication window)", 
				inconsistentCount, len(allNodeAddrs)-1)
		} else {
			t.Logf("No inconsistency detected (replication may have completed very quickly)")
		}
	})

	// Test 2: After Coordinator acknowledges write, read from Coordinator (should be consistent)
	t.Run("AfterCoordinatorAck_ReadFromCoordinator_Consistent", func(t *testing.T) {
		rand.Seed(time.Now().UnixNano())
		coordinatorIdx := rand.Intn(len(allNodeAddrs))
		coordinatorAddr := allNodeAddrs[coordinatorIdx]

		key := fmt.Sprintf("test_coordinator_%d", time.Now().UnixNano())
		value := "value2"

		// Write to random node (becomes Write Coordinator)
		writeResp, err := client.Write(coordinatorAddr, key, value)
		if err != nil {
			t.Fatalf("Failed to write to coordinator: %v", err)
		}

		// After Coordinator acknowledges (write completed), read from Coordinator
		// Coordinator should have the value since it wrote locally first
		readResp, err := client.Read(coordinatorAddr, key)
		if err != nil {
			t.Fatalf("Failed to read from coordinator: %v", err)
		}

		// Should be consistent
		if readResp.Value != value {
			t.Errorf("Inconsistent read from coordinator: expected %s, got %s", value, readResp.Value)
		}
		if readResp.Version != writeResp.Version {
			t.Errorf("Version mismatch: expected %d, got %d", writeResp.Version, readResp.Version)
		}
	})

	// Test 3: After Coordinator acknowledges write, read from another node (should be consistent)
	t.Run("AfterCoordinatorAck_ReadFromOtherNode_Consistent", func(t *testing.T) {
		rand.Seed(time.Now().UnixNano())
		coordinatorIdx := rand.Intn(len(allNodeAddrs))
		coordinatorAddr := allNodeAddrs[coordinatorIdx]

		// Pick a different node to read from
		readerIdx := (coordinatorIdx + 1) % len(allNodeAddrs)
		readerAddr := allNodeAddrs[readerIdx]

		key := fmt.Sprintf("test_other_node_%d", time.Now().UnixNano())
		value := "value3"

		// Write to random node (becomes Write Coordinator)
		writeResp, err := client.Write(coordinatorAddr, key, value)
		if err != nil {
			t.Fatalf("Failed to write to coordinator: %v", err)
		}

		// Wait for replication to complete (Coordinator has acknowledged)
		time.Sleep(4 * time.Second)

		// Read from another node
		readResp, err := client.Read(readerAddr, key)
		if err != nil {
			t.Fatalf("Failed to read from other node: %v", err)
		}

		// Should be consistent after replication
		if readResp.Value != value {
			t.Errorf("Inconsistent read from other node: expected %s, got %s", value, readResp.Value)
		}
		if readResp.Version != writeResp.Version {
			t.Errorf("Version mismatch: expected %d, got %d", writeResp.Version, readResp.Version)
		}
	})
}

// TestLeaderlessConsistencyHighLoad tests consistency under high load
func TestLeaderlessConsistencyHighLoad(t *testing.T) {
	client := NewConsistencyTestClient()
	allNodeAddrs := []string{
		"localhost:8080",
		"localhost:8081",
		"localhost:8082",
		"localhost:8083",
		"localhost:8084",
	}

	rand.Seed(time.Now().UnixNano())
	numWrites := 10
	inconsistencies := 0

	for i := 0; i < numWrites; i++ {
		// Write to random node
		coordinatorIdx := rand.Intn(len(allNodeAddrs))
		coordinatorAddr := allNodeAddrs[coordinatorIdx]

		key := fmt.Sprintf("load_test_%d", i)
		value := fmt.Sprintf("value_%d", i)

		writeResp, err := client.Write(coordinatorAddr, key, value)
		if err != nil {
			t.Logf("Write %d failed: %v", i, err)
			continue
		}

		// Immediately check other nodes
		for j, nodeAddr := range allNodeAddrs {
			if j == coordinatorIdx {
				continue
			}
			readResp, err := client.Read(nodeAddr, key)
			if err != nil || readResp.Value != value || readResp.Version != writeResp.Version {
				inconsistencies++
			}
		}
	}

	t.Logf("Detected %d inconsistencies out of %d checks (expected at high load)", 
		inconsistencies, numWrites*(len(allNodeAddrs)-1))
}






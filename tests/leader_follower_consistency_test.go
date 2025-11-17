package tests

import (
	"fmt"
	"testing"
	"time"
)

// TestLeaderFollowerConsistency tests consistency for Leader-Follower database
func TestLeaderFollowerConsistency(t *testing.T) {
	client := NewConsistencyTestClient()
	
	leaderAddr := "localhost:8080"
	followerAddrs := []string{
		"localhost:8081",
		"localhost:8082",
		"localhost:8083",
		"localhost:8084",
	}

	// Test 1: Write to Leader, read from Leader (should be consistent)
	t.Run("WriteToLeader_ReadFromLeader_Consistent", func(t *testing.T) {
		key := fmt.Sprintf("test_leader_leader_%d", time.Now().UnixNano())
		value := "value1"

		// Write to Leader
		writeResp, err := client.Write(leaderAddr, key, value)
		if err != nil {
			t.Fatalf("Failed to write to leader: %v", err)
		}
		if writeResp.Value != value {
			t.Fatalf("Write response value mismatch: expected %s, got %s", value, writeResp.Value)
		}

		// Read from Leader after write acknowledgment
		readResp, err := client.Read(leaderAddr, key)
		if err != nil {
			t.Fatalf("Failed to read from leader: %v", err)
		}

		// Should be consistent
		if readResp.Value != value {
			t.Errorf("Inconsistent read from leader: expected %s, got %s", value, readResp.Value)
		}
		if readResp.Version != writeResp.Version {
			t.Errorf("Version mismatch: expected %d, got %d", writeResp.Version, readResp.Version)
		}
	})

	// Test 2: Write to Leader, read from Follower (should be consistent after ack)
	t.Run("WriteToLeader_ReadFromFollower_Consistent", func(t *testing.T) {
		key := fmt.Sprintf("test_leader_follower_%d", time.Now().UnixNano())
		value := "value2"

		// Write to Leader
		writeResp, err := client.Write(leaderAddr, key, value)
		if err != nil {
			t.Fatalf("Failed to write to leader: %v", err)
		}

		// Wait for replication to complete (Leader has acknowledged)
		time.Sleep(3 * time.Second)

		// Read from Follower
		readResp, err := client.Read(followerAddrs[0], key)
		if err != nil {
			t.Fatalf("Failed to read from follower: %v", err)
		}

		// Should be consistent after replication
		if readResp.Value != value {
			t.Errorf("Inconsistent read from follower: expected %s, got %s", value, readResp.Value)
		}
		if readResp.Version != writeResp.Version {
			t.Errorf("Version mismatch: expected %d, got %d", writeResp.Version, readResp.Version)
		}
	})

	// Test 3: Write to Leader, immediately local_read from Followers (may show inconsistency)
	t.Run("WriteToLeader_ImmediateLocalReadFromFollowers_MayShowInconsistency", func(t *testing.T) {
		key := fmt.Sprintf("test_inconsistency_%d", time.Now().UnixNano())
		value := "value3"

		// Write to Leader
		writeResp, err := client.Write(leaderAddr, key, value)
		if err != nil {
			t.Fatalf("Failed to write to leader: %v", err)
		}

		// Immediately read locally from followers (within update window)
		inconsistentCount := 0
		for i, followerAddr := range followerAddrs {
			// Small delay to catch some followers mid-replication
			time.Sleep(50 * time.Millisecond * time.Duration(i))
			
			localResp, err := client.LocalRead(followerAddr, key)
			if err != nil {
				// Key not found is also inconsistency
				inconsistentCount++
				t.Logf("Follower %d: key not found (inconsistent)", i+1)
			} else if localResp.Value != value || localResp.Version != writeResp.Version {
				// Different value or version is inconsistency
				inconsistentCount++
				t.Logf("Follower %d: inconsistent value (expected %s v%d, got %s v%d)", 
					i+1, value, writeResp.Version, localResp.Value, localResp.Version)
			}
		}

		// At high load, we should see inconsistency. For now, just log it.
		if inconsistentCount > 0 {
			t.Logf("Detected inconsistency in %d/%d followers (expected during replication window)", 
				inconsistentCount, len(followerAddrs))
		} else {
			t.Logf("No inconsistency detected (replication may have completed quickly)")
		}

		// After waiting, all should be consistent
		time.Sleep(3 * time.Second)
		for i, followerAddr := range followerAddrs {
			localResp, err := client.LocalRead(followerAddr, key)
			if err != nil {
				t.Errorf("Follower %d: key not found after replication", i+1)
			} else if localResp.Value != value {
				t.Errorf("Follower %d: still inconsistent after replication: expected %s, got %s", 
					i+1, value, localResp.Value)
			}
		}
	})
}

// TestLeaderFollowerConsistencyHighLoad tests consistency under high load
func TestLeaderFollowerConsistencyHighLoad(t *testing.T) {
	client := NewConsistencyTestClient()
	leaderAddr := "localhost:8080"
	followerAddrs := []string{"localhost:8081", "localhost:8082", "localhost:8083", "localhost:8084"}

	// Send multiple writes quickly
	numWrites := 10
	inconsistencies := 0

	for i := 0; i < numWrites; i++ {
		key := fmt.Sprintf("load_test_%d", i)
		value := fmt.Sprintf("value_%d", i)

		// Write to Leader
		writeResp, err := client.Write(leaderAddr, key, value)
		if err != nil {
			t.Logf("Write %d failed: %v", i, err)
			continue
		}

		// Immediately check followers
		for _, followerAddr := range followerAddrs {
			localResp, err := client.LocalRead(followerAddr, key)
			if err != nil || localResp.Value != value || localResp.Version != writeResp.Version {
				inconsistencies++
			}
		}
	}

	t.Logf("Detected %d inconsistencies out of %d checks (expected at high load)", 
		inconsistencies, numWrites*len(followerAddrs))
}






package leaderless

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ReplicationClient handles communication between nodes
type ReplicationClient struct {
	httpClient *http.Client
}

// NewReplicationClient creates a new replication client
func NewReplicationClient() *ReplicationClient {
	return &ReplicationClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ReplicateWriteRequest represents a write replication request
type ReplicateWriteRequest struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Version int64  `json:"version"`
}

// ReplicateWriteResponse represents a write replication response
type ReplicateWriteResponse struct {
	Success bool   `json:"success"`
	Version int64  `json:"version"`
	Error   string `json:"error,omitempty"`
}

// ReplicateWrite sends a write request to another node
// Returns the response and any error
func (c *ReplicationClient) ReplicateWrite(addr string, key string, value string, version int64, addDelay bool) (*ReplicateWriteResponse, error) {
	reqBody := ReplicateWriteRequest{
		Key:     key,
		Value:   value,
		Version: version,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("http://%s/internal/replicate_write", addr)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Write Coordinator sleeps 200ms after each message to another node
	if addDelay {
		time.Sleep(200 * time.Millisecond)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &ReplicateWriteResponse{
			Success: false,
			Error:   string(body),
		}, nil
	}

	var response ReplicateWriteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}


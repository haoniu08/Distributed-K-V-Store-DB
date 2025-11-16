package leaderfollower

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

// ReadRequest represents a read request from another node
type ReadRequest struct {
	Key string `json:"key"`
}

// ReadResponse represents a read response
type ReadResponse struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Version int64  `json:"version"`
	Exists  bool   `json:"exists"`
}

// ReplicateWrite sends a write request to a follower node
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

	// Leader sleeps 200ms after each message to a Follower
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

// ReadFromNode reads a value from another node
func (c *ReplicationClient) ReadFromNode(addr string, key string, addDelay bool) (*ReadResponse, error) {
	url := fmt.Sprintf("http://%s/internal/read?key=%s", addr, key)
	
	// Follower sleeps 50ms when receiving read request from Leader
	if addDelay {
		time.Sleep(50 * time.Millisecond)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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

	var response ReadResponse
	if resp.StatusCode == http.StatusNotFound {
		response.Exists = false
		return &response, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	response.Exists = true
	return &response, nil
}



package models

// KeyValue represents a key-value pair with versioning
type KeyValue struct {
	Key     string
	Value   string
	Version int64 // Logical version number for replication
}

// SetRequest represents a set operation request
type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetResponse represents a get operation response
type GetResponse struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Version int64  `json:"version"`
}

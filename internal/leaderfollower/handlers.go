package leaderfollower

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourusername/distributed-kv-store/internal/kvstore"
)

// Handler provides HTTP handlers for Leader-Follower database
type Handler struct {
	store    *kvstore.Store
	config   *Config
	replicator *ReplicationManager
}

// NewHandler creates a new Leader-Follower handler
func NewHandler(store *kvstore.Store, config *Config) *Handler {
	replicator := NewReplicationManager(store, config)
	return &Handler{
		store:      store,
		config:     config,
		replicator: replicator,
	}
}

// SetHandler handles write requests (only from Leader)
func (h *Handler) SetHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}

	// Only Leader can accept writes
	if !h.config.IsLeader() {
		http.Error(w, "only leader accepts write requests", http.StatusForbidden)
		return
	}

	// Perform write with replication
	result, err := h.replicator.Write(req.Key, req.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":     req.Key,
		"value":   req.Value,
		"version": result.Version,
		"status":  "created",
	})
}

// GetHandler handles read requests (can go to any node)
func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key parameter is required", http.StatusBadRequest)
		return
	}

	// Perform read with replication strategy
	kv, err := h.replicator.Read(key)
	if err != nil {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":     kv.Key,
		"value":   kv.Value,
		"version": kv.Version,
	})
}

// LocalReadHandler handles local reads (for testing)
func (h *Handler) LocalReadHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key parameter is required", http.StatusBadRequest)
		return
	}

	kv, exists := h.store.LocalRead(key)
	if !exists {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":     kv.Key,
		"value":   kv.Value,
		"version": kv.Version,
	})
}

// ReplicateWriteHandler handles internal write replication requests from Leader
func (h *Handler) ReplicateWriteHandler(w http.ResponseWriter, r *http.Request) {
	var req ReplicateWriteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Follower sleeps 100ms when receiving update before responding
	time.Sleep(100 * time.Millisecond)

	// Set the value with the provided version
	if err := h.store.SetWithVersion(req.Key, req.Value, req.Version); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ReplicateWriteResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ReplicateWriteResponse{
		Success: true,
		Version: req.Version,
	})
}

// InternalReadHandler handles internal read requests from other nodes
func (h *Handler) InternalReadHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key parameter is required", http.StatusBadRequest)
		return
	}

	// Follower sleeps 50ms when receiving read request
	if !h.config.IsLeader() {
		time.Sleep(50 * time.Millisecond)
	}

	kv, exists := h.store.Get(key)
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ReadResponse{
			Exists: false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ReadResponse{
		Key:     kv.Key,
		Value:   kv.Value,
		Version: kv.Version,
		Exists:  true,
	})
}

// ConfigHandler handles configuration requests
func (h *Handler) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Get current configuration
		readR, writeW := h.config.GetReplicationParams()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"node_id": h.config.NodeID,
			"role":    h.config.Role,
			"n":       h.config.N,
			"r":       readR,
			"w":       writeW,
		})
		return
	}

	if r.Method == "POST" {
		// Set R and W values
		var req struct {
			R int `json:"r"`
			W int `json:"w"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.R < 1 || req.R > h.config.N || req.W < 1 || req.W > h.config.N {
			http.Error(w, "R and W must be between 1 and N", http.StatusBadRequest)
			return
		}

		h.config.SetReplicationParams(req.R, req.W)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "configuration updated",
			"r":      req.R,
			"w":      req.W,
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HealthHandler provides a health check endpoint
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"role":   h.config.Role,
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}


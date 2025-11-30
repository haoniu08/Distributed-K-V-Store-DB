package leaderless

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourusername/distributed-kv-store/internal/kvstore"
	"github.com/yourusername/distributed-kv-store/internal/vectorclock"
)

// Handler provides HTTP handlers for Leaderless database
type Handler struct {
	store      *kvstore.Store
	config     *Config
	replicator *ReplicationManager
}

// NewHandler creates a new Leaderless handler
func NewHandler(store *kvstore.Store, config *Config) *Handler {
	replicator := NewReplicationManager(store, config)
	return &Handler{
		store:      store,
		config:     config,
		replicator: replicator,
	}
}

// SetHandler handles write requests (any node can receive writes)
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

	// This node becomes the Write Coordinator
	// It must coordinate writes to all other nodes
	result, err := h.replicator.WriteWithCoordination(req.Key, req.Value)
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

// GetHandler handles read requests (any node can receive reads)
// Returns local value immediately (R=1)
func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key parameter is required", http.StatusBadRequest)
		return
	}

	// Read local value immediately (no coordination)
	kv, err := h.replicator.ReadLocal(key)
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

// ReplicateWriteHandler handles internal write replication requests from Write Coordinator
func (h *Handler) ReplicateWriteHandler(w http.ResponseWriter, r *http.Request) {
	var req ReplicateWriteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Node sleeps 100ms when receiving update before responding
	time.Sleep(100 * time.Millisecond)

	// Convert map to vector clock if present
	var vc vectorclock.VectorClock
	if req.VectorClock != nil && len(req.VectorClock) > 0 {
		vc = make(vectorclock.VectorClock)
		for k, v := range req.VectorClock {
			vc[k] = v
		}
	}

	// Set the value with the provided version and vector clock
	var err error
	if vc != nil {
		err = h.store.SetWithVersionAndClock(req.Key, req.Value, req.Version, vc)
		// Update local clock AFTER accepting the write (receive event)
		if err == nil && h.replicator != nil {
			h.replicator.UpdateClock(vc)
		}
	} else {
		err = h.store.SetWithVersion(req.Key, req.Value, req.Version)
	}
	
	if err != nil {
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

// HealthHandler provides a health check endpoint
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"mode":   "leaderless",
		"node_id": h.config.NodeID,
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}


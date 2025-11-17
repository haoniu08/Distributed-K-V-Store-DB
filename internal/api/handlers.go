package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourusername/distributed-kv-store/internal/kvstore"
)

// Handler wraps the KV store and provides HTTP handlers
type Handler struct {
	store *kvstore.Store
}

// NewHandler creates a new API handler
func NewHandler(store *kvstore.Store) *Handler {
	return &Handler{
		store: store,
	}
}

// SetHandler handles PUT/POST requests to set a key-value pair
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

	version, err := h.store.Set(req.Key, req.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":     req.Key,
		"value":   req.Value,
		"version": version,
		"status":  "created",
	})
}

// GetHandler handles GET requests to retrieve a value
func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key parameter is required", http.StatusBadRequest)
		return
	}

	kv, exists := h.store.Get(key)
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

// LocalReadHandler handles GET requests for local reads (testing only)
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

// HealthHandler provides a health check endpoint
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}


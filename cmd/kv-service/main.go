package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/yourusername/distributed-kv-store/internal/api"
	"github.com/yourusername/distributed-kv-store/internal/kvstore"
)

func main() {
	// Create KV store
	store := kvstore.NewStore()

	// Create API handler
	handler := api.NewHandler(store)

	// Setup router
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/set", handler.SetHandler).Methods("POST", "PUT")
	r.HandleFunc("/get", handler.GetHandler).Methods("GET")
	r.HandleFunc("/local_read", handler.LocalReadHandler).Methods("GET") // For testing
	r.HandleFunc("/health", handler.HealthHandler).Methods("GET")

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting KV service on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

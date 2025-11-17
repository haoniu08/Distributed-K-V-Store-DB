package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/yourusername/distributed-kv-store/internal/kvstore"
	"github.com/yourusername/distributed-kv-store/internal/leaderless"
)

func main() {
	// Parse command line flags
	nodeID := flag.String("node-id", "", "Unique identifier for this node (required)")
	allNodeAddrsStr := flag.String("all-node-addrs", "", "Comma-separated list of all node addresses (required)")
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	// Validate required flags
	if *nodeID == "" {
		log.Fatal("--node-id is required")
	}
	if *allNodeAddrsStr == "" {
		log.Fatal("--all-node-addrs is required")
	}

	// Parse all node addresses
	allNodeAddrs := strings.Split(*allNodeAddrsStr, ",")
	// Trim whitespace
	for i, addr := range allNodeAddrs {
		allNodeAddrs[i] = strings.TrimSpace(addr)
	}

	// Determine this node's address
	myAddr := "localhost:" + *port
	if envAddr := os.Getenv("MY_ADDR"); envAddr != "" {
		myAddr = envAddr
	}

	// Verify this node's address is in the list
	found := false
	for _, addr := range allNodeAddrs {
		if addr == myAddr || addr == "localhost:"+*port {
			found = true
			break
		}
	}
	if !found {
		log.Printf("Warning: node address %s not found in all-node-addrs list", myAddr)
	}

	// Create node configuration
	config := leaderless.NewConfig(*nodeID, myAddr, allNodeAddrs)

	// Create KV store
	store := kvstore.NewStore()

	// Create handler
	handler := leaderless.NewHandler(store, config)

	// Setup router
	r := mux.NewRouter()

	// External API routes
	r.HandleFunc("/set", handler.SetHandler).Methods("POST", "PUT")
	r.HandleFunc("/get", handler.GetHandler).Methods("GET")
	r.HandleFunc("/local_read", handler.LocalReadHandler).Methods("GET") // For testing
	r.HandleFunc("/health", handler.HealthHandler).Methods("GET")

	// Internal API routes (for replication)
	r.HandleFunc("/internal/replicate_write", handler.ReplicateWriteHandler).Methods("POST")

	// Get port from environment or flag
	listenPort := os.Getenv("PORT")
	if listenPort == "" {
		listenPort = *port
	}

	log.Printf("Starting Leaderless node: %s on port %s", *nodeID, listenPort)
	log.Printf("All node addresses: %v", allNodeAddrs)
	log.Printf("This node address: %s", myAddr)
	log.Printf("Configuration: W=%d (N), R=1", config.GetN())
	log.Fatal(http.ListenAndServe(":"+listenPort, r))
}


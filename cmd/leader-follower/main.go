package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/yourusername/distributed-kv-store/internal/kvstore"
	"github.com/yourusername/distributed-kv-store/internal/leaderfollower"
)

func main() {
	// Parse command line flags
	nodeID := flag.String("node-id", "", "Unique identifier for this node (required)")
	role := flag.String("role", "", "Node role: 'leader' or 'follower' (required)")
	leaderAddr := flag.String("leader-addr", "", "Address of the leader node (e.g., localhost:8080)")
	followerAddrsStr := flag.String("follower-addrs", "", "Comma-separated list of follower addresses (e.g., localhost:8081,localhost:8082)")
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	// Validate required flags
	if *nodeID == "" {
		log.Fatal("--node-id is required")
	}
	if *role != "leader" && *role != "follower" {
		log.Fatal("--role must be 'leader' or 'follower'")
	}
	if *leaderAddr == "" {
		log.Fatal("--leader-addr is required")
	}

	// Parse follower addresses
	var followerAddrs []string
	if *followerAddrsStr != "" {
		followerAddrs = strings.Split(*followerAddrsStr, ",")
		// Trim whitespace
		for i, addr := range followerAddrs {
			followerAddrs[i] = strings.TrimSpace(addr)
		}
	}

	// Determine this node's address
	myAddr := "localhost:" + *port
	if envAddr := os.Getenv("MY_ADDR"); envAddr != "" {
		myAddr = envAddr
	}

	// Create node configuration
	var config *leaderfollower.Config
	if *role == "leader" {
		config = leaderfollower.NewConfig(*nodeID, leaderfollower.RoleLeader, myAddr, *leaderAddr, followerAddrs)
	} else {
		config = leaderfollower.NewConfig(*nodeID, leaderfollower.RoleFollower, myAddr, *leaderAddr, followerAddrs)
	}

	// Set default replication parameters (can be changed via API)
	// Default to W=5, R=1 for initial setup
	config.SetReplicationParams(1, 5)

	// Create KV store
	store := kvstore.NewStore()

	// Create handler
	handler := leaderfollower.NewHandler(store, config)

	// Setup router
	r := mux.NewRouter()

	// External API routes
	r.HandleFunc("/set", handler.SetHandler).Methods("POST", "PUT")
	r.HandleFunc("/get", handler.GetHandler).Methods("GET")
	r.HandleFunc("/local_read", handler.LocalReadHandler).Methods("GET") // For testing
	r.HandleFunc("/health", handler.HealthHandler).Methods("GET")
	r.HandleFunc("/config", handler.ConfigHandler).Methods("GET", "POST")

	// Internal API routes (for replication)
	r.HandleFunc("/internal/replicate_write", handler.ReplicateWriteHandler).Methods("POST")
	r.HandleFunc("/internal/read", handler.InternalReadHandler).Methods("GET")

	// Get port from environment or flag
	listenPort := os.Getenv("PORT")
	if listenPort == "" {
		listenPort = *port
	}

	log.Printf("Starting Leader-Follower node: %s (role: %s) on port %s", *nodeID, *role, listenPort)
	log.Printf("Leader address: %s", *leaderAddr)
	if len(followerAddrs) > 0 {
		log.Printf("Follower addresses: %v", followerAddrs)
	}
	log.Fatal(http.ListenAndServe(":"+listenPort, r))
}


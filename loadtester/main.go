package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/yourusername/distributed-kv-store/loadtester/client"
	"github.com/yourusername/distributed-kv-store/loadtester/generator"
	"github.com/yourusername/distributed-kv-store/loadtester/stats"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "", "Path to configuration JSON file")
	outputDir := flag.String("output", "results", "Output directory for results")
	duration := flag.Duration("duration", 60*time.Second, "Test duration")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent workers")
	flag.Parse()

	if *configFile == "" {
		log.Fatal("--config is required")
	}

	// Load configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Initialize statistics collector
	statsCollector := stats.NewCollector()

	// Create HTTP client
	httpClient := client.NewLoadTestClient()

	// Create key generator (local-in-time)
	keyGen := generator.NewLocalInTimeKeyGenerator(config.NumKeys, config.KeyClusterSize)

	// Create request generator
	reqGen := generator.NewRequestGenerator(keyGen, config.WriteRatio, config.ReadRatio)

	// Track versions for stale read detection
	versionTracker := sync.Map{} // key -> latest known version

	// Start load test
	log.Printf("Starting load test:")
	log.Printf("  Configuration: %s", config.Name)
	log.Printf("  Duration: %v", *duration)
	log.Printf("  Concurrency: %d workers", *concurrency)
	log.Printf("  Write ratio: %.1f%%, Read ratio: %.1f%%", config.WriteRatio*100, config.ReadRatio*100)
	log.Printf("  Target: %s", config.TargetAddr)

	startTime := time.Now()
	endTime := startTime.Add(*duration)

	// Channel for requests
	requestChan := make(chan generator.Request, *concurrency*10)
	done := make(chan bool)

	// Start request generator
	go func() {
		defer close(requestChan)
		for time.Now().Before(endTime) {
			req := reqGen.Generate()
			requestChan <- req
			// Small delay to avoid overwhelming
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for req := range requestChan {
				if time.Now().After(endTime) {
					break
				}
				processRequest(workerID, req, config, httpClient, statsCollector, &versionTracker)
			}
		}(i)
	}

	// Wait for generator to finish
	<-done

	// Wait for all workers to finish
	wg.Wait()

	// Finalize statistics
	statsCollector.Finalize()

	// Export results
	log.Printf("Exporting results...")
	if err := exportResults(*outputDir, config, statsCollector); err != nil {
		log.Fatalf("Failed to export results: %v", err)
	}

	log.Printf("Load test complete!")
	log.Printf("Results saved to: %s", *outputDir)
}

func processRequest(
	workerID int,
	req generator.Request,
	config *Config,
	httpClient *client.LoadTestClient,
	statsCollector *stats.Collector,
	versionTracker *sync.Map,
) {
	startTime := time.Now()
	var err error
	var response *client.Response
	var isStale bool

	if req.Type == generator.RequestTypeWrite {
		// Write request
		response, err = httpClient.Write(config.TargetAddr, req.Key, req.Value)
		if err == nil {
			// Update version tracker
			versionTracker.Store(req.Key, response.Version)
		}
	} else {
		// Read request
		response, err = httpClient.Read(config.TargetAddr, req.Key)
		if err == nil {
			// Check for stale read
			if storedVersion, ok := versionTracker.Load(req.Key); ok {
				if response.Version < storedVersion.(int64) {
					isStale = true
				}
			}
			// Update version tracker
			versionTracker.Store(req.Key, response.Version)
		}
	}

	latency := time.Since(startTime)

	// Record statistics
	statsCollector.RecordRequest(stats.RequestRecord{
		Timestamp:    startTime,
		Type:         string(req.Type),
		Key:          req.Key,
		Latency:      latency,
		Success:      err == nil,
		IsStale:      isStale,
		Version:      getVersion(response),
		Error:        getError(err),
	})
}

func getVersion(resp *client.Response) int64 {
	if resp == nil {
		return 0
	}
	return resp.Version
}

func getError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func exportResults(outputDir string, config *Config, statsCollector *stats.Collector) error {
	// Export CSV
	csvFile, err := os.Create(fmt.Sprintf("%s/results.csv", outputDir))
	if err != nil {
		return err
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"timestamp", "type", "key", "latency_ms", "success", "is_stale", "version"})

	// Write data
	for _, record := range statsCollector.GetRecords() {
		writer.Write([]string{
			record.Timestamp.Format(time.RFC3339Nano),
			record.Type,
			record.Key,
			fmt.Sprintf("%.2f", float64(record.Latency.Nanoseconds())/1e6),
			fmt.Sprintf("%t", record.Success),
			fmt.Sprintf("%t", record.IsStale),
			fmt.Sprintf("%d", record.Version),
		})
	}

	// Export summary JSON
	summary := statsCollector.GetSummary()
	summary.Config = config.Name
	summaryFile, err := os.Create(fmt.Sprintf("%s/summary.json", outputDir))
	if err != nil {
		return err
	}
	defer summaryFile.Close()

	encoder := json.NewEncoder(summaryFile)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summary)
}

// Config represents load test configuration
type Config struct {
	Name         string  `json:"name"`
	TargetAddr   string  `json:"target_addr"`
	WriteRatio   float64 `json:"write_ratio"`   // 0.0 to 1.0
	ReadRatio    float64 `json:"read_ratio"`    // 0.0 to 1.0
	NumKeys      int     `json:"num_keys"`       // Total number of keys
	KeyClusterSize int   `json:"key_cluster_size"` // Keys per cluster for local-in-time
}


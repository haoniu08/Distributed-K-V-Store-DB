package stats

import (
	"sync"
	"time"
)

// RequestRecord represents a single request record
type RequestRecord struct {
	Timestamp time.Time
	Type      string // "write" or "read"
	Key       string
	Latency   time.Duration
	Success   bool
	IsStale   bool
	Version   int64
	Error     string
}

// Collector collects statistics during load testing
type Collector struct {
	mu      sync.RWMutex
	records []RequestRecord
}

// NewCollector creates a new statistics collector
func NewCollector() *Collector {
	return &Collector{
		records: make([]RequestRecord, 0),
	}
}

// RecordRequest records a request
func (c *Collector) RecordRequest(record RequestRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = append(c.records, record)
}

// GetRecords returns all records
func (c *Collector) GetRecords() []RequestRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]RequestRecord{}, c.records...)
}

// Summary represents test summary statistics
type Summary struct {
	Config              string    `json:"config"`
	TotalRequests       int       `json:"total_requests"`
	TotalWrites         int       `json:"total_writes"`
	TotalReads          int       `json:"total_reads"`
	SuccessfulRequests  int       `json:"successful_requests"`
	FailedRequests      int       `json:"failed_requests"`
	SuccessRate         float64   `json:"success_rate_percent"`
	ErrorRate           float64   `json:"error_rate_percent"`
	StaleReads          int       `json:"stale_reads"`
	StaleReadRate       float64   `json:"stale_read_rate_percent"`
	SuccessfulReads     int       `json:"successful_reads"`
	StaleReadRateOfSuccessful float64 `json:"stale_read_rate_of_successful_percent"`
	WriteLatency        LatencyStats `json:"write_latency"`
	ReadLatency         LatencyStats `json:"read_latency"`
	StartTime           time.Time `json:"start_time"`
	EndTime             time.Time `json:"end_time"`
	Duration            string    `json:"duration"`
}

// LatencyStats represents latency statistics
type LatencyStats struct {
	Min    float64 `json:"min_ms"`
	Max    float64 `json:"max_ms"`
	Mean   float64 `json:"mean_ms"`
	Median float64 `json:"median_ms"`
	P50    float64 `json:"p50_ms"`
	P95    float64 `json:"p95_ms"`
	P99    float64 `json:"p99_ms"`
	P999   float64 `json:"p999_ms"`
}

// Finalize finalizes statistics collection
func (c *Collector) Finalize() {
	// Statistics are computed on-demand in GetSummary()
}

// GetSummary computes and returns summary statistics
func (c *Collector) GetSummary() Summary {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.records) == 0 {
		return Summary{}
	}

	var startTime, endTime time.Time
	var writeLatencies []float64
	var readLatencies []float64
	totalRequests := len(c.records)
	totalWrites := 0
	totalReads := 0
	successfulRequests := 0
	failedRequests := 0
	staleReads := 0
	successfulReads := 0

	for _, record := range c.records {
		if startTime.IsZero() || record.Timestamp.Before(startTime) {
			startTime = record.Timestamp
		}
		if endTime.IsZero() || record.Timestamp.After(endTime) {
			endTime = record.Timestamp
		}

		if record.Success {
			successfulRequests++
		} else {
			failedRequests++
		}

		latencyMs := float64(record.Latency.Nanoseconds()) / 1e6

		if record.Type == "write" {
			totalWrites++
			if record.Success {
				writeLatencies = append(writeLatencies, latencyMs)
			}
		} else if record.Type == "read" {
			totalReads++
			if record.Success {
				successfulReads++
				readLatencies = append(readLatencies, latencyMs)
			}
			if record.IsStale {
				staleReads++
			}
		}
	}

	// Calculate rates
	successRate := 0.0
	errorRate := 0.0
	if totalRequests > 0 {
		successRate = float64(successfulRequests) / float64(totalRequests) * 100.0
		errorRate = float64(failedRequests) / float64(totalRequests) * 100.0
	}

	staleReadRate := 0.0
	staleReadRateOfSuccessful := 0.0
	if totalReads > 0 {
		staleReadRate = float64(staleReads) / float64(totalReads) * 100.0
	}
	if successfulReads > 0 {
		staleReadRateOfSuccessful = float64(staleReads) / float64(successfulReads) * 100.0
	}

	return Summary{
		Config:                      "",
		TotalRequests:               totalRequests,
		TotalWrites:                 totalWrites,
		TotalReads:                  totalReads,
		SuccessfulRequests:          successfulRequests,
		FailedRequests:              failedRequests,
		SuccessRate:                 successRate,
		ErrorRate:                   errorRate,
		StaleReads:                  staleReads,
		StaleReadRate:               staleReadRate,
		SuccessfulReads:             successfulReads,
		StaleReadRateOfSuccessful:   staleReadRateOfSuccessful,
		WriteLatency:                computeLatencyStats(writeLatencies),
		ReadLatency:                 computeLatencyStats(readLatencies),
		StartTime:                   startTime,
		EndTime:                     endTime,
		Duration:                    endTime.Sub(startTime).String(),
	}
}

func computeLatencyStats(latencies []float64) LatencyStats {
	if len(latencies) == 0 {
		return LatencyStats{}
	}

	// Sort latencies
	sorted := make([]float64, len(latencies))
	copy(sorted, latencies)
	
	// Simple bubble sort (for small datasets)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	var sum float64
	for _, l := range sorted {
		sum += l
	}

	mean := sum / float64(len(sorted))
	median := sorted[len(sorted)/2]

	return LatencyStats{
		Min:    sorted[0],
		Max:    sorted[len(sorted)-1],
		Mean:   mean,
		Median: median,
		P50:    sorted[int(float64(len(sorted))*0.50)],
		P95:    sorted[int(float64(len(sorted))*0.95)],
		P99:    sorted[int(float64(len(sorted))*0.99)],
		P999:   sorted[int(float64(len(sorted))*0.999)],
	}
}






package ingestion

import (
	"sync"
	"time"
)

type ConnectorStats struct {
	mu sync.RWMutex

	Name               string
	LastStartedAt      time.Time
	LastFinishedAt     time.Time
	LastSuccessAt      time.Time
	LastFailureAt      time.Time
	TotalRuns          int
	SuccessfulRuns     int
	FailedRuns         int
	RecordsFetched     int
	RecordsNormalized  int
	RecordsInserted    int
	RecordsDuplicate   int
	RecordsRejected    int
	ConsecutiveFailures int
	LastDuration       time.Duration
	LastError          string
	
	// Internal state
	isRunning bool
}

func NewConnectorStats(name string) *ConnectorStats {
	return &ConnectorStats{Name: name}
}

// TryStart attempts to mark the connector as running. Returns true if successful, false if it is already running.
func (s *ConnectorStats) TryStart() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.isRunning {
		return false
	}
	s.isRunning = true
	s.LastStartedAt = time.Now()
	s.TotalRuns++
	return true
}

func (s *ConnectorStats) MarkSuccess(duration time.Duration, fetched, normalized, inserted, duplicate, rejected int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isRunning = false
	s.LastFinishedAt = time.Now()
	s.LastSuccessAt = s.LastFinishedAt
	s.LastDuration = duration
	s.LastError = ""
	s.ConsecutiveFailures = 0
	s.SuccessfulRuns++

	s.RecordsFetched += fetched
	s.RecordsNormalized += normalized
	s.RecordsInserted += inserted
	s.RecordsDuplicate += duplicate
	s.RecordsRejected += rejected
}

func (s *ConnectorStats) MarkFailure(duration time.Duration, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isRunning = false
	s.LastFinishedAt = time.Now()
	s.LastFailureAt = s.LastFinishedAt
	s.LastDuration = duration
	s.ConsecutiveFailures++
	s.FailedRuns++
	
	if err != nil {
		s.LastError = err.Error()
	} else {
		s.LastError = "unknown error"
	}
}

func (s *ConnectorStats) Snapshot() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"name":                 s.Name,
		"is_running":           s.isRunning,
		"last_started_at":      s.LastStartedAt,
		"last_finished_at":     s.LastFinishedAt,
		"last_success_at":      s.LastSuccessAt,
		"last_failure_at":      s.LastFailureAt,
		"total_runs":           s.TotalRuns,
		"successful_runs":      s.SuccessfulRuns,
		"failed_runs":          s.FailedRuns,
		"records_fetched":      s.RecordsFetched,
		"records_normalized":   s.RecordsNormalized,
		"records_inserted":     s.RecordsInserted,
		"records_duplicate":    s.RecordsDuplicate,
		"records_rejected":     s.RecordsRejected,
		"consecutive_failures": s.ConsecutiveFailures,
		"last_duration_ms":     s.LastDuration.Milliseconds(),
		"last_error":           s.LastError,
	}
}

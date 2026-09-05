package ingestion

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/global-news/news-service/internal/domain"
	"github.com/google/uuid"
)

type MockRepo struct {
	saveResult domain.SaveResult
	saveErr    error
}

func (m *MockRepo) GetOrCreateSource(ctx context.Context, name, sourceType, baseURL string) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *MockRepo) SaveItemAndEvent(ctx context.Context, sourceID uuid.UUID, item *domain.SourceItem, event *domain.ThreatEvent) (domain.SaveResult, error) {
	return m.saveResult, m.saveErr
}

type MockConnector struct {
	name      string
	fetchFunc func() ([]domain.ExternalRecord, error)
	normFunc  func(rec domain.ExternalRecord) (*domain.ThreatEvent, error)
}

func (m *MockConnector) Name() string                                       { return m.name }
func (m *MockConnector) SourceType() string                                 { return "mock" }
func (m *MockConnector) BaseURL() string                                    { return "http://mock" }
func (m *MockConnector) Fetch(ctx context.Context) ([]domain.ExternalRecord, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc()
	}
	return nil, nil
}
func (m *MockConnector) Normalize(rec domain.ExternalRecord) (*domain.ThreatEvent, error) {
	if m.normFunc != nil {
		return m.normFunc(rec)
	}
	return &domain.ThreatEvent{}, nil
}

func TestScheduler_Isolation(t *testing.T) {
	repo := &MockRepo{}
	scheduler := NewScheduler(repo, nil, 1*time.Second)

	panicConnector := &MockConnector{
		name: "PanicConnector",
		fetchFunc: func() ([]domain.ExternalRecord, error) {
			panic("Intentional panic")
		},
	}

	successConnector := &MockConnector{
		name: "SuccessConnector",
		fetchFunc: func() ([]domain.ExternalRecord, error) {
			return []domain.ExternalRecord{{ExternalID: "1"}}, nil
		},
	}

	scheduler.Register(panicConnector)
	scheduler.Register(successConnector)

	// Run once directly to test isolation without timer
	scheduler.launchConnector(context.Background(), panicConnector)
	scheduler.launchConnector(context.Background(), successConnector)

	// Wait for goroutines to finish
	scheduler.wg.Wait()

	stats := scheduler.GetStats()
	pStats := stats["PanicConnector"].(map[string]interface{})
	sStats := stats["SuccessConnector"].(map[string]interface{})

	if pStats["failed_runs"].(int) != 1 {
		t.Errorf("Expected 1 failed run for panic connector")
	}
	if sStats["successful_runs"].(int) != 1 {
		t.Errorf("Expected 1 successful run for success connector")
	}
}

func TestScheduler_OverlapProtection(t *testing.T) {
	repo := &MockRepo{}
	scheduler := NewScheduler(repo, nil, 1*time.Second)

	blockCh := make(chan struct{})
	slowConnector := &MockConnector{
		name: "SlowConnector",
		fetchFunc: func() ([]domain.ExternalRecord, error) {
			<-blockCh
			return nil, nil
		},
	}
	scheduler.Register(slowConnector)

	// Launch first
	scheduler.launchConnector(context.Background(), slowConnector)
	time.Sleep(50 * time.Millisecond) // Ensure it's running

	// Launch second (should be skipped due to overlap)
	scheduler.launchConnector(context.Background(), slowConnector)
	
	stats := scheduler.GetStats()
	s := stats["SlowConnector"].(map[string]interface{})
	
	if s["total_runs"].(int) != 1 {
		t.Errorf("Expected only 1 total run despite 2 launches due to overlap protection")
	}

	close(blockCh)
	scheduler.wg.Wait()
}

func TestScheduler_Statistics(t *testing.T) {
	repo := &MockRepo{
		saveResult: domain.SaveDuplicate, // Force duplicate to check stats
	}
	scheduler := NewScheduler(repo, nil, 1*time.Second)

	mockConnector := &MockConnector{
		name: "StatsConnector",
		fetchFunc: func() ([]domain.ExternalRecord, error) {
			return []domain.ExternalRecord{
				{ExternalID: "valid"},
				{ExternalID: "invalid"},
			}, nil
		},
		normFunc: func(rec domain.ExternalRecord) (*domain.ThreatEvent, error) {
			if rec.ExternalID == "invalid" {
				return nil, errors.New("normalization failed")
			}
			return &domain.ThreatEvent{}, nil
		},
	}

	scheduler.Register(mockConnector)
	scheduler.launchConnector(context.Background(), mockConnector)
	scheduler.wg.Wait()

	stats := scheduler.GetStats()
	s := stats["StatsConnector"].(map[string]interface{})

	if s["records_fetched"].(int) != 2 {
		t.Errorf("Expected 2 fetched, got %v", s["records_fetched"])
	}
	if s["records_normalized"].(int) != 1 {
		t.Errorf("Expected 1 normalized, got %v", s["records_normalized"])
	}
	if s["records_rejected"].(int) != 1 {
		t.Errorf("Expected 1 rejected, got %v", s["records_rejected"])
	}
	if s["records_duplicate"].(int) != 1 {
		t.Errorf("Expected 1 duplicate, got %v", s["records_duplicate"])
	}
	if s["records_inserted"].(int) != 0 {
		t.Errorf("Expected 0 inserted, got %v", s["records_inserted"])
	}
}

package enrichment

import (
	"context"
	"testing"

	"github.com/global-news/news-service/internal/domain"
)

type MockEnricher struct {
	name    string
	err     error
	invoked bool
}

func (m *MockEnricher) Name() string {
	return m.name
}

func (m *MockEnricher) Enrich(ctx context.Context, record domain.ExternalRecord, event *domain.ThreatEvent) error {
	m.invoked = true
	return m.err
}

func TestEnrichmentPipeline_Execution(t *testing.T) {
	e1 := &MockEnricher{name: "E1"}
	e2 := &MockEnricher{name: "E2"}

	pipeline := NewPipeline(e1, e2)

	err := pipeline.Enrich(context.Background(), domain.ExternalRecord{}, &domain.ThreatEvent{})
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	if !e1.invoked || !e2.invoked {
		t.Error("Expected all enrichers to be invoked")
	}
}

func TestEnrichmentPipeline_FailureIsNonFatal(t *testing.T) {
	e1 := &MockEnricher{name: "E1", err: context.DeadlineExceeded}
	e2 := &MockEnricher{name: "E2"}

	pipeline := NewPipeline(e1, e2)

	// Even if E1 fails, E2 should run and overall pipeline returns nil.
	err := pipeline.Enrich(context.Background(), domain.ExternalRecord{}, &domain.ThreatEvent{})
	if err != nil {
		t.Fatalf("Expected nil (non-fatal error handling), got %v", err)
	}

	if !e1.invoked || !e2.invoked {
		t.Error("Expected all enrichers to be invoked despite failure")
	}
}

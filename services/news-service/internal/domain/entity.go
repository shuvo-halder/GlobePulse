package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EntityType represents the controlled taxonomy of intelligence entities
type EntityType string

const (
	EntityTypeCountry        EntityType = "country"
	EntityTypeLocation       EntityType = "location"
	EntityTypeOrganization   EntityType = "organization"
	EntityTypePerson         EntityType = "person"
	EntityTypeGroup          EntityType = "group"
	EntityTypeDomain         EntityType = "domain"
	EntityTypeIP             EntityType = "ip"
	EntityTypeURL            EntityType = "url"
	EntityTypeInfrastructure EntityType = "infrastructure"
)

// ExtractionConfidence defines the certainty level of the entity extraction
type ExtractionConfidence string

const (
	ConfidenceExact  ExtractionConfidence = "exact"
	ConfidenceHigh   ExtractionConfidence = "high"
	ConfidenceMedium ExtractionConfidence = "medium"
	ConfidenceLow    ExtractionConfidence = "low"
)

// ExtractionMethod defines the mechanism used to extract the entity
type ExtractionMethod string

const (
	MethodStructured    ExtractionMethod = "structured"
	MethodDeterministic ExtractionMethod = "deterministic"
	MethodRegex         ExtractionMethod = "regex"
)

// IntelligenceEntity represents an extracted, normalized entity associated with threat intelligence
type IntelligenceEntity struct {
	ID               uuid.UUID            `json:"id"`
	Type             EntityType           `json:"type"`
	Value            string               `json:"value"`            // Original raw extracted value
	NormalizedValue  string               `json:"normalized_value"` // Canonical normalized value
	Confidence       ExtractionConfidence `json:"confidence"`
	ExtractionMethod ExtractionMethod     `json:"extraction_method"`
	SourceField      string               `json:"source_field,omitempty"`
	Metadata         []byte               `json:"metadata,omitempty"`
	CreatedAt        time.Time            `json:"created_at,omitempty"`
	UpdatedAt        time.Time            `json:"updated_at,omitempty"`
}

// EntityExtractor defines the contract for extracting entities from a canonical ThreatEvent
type EntityExtractor interface {
	Name() string
	Extract(ctx context.Context, event *ThreatEvent) ([]IntelligenceEntity, error)
}

// EntityPipeline coordinates multiple extractors and deduplicates extracted entities
type EntityPipeline interface {
	Extract(ctx context.Context, event *ThreatEvent) ([]IntelligenceEntity, error)
}

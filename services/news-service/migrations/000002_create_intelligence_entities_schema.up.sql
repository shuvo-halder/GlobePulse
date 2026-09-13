CREATE TABLE intelligence_entities (
    id UUID PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    value TEXT NOT NULL,
    normalized_value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_intelligence_entities_type_norm UNIQUE (entity_type, normalized_value)
);

CREATE INDEX idx_intelligence_entities_type ON intelligence_entities(entity_type);
CREATE INDEX idx_intelligence_entities_normalized_val ON intelligence_entities(normalized_value);

CREATE TABLE threat_event_entities (
    threat_event_id UUID NOT NULL REFERENCES threat_events(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES intelligence_entities(id) ON DELETE CASCADE,
    confidence VARCHAR(50) NOT NULL,
    extraction_method VARCHAR(50) NOT NULL,
    source_field VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (threat_event_id, entity_id)
);

CREATE INDEX idx_threat_event_entities_entity_id ON threat_event_entities(entity_id);
CREATE INDEX idx_threat_event_entities_threat_event_id ON threat_event_entities(threat_event_id);

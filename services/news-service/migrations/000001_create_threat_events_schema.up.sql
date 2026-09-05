CREATE TABLE sources (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    base_url VARCHAR(255),
    license_info TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE source_items (
    id UUID PRIMARY KEY,
    source_id UUID NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    external_id VARCHAR(255) NOT NULL,
    url TEXT,
    title TEXT NOT NULL,
    raw_metadata JSONB,
    published_at TIMESTAMP WITH TIME ZONE,
    collected_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_source_external_id UNIQUE (source_id, external_id)
);

CREATE TABLE threat_events (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    event_type VARCHAR(100) NOT NULL,
    category VARCHAR(100),
    severity VARCHAR(50),
    confidence NUMERIC(5, 2),
    occurred_at TIMESTAMP WITH TIME ZONE,
    detected_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    country VARCHAR(100),
    location_details TEXT,
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_threat_events_occurred_at ON threat_events(occurred_at);
CREATE INDEX idx_threat_events_event_type ON threat_events(event_type);
CREATE INDEX idx_threat_events_country ON threat_events(country);
CREATE INDEX idx_threat_events_location ON threat_events(latitude, longitude);

CREATE TABLE threat_event_source_items (
    threat_event_id UUID NOT NULL REFERENCES threat_events(id) ON DELETE CASCADE,
    source_item_id UUID NOT NULL REFERENCES source_items(id) ON DELETE CASCADE,
    PRIMARY KEY (threat_event_id, source_item_id)
);

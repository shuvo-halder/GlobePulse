CREATE TABLE countries (
    code VARCHAR(3) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    region VARCHAR(100),
    population BIGINT,
    risk_score FLOAT DEFAULT 0.0,
    sentiment FLOAT DEFAULT 0.0,
    trending_topics JSONB DEFAULT '[]',
    related_countries JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_countries_name ON countries(name);
CREATE INDEX idx_countries_risk ON countries(risk_score);

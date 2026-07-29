CREATE TABLE analytics_events (
    id UUID PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    payload TEXT,
    country_code VARCHAR(3) NOT NULL,
    sentiment FLOAT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE country_metrics (
    country_code VARCHAR(3) NOT NULL,
    date DATE NOT NULL,
    total_events BIGINT DEFAULT 0,
    avg_sentiment FLOAT DEFAULT 0.0,
    trending_score FLOAT DEFAULT 0.0,
    PRIMARY KEY (country_code, date)
);

CREATE INDEX idx_analytics_events_country_time ON analytics_events(country_code, timestamp);
CREATE INDEX idx_country_metrics_date ON country_metrics(date);

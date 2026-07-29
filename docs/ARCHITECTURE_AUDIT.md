# Architecture Reconciliation Audit

This document serves as the complete architecture reconciliation audit for the Globepulse monorepo, analyzing the existing services (`auth-service`, `news-service`, `country-service`, `analytics-service`, `ai-service`) for architectural drift and inconsistencies, and standardizing their contracts.

---

## 1. Architecture Drift Report

### Auth Service vs. Other Services
**Issue #1: Missing Distributed Authentication Verification**
* **Drift:** `auth-service` issues JWTs and tracks sessions in Redis. However, downstream services (`news-service`, `country-service`, `analytics-service`) do not have corresponding JWT validation middlewares or access to the shared session store logic to verify these tokens.
* **Fix:** Standardize an API Gateway pattern or shared JWT validation library/middleware across all Go and Python services.

### AI Service vs. News Service
**Issue #2: Async Queue Serialization Mismatch (Celery vs. Raw AMQP)**
* **Drift:** `ai-service` uses Celery for background processing on the `ai_analysis_queue`. Celery uses a proprietary envelope format for its messages. If `news-service` (Go) attempts to publish a standard JSON message directly to `ai_analysis_queue` using `streadway/amqp`, Celery workers will fail to parse it.
* **Fix:** Either `news-service` must publish using Celery's exact JSON message protocol, or `news-service` should trigger analysis via the `ai-service`'s REST endpoint (`/api/v1/analysis/analyze`), which then handles the Celery delegation internally.

### Analytics Service vs. AI Service
**Issue #3: Event Payload Schemas**
* **Drift:** `analytics-service` consumes an `AnalyticsEvent` where the `payload` is just a string (TEXT). When `ai-service` completes an analysis, the structured output (topics, entities, sentiment) needs to be tracked. The lack of a strongly typed payload makes querying nested AI insights difficult.
* **Fix:** Define a canonical JSON schema for `AnalyticsEvent.payload` and migrate the column to `JSONB` in PostgreSQL.

### Country Service vs. Analytics Service
**Issue #4: Naming Conventions & Data Types**
* **Drift:** `country-service` refers to the country identifier as `Code` (VARCHAR 3), while `analytics-service` refers to it as `CountryCode` (VARCHAR 3). 
* **Fix:** Standardize field naming to `country_code` across all cross-service DTOs and event contracts.

### AI Service DB Implementation
**Issue #5: Unused Relational Database Configuration**
* **Drift:** `ai-service` configures PostgreSQL (`database.py`) and creates tables (`Base.metadata.create_all`), but no SQLAlchemy ORM models actually exist, and analysis results are only stored in Redis.
* **Fix:** Define SQLAlchemy models for `AnalysisResult` in `ai-service` to persist historical AI insights permanently, or remove the PostgreSQL dependency if Redis is sufficient for the caching layer.

---

## 2. Shared Contract Specification

To ensure consistency, the following canonical schemas should be implemented in a new `packages/shared-types` library (for Go/Python).

### `User`
```json
{
  "id": "uuid",
  "email": "string",
  "first_name": "string",
  "last_name": "string",
  "role": "string (admin | user)",
  "is_verified": "boolean"
}
```

### `Country`
```json
{
  "country_code": "string(3)",
  "name": "string",
  "region": "string",
  "risk_score": "float",
  "sentiment": "float"
}
```

### `NewsArticle`
```json
{
  "id": "uuid",
  "source_id": "string",
  "title": "string",
  "summary": "string",
  "content": "string",
  "url": "string",
  "language": "string",
  "published_at": "timestamp"
}
```

### `AIAnalysisResult` (replaces implicit types)
```json
{
  "news_id": "uuid",
  "summary": "string",
  "sentiment_score": "float",
  "sentiment_label": "string",
  "entities": [{"name": "string", "type": "string", "confidence": "float"}],
  "topics": [{"name": "string", "score": "float"}],
  "countries": ["string"],
  "event_type": "string",
  "importance_score": "float"
}
```

### `AnalyticsEvent`
```json
{
  "id": "uuid",
  "event_type": "string",
  "country_code": "string",
  "sentiment": "float",
  "timestamp": "timestamp",
  "payload": "jsonb" 
}
```

---

## 3. Service Dependency Matrix

| Source Service | Target Service | Communication Type | Purpose |
| :--- | :--- | :--- | :--- |
| `API Gateway / Client` | `auth-service` | REST | Login, Registration, Token generation |
| `API Gateway / Client` | `news-service` | REST | Fetching news feeds |
| `news-service` | `ai-service` | REST | Triggering news analysis (`/api/v1/analysis/analyze`) |
| `ai-service` (Worker) | `analytics-service`| RabbitMQ (`analytics_events_queue`) | Publishing `NewsAnalyzed` events for metrics aggregation |
| `API Gateway / Client` | `analytics-service`| REST | Fetching global and country metrics dashboards |
| `API Gateway / Client` | `country-service` | REST | Fetching country metadata and current status |
| `analytics-service` | `country-service` | REST | Analytics fetching baseline country data (e.g. for heatmaps) |

---

## 4. Event Contract Specification

### `NewsCreated`
Published by `news-service` when a new article is ingested.
```json
{
  "event_id": "uuid",
  "timestamp": "iso8601",
  "type": "news.created",
  "data": {
    "article_id": "uuid",
    "title": "string",
    "content": "string",
    "language": "string"
  }
}
```

### `AIAnalysisCompleted`
Published by `ai-service` when processing finishes.
```json
{
  "event_id": "uuid",
  "timestamp": "iso8601",
  "type": "analysis.completed",
  "data": {
    "article_id": "uuid",
    "sentiment_score": 0.85,
    "event_type": "market crash",
    "countries_detected": ["USA", "GBR"]
  }
}
```

### `AnalyticsCalculated`
Published by `analytics-service` after daily metric aggregation.
```json
{
  "event_id": "uuid",
  "timestamp": "iso8601",
  "type": "analytics.daily_calculated",
  "data": {
    "date": "YYYY-MM-DD",
    "country_code": "USA",
    "trending_score": 95.5
  }
}
```

---

## 5. API Compatibility Report

* **Missing Endpoints:** 
  * `ai-service` does not have a webhook or callback endpoint to notify `news-service` that async processing is complete.
  * No centralized `/health` aggregator. Each service implements its own.
* **Backward Compatibility Risks:**
  * Changing `analytics-service` payload from `TEXT` to `JSONB` will require a database migration and a code update in the event consumer.
* **Duplicate Endpoints:**
  * Sentiment and trending topics are being implicitly tracked in both `country-service` (as state) and `analytics-service` (as time-series). A clear boundary needs to be established (e.g., `analytics-service` is the source of truth, `country-service` queries it).

---

## 6. Database Consistency Report

* **Auth Service DB (`users`)**
  * Consistent. Properly uses UUIDs.
* **Analytics Service DB (`analytics_events`, `country_metrics`)**
  * `payload` should be `JSONB` instead of `TEXT` for robust querying.
  * Needs a migration to convert `VARCHAR(3)` country codes to strictly adhere to ISO 3166-1 alpha-3.
* **AI Service DB**
  * Missing table schemas. Recommendation: Create an `ai_analysis_results` table with foreign key reference `news_id` pointing logically to `news-service`.
* **Cross-DB Integrity:**
  * Microservices share the same `postgres` host but separate logical schemas/tables. Foreign keys across these microservice boundaries cannot be enforced at the database level. Application-level validation or eventual consistency mechanisms (via RabbitMQ) must handle orphaned records (e.g., a country is deleted in `country-service`, we must emit a `CountryDeleted` event to purge metrics in `analytics-service`).

---

## 7. Final Unified Architecture

### System Topology

1. **Edge Layer (To Be Implemented)**
   * An API Gateway (e.g., NGINX / Kong / Traefik) that routes `/api/v1/auth/*` to `auth-service`, `/api/v1/news/*` to `news-service`, etc.
   * Gateway handles JWT validation using public keys or via a middleware call to `auth-service`, ensuring downstream services do not need to reimplement auth validation.

2. **Core Services Layer**
   * **Auth Service:** Manages identity, JWTs, and Redis sessions.
   * **News Service:** Ingests and stores raw news. Triggers AI analysis.
   * **Country Service:** Manages static and slowly-changing country metadata.
   * **Analytics Service:** Time-series aggregation, metrics, and heatmap data. Source of truth for historical trends.

3. **Intelligence Layer**
   * **AI Service & Workers:** Python-based ML pipelines. Receives analysis requests from `news-service` (via REST), processes them asynchronously via Celery/Redis, and publishes `AIAnalysisCompleted` events to RabbitMQ.

4. **Event Bus (RabbitMQ)**
   * Central nervous system for decoupled updates. `analytics-service` subscribes to analysis events to update realtime metrics. `country-service` subscribes to daily analytics events to update its cached "current sentiment" fields.

### Data Storage Strategy
* **PostgreSQL:** Primary persistence (Auth, News, Countries, Time-series metrics).
* **Redis:** Short-lived caching (Country metrics, Heatmaps), Session storage (Auth), and Celery Broker/Backend (AI Service).
* **RabbitMQ:** Asynchronous event broadcasting (Pub/Sub pattern).

# Complete Refactoring Plan

Based on the Architecture Drift Report, this step-by-step migration plan ensures 100% API contract compatibility, event compatibility, and database consistency across the Globepulse monorepo.

---

## 1. Shared Infrastructure & Authentication

**Goal:** Centralize authentication validation and define canonical types.

* **Files to create/modify:**
  * Create `packages/shared-types` with Go (`.go`) and Python (`.py`) definitions for `User`, `Country`, `NewsArticle`, `AIAnalysisResult`, and `AnalyticsEvent`.
  * Create `packages/auth-middleware` (or equivalent per language) to standardly validate JWTs using a shared secret or JWKS endpoint exposed by `auth-service`.

---

## 2. Auth Service

**Goal:** Provide the necessary endpoints for downstream services to validate JWTs.

* **Files to modify:**
  * `auth-service/internal/handler/auth.go`
  * `auth-service/internal/handler/router.go`
* **APIs to modify:**
  * Expose a `GET /api/v1/auth/verify` endpoint (or a JWKS endpoint) that other services can call or fetch keys from to validate tokens locally.
* **Database changes:**
  * None required (users table is compliant).
* **Event schema changes:**
  * None required.

---

## 3. News Service

**Goal:** Fix the async queue serialization mismatch and adopt standard DTOs.

* **Files to modify:**
  * `news-service/internal/service/news_service.go`
  * `news-service/internal/domain/news.go`
  * `news-service/internal/handler/http/news_handler.go`
* **DTOs to modify:**
  * Use the canonical `NewsArticle` schema from `shared-types`.
* **APIs to modify:**
  * Wrap protected endpoints with the new JWT validation middleware.
  * **Critical:** Instead of pushing raw JSON to the RabbitMQ `ai_analysis_queue` (which breaks Celery), refactor the service to make an async HTTP POST request to `ai-service` at `/api/v1/analysis/analyze`.
* **Database changes:**
  * Ensure the `author_id` foreign key correctly references a standard UUID.
* **Event schema changes:**
  * Publish a strictly-typed `NewsCreated` AMQP event to RabbitMQ when an article is successfully ingested.

---

## 4. AI Service

**Goal:** Implement unused relational DB, standard schemas, and event publishing.

* **Files to modify:**
  * `ai-service/app/models/domain.py` (New - SQLAlchemy models)
  * `ai-service/app/models/schemas.py`
  * `ai-service/app/worker/tasks.py`
  * `ai-service/app/api/endpoints.py`
* **DTOs to modify:**
  * Update `AnalysisResult` Pydantic model to strictly match the canonical `AIAnalysisResult` schema.
* **APIs to modify:**
  * Wrap endpoints with JWT validation dependency (FastAPI `Depends`).
* **Database changes:**
  * Define `AIAnalysisResult` as an SQLAlchemy model.
  * Execute Alembic/SQLAlchemy migration to create the `ai_analysis_results` table, tying it logically via `news_id` to the news-service.
* **Event schema changes:**
  * Inside the Celery task (`analyze_news_task`), after saving to DB and Redis, publish an `AIAnalysisCompleted` event directly to RabbitMQ using standard `pika` (raw AMQP, not Celery envelope) so `analytics-service` can consume it seamlessly.

---

## 5. Analytics Service

**Goal:** Fix payload data type, naming drifts, and publish aggregated events.

* **Files to modify:**
  * `analytics-service/internal/domain/analytics.go`
  * `analytics-service/internal/repository/postgres/analytics_repo.go`
  * `analytics-service/internal/event/consumer.go`
  * `analytics-service/migrations/000002_alter_payload_jsonb.up.sql` (New)
* **DTOs to modify:**
  * Refactor `AnalyticsEvent.Payload` from `string` to a typed nested struct (`map[string]interface{}` or typed `AIAnalysisResult`).
  * Ensure the `CountryCode` field in `CountryMetrics` adheres to canonical naming tags (`json:"country_code" db:"country_code"`).
* **APIs to modify:**
  * Wrap endpoints with JWT validation middleware.
* **Database changes:**
  * Create a migration: `ALTER TABLE analytics_events ALTER COLUMN payload TYPE jsonb USING payload::jsonb;`
* **Event schema changes:**
  * Update the RabbitMQ consumer (`consumer.go`) to explicitly listen for and parse `AIAnalysisCompleted` events from `ai-service`.
  * After the daily scheduled aggregation, publish an `AnalyticsCalculated` event to RabbitMQ for `country-service` to consume.

---

## 6. Country Service

**Goal:** Standardize field naming and establish a clear source of truth for metrics.

* **Files to modify:**
  * `country-service/internal/domain/country.go`
  * `country-service/internal/repository/postgres/country_repo.go`
  * `country-service/internal/handler/http/country_handler.go`
  * `country-service/internal/event/consumer.go` (New)
  * `country-service/migrations/000002_rename_code_column.up.sql` (New)
* **DTOs to modify:**
  * Rename the identifier field from `Code` to `CountryCode` (with `json:"country_code" db:"country_code"`).
* **APIs to modify:**
  * Wrap endpoints with JWT validation middleware.
  * Ensure endpoints return the canonical `Country` schema.
* **Database changes:**
  * Create a migration: `ALTER TABLE countries RENAME COLUMN code TO country_code;`
* **Event schema changes:**
  * Implement an AMQP consumer to listen for `AnalyticsCalculated` events published by `analytics-service`.
  * Use these events to update the `sentiment` and `risk_score` caches in the `countries` table, strictly treating `analytics-service` as the source of truth.

---

## 7. Migration Execution Steps

1. **Step 1: Create Shared Libraries:** Implement `packages/shared-types` and authentication middleware.
2. **Step 2: Database Migrations:** Run up-migrations on `analytics-service` (TEXT to JSONB), `country-service` (code to country_code), and `ai-service` (create tables).
3. **Step 3: Update DTOs & DB Models:** Update the domain structs in all 4 downstream services to match the database changes and canonical schemas.
4. **Step 4: API & JWT Refactoring:** Inject JWT middleware across all routers/handlers. Refactor `news-service` to call `ai-service` via HTTP.
5. **Step 5: Event Bus Rewiring:** 
   - Add RabbitMQ publish logic to `ai-service` worker.
   - Update `analytics-service` consumer to handle the new `AIAnalysisCompleted` JSON structure.
   - Add RabbitMQ publish logic for `AnalyticsCalculated` to `analytics-service`.
   - Add RabbitMQ consumer to `country-service` to listen to `AnalyticsCalculated`.
6. **Step 6: Testing & Verification:** Verify End-to-End flow: News Creation -> HTTP to AI Service -> Worker processes -> Publishes `AIAnalysisCompleted` -> Analytics consumes -> Publishes `AnalyticsCalculated` -> Country consumes.
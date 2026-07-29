# Comprehensive Architecture Reconciliation Report

This report outlines every instance of architectural drift, API mismatch, duplicate model, inconsistent DTO, and incompatible contract currently present across the Globepulse microservices monorepo.

---

## 1. Domain Model & DTO Inconsistencies

### 1.1 `NewsArticle` Authorship Missing
* **Service:** `news-service`
* **Issue:** The `NewsArticle` struct (`internal/domain/news.go`) lacks any concept of an `AuthorID` to map the article back to the `auth-service`'s `User` model. It only tracks a string `SourceID`.
* **Correction Required:** Add an `AuthorID` (UUID) field to the canonical `NewsArticle` schema.

### 1.2 `NewsArticle` Language Missing
* **Service:** `news-service` vs `ai-service`
* **Issue:** `ai-service`'s `AnalysisRequest` schema explicitly expects a `language: str = "en"` field to determine the correct NLP pipeline path. However, `news-service`'s `NewsArticle` struct entirely omits the `Language` field.
* **Correction Required:** Add `Language string` to `NewsArticle` and migrate the Go database schema.

### 1.3 `Country` Identity Field Drift
* **Service:** `country-service` vs `analytics-service`
* **Issue:** `country-service` identifies countries using the field `Code` (e.g. `json:"code" db:"code"`). `analytics-service` uses `CountryCode` (e.g. `json:"country_code" db:"country_code"`).
* **Correction Required:** Standardize to `country_code` across the board in the canonical `shared-types` definitions.

### 1.4 Sentiment Data Types and Scope
* **Service:** `ai-service` vs `news-service` / `country-service`
* **Issue:** `ai-service` outputs `sentiment_score` (float) and `sentiment_label` (string: POSITIVE, NEGATIVE, NEUTRAL). `news-service` and `country-service` only capture a single `Sentiment float64`.
* **Correction Required:** Ensure DTOs globally support both the float score and the categorical string label to preserve intelligence resolution.

---

## 2. API Contract & Invocation Mismatches

### 2.1 The Broken AI Trigger Pipeline
* **Service:** `news-service` -> `ai-service`
* **Issue:** Currently, `news-service`'s `CreateArticle` method *only* saves the article to the PostgreSQL database. It never triggers the AI analysis pipeline, rendering the entire `ai-service` orphaned.
* **Correction Required:** `news-service` must perform an asynchronous HTTP POST to `ai-service` (`/api/v1/analysis/analyze`) immediately after database persistence.

### 2.2 Celery vs. Raw AMQP Envelope Mismatch
* **Service:** `ai-service` vs `analytics-service`
* **Issue:** The architecture dictates RabbitMQ as the event bus. However, `ai-service` relies exclusively on Celery for background tasks. Celery wraps messages in proprietary JSON envelopes. If `analytics-service` (Go) attempts to consume from a Celery-managed queue, or if `news-service` tries to publish directly to one, serialization will fail.
* **Correction Required:** The Python `ai-service` must use standard `pika` to publish raw, canonical JSON `AIAnalysisCompleted` events to a dedicated RabbitMQ fanout exchange, independent of its internal Celery state.

### 2.3 Analytics Payload Opaqueness
* **Service:** `analytics-service`
* **Issue:** `analytics-service` expects an `AnalyticsEvent` where the `payload` is defined as a `string` (TEXT). AI results are heavily structured (lists of topics, entities, scores). Stringifying them destroys the ability to perform complex OLAP queries.
* **Correction Required:** Migrate the `payload` column to `JSONB` in Postgres, and update the Go struct to use `map[string]interface{}` or a strongly-typed nested struct.

---

## 3. Duplication & Siloed Infrastructure

### 3.1 Authentication Isolation
* **Issue:** `auth-service` issues JWTs and tracks sessions in Redis. But `news-service`, `country-service`, and `analytics-service` have **zero** JWT validation middleware. They cannot protect their own routes because the auth logic is completely siloed.
* **Correction Required:** Extract a `packages/shared-auth` library containing the JWT validation middleware so all services can uniformly protect endpoints.

### 3.2 Boilerplate Duplication
* **Issue:** Every single Go service (`auth-service`, `news-service`, `country-service`, `analytics-service`) has its own identical copy of:
  * `pkg/logger/logger.go` (Zap setup)
  * `internal/config/config.go` (Viper parsing)
* **Correction Required:** Extract these into `packages/shared-logger` and `packages/shared-config` to enforce a single source of truth for telemetry and initialization.

---

## 4. Event Bus Contract Definitions (Target State)

No events are currently being published in the monorepo. We must establish the following canonical event schemas in `packages/shared-events`:

1. **`NewsCreated`**
   * **Publisher:** `news-service`
   * **Subscribers:** (Future email digest workers, audit logs)
   * **Schema:** `article_id`, `title`, `language`, `timestamp`

2. **`AIAnalysisCompleted`**
   * **Publisher:** `ai-service`
   * **Subscribers:** `analytics-service` (for aggregation)
   * **Schema:** `article_id`, `sentiment_score`, `sentiment_label`, `topics`, `entities`, `countries_detected`

3. **`AnalyticsCalculated`**
   * **Publisher:** `analytics-service` (Daily cron)
   * **Subscribers:** `country-service` (to update its read-optimized caches)
   * **Schema:** `country_code`, `date`, `trending_score`, `avg_sentiment`

---

## 5. Database Consistency Report

1. **AI Service Persistence:** The `ai-service` spins up an SQLAlchemy engine but never actually defines the `Base` ORM models to store `AIAnalysisResult`. It just caches to Redis. Historical intelligence will be lost upon cache expiration.
2. **Missing Foreign Keys:** Because microservices have separate logical schemas, hard Foreign Keys (e.g. `news_id` in Analytics pointing to `id` in News) cannot be enforced by the RDBMS. Application-level cascades or event-driven deletions (e.g. listening for `NewsDeleted`) must be implemented to prevent orphaned metrics.

# Monorepo Architecture Correction Report

This report outlines the structural deficiencies in the current monorepo implementation and defines the target state for a highly cohesive, decoupled, and DRY (Don't Repeat Yourself) microservices architecture.

---

## 1. Current State vs. Target State

### The Problem
Currently, the services (`auth-service`, `news-service`, `country-service`, `analytics-service`, `ai-service`) were generated incrementally. This led to isolated silos with massive duplication:
* Every Go service has its own copy of `pkg/logger`.
* Configuration parsing (Viper) is duplicated across all Go services.
* JWT logic exists only in `auth-service`, making it impossible for other services to validate requests.
* Event publishing and consumption lack a centralized contract.
* Type definitions (like `Country` or `AnalyticsEvent`) drift between services.
* The Python `ai-service` is isolated and lacks typed contracts with the Go services.

### The Ideal Monorepo Structure
We must transition to a workspace-based monorepo (e.g., using Go Workspaces and Python local packages) with a dedicated `packages/` directory for shared libraries.

```text
globepulse/
├── packages/
│   ├── shared-types/     # Polyglot DTOs, JSON Schemas, Protobufs
│   ├── shared-auth/      # JWT validation, Role-Based Access Control (RBAC) middlewares
│   ├── shared-events/    # RabbitMQ clients, canonical event schemas
│   ├── shared-config/    # Environment parsing, secrets management
│   ├── shared-logger/    # Standardized structured logging (Zap for Go, logging for Py)
│   └── shared-utils/     # Error handlers, pagination, DB/Redis connection pooling
├── services/
│   ├── auth-service/     # Identity, session, and JWT issuance
│   ├── news-service/     # Article ingestion and serving
│   ├── country-service/  # Static country metadata and sentiment state
│   ├── analytics-service/# Time-series aggregation and heatmap generation
│   └── ai-service/       # Python-based NLP/ML processing
├── infra/                # Docker Compose, Terraform, K8s manifests
└── docs/                 # Architecture audits, API specs, Playbooks
```

---

## 2. Missing Shared Packages Definition

### `packages/shared-types`
* **Purpose:** Single source of truth for cross-service communication payloads.
* **Contents:**
  * Canonical definitions for `User`, `Country`, `NewsArticle`, `Topic`, `Event`, `Sentiment`, `AnalyticsMetric`, and `AIAnalysisResult`.
  * **Polyglot Strategy:** Since `ai-service` is Python and the rest are Go, this package should ideally use JSON Schema, Protobufs, or OpenAPI specifications to auto-generate native types for both languages.

### `packages/shared-auth`
* **Purpose:** Standardize API security and session handling.
* **Contents:**
  * JWT parsing and validation logic.
  * Public key / Secret key retrieval mechanisms.
  * Gin (Go) and FastAPI (Python) middlewares to protect routes.
  * RBAC utilities (e.g., `RequireAdmin()`).

### `packages/shared-events`
* **Purpose:** Strictly typed asynchronous communication.
* **Contents:**
  * RabbitMQ connection lifecycle managers (reconnect logic, dead-letter queues).
  * Structs/Pydantic models for `NewsCreated`, `NewsUpdated`, `CountryUpdated`, `AnalyticsCalculated`, and `AIAnalysisCompleted`.
  * Standardized envelope format (Event ID, Timestamp, Correlation ID, Payload).

### `packages/shared-config`
* **Purpose:** Standardize how 12-factor apps load environment variables.
* **Contents:**
  * Shared Viper setup for Go.
  * Base Pydantic `BaseSettings` setup for Python.
  * Common constants (e.g., default pagination limits, standard timeouts).

### `packages/shared-logger`
* **Purpose:** Uniform telemetry and observability.
* **Contents:**
  * Standardized Zap logger configuration for Go (JSON in production, colored console in dev).
  * Standardized Python logger for `ai-service`.
  * Request correlation ID injection.

### `packages/shared-utils`
* **Purpose:** Boilerplate reduction.
* **Contents:**
  * Standardized API Error Responses (`{ "error": { "code": "...", "message": "..." } }`).
  * Database (PostgreSQL/Redis) connection bootstrap logic.

---

## 3. Service Consumption Matrix

How each service will integrate with the new `packages/` directory:

### `auth-service`
* **Consumes `shared-types`:** Uses the canonical `User` model.
* **Consumes `shared-auth`:** Uses core JWT issuance libraries.
* **Consumes `shared-utils/config/logger`:** Replaces its local `pkg/` implementations for DB/Redis connections and logging.

### `news-service`
* **Consumes `shared-types`:** Uses canonical `NewsArticle` and `AIAnalysisResult` (when receiving callbacks).
* **Consumes `shared-auth`:** Applies generic auth middleware to endpoints (e.g., `POST /news`).
* **Consumes `shared-events`:** Uses the standardized RabbitMQ client to publish `NewsCreated` events securely.

### `ai-service` (Python)
* **Consumes `shared-types`:** Generates Pydantic schemas from the shared definitions to ensure `AnalysisResult` perfectly matches what Go expects.
* **Consumes `shared-auth`:** Uses the shared FastAPI auth middleware to secure the `/analyze` POST endpoint.
* **Consumes `shared-events`:** Uses the shared AMQP publishing structure to emit `AIAnalysisCompleted` safely to RabbitMQ (avoiding Celery envelope mismatches).

### `analytics-service`
* **Consumes `shared-types`:** Uses `AnalyticsMetric`, `AIAnalysisResult`, and `Country` canonical types.
* **Consumes `shared-auth`:** Applies JWT middleware to dashboard metric APIs.
* **Consumes `shared-events`:** 
  * Uses shared AMQP consumer wrappers to ingest `AIAnalysisCompleted`.
  * Publishes `AnalyticsCalculated` events.

### `country-service`
* **Consumes `shared-types`:** Uses canonical `Country` model.
* **Consumes `shared-auth`:** Protects country endpoints using shared middleware.
* **Consumes `shared-events`:** Subscribes to `AnalyticsCalculated` using robust shared AMQP consumer wrappers to update its own read-heavy caches.

---

## 4. Path to Implementation

1. **Move Services:** Move existing root-level folders (`auth-service`, `news-service`, etc.) into a new `services/` directory.
2. **Create Workspaces:**
   * For Go: Run `go work init` and add the services and new `packages/*` directories.
   * For Python: Configure `pyproject.toml` or `setup.py` in the shared Python packages to allow `pip install -e ../packages/shared-python` inside `ai-service`.
3. **Extract Duplication:** Delete `pkg/logger`, `pkg/jwt`, and `config/` from individual services and import them from `packages/` instead.
4. **Enforce Types:** Replace localized DTOs with imports from `shared-types`.
5. **Standardize Events:** Rip out raw `streadway/amqp` and `pika` implementations in favor of the resilient wrappers provided by `shared-events`.

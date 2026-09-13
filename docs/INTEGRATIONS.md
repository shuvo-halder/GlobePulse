# Integrations

This document details the external integrations present or planned within the repository.

## 1. Gemini AI API
- **Purpose:** Used for AI threat assessment, NLP text analysis, and geopolitical forecasting.
- **Configuration:** Managed via the `GEMINI_API_KEY` environment variable (defined in `.env.example`).
- **Current Status:** **Planned**. The environment variable is defined and documented, and the `@google/genai` SDK is present in the frontend `package.json`, but no actual implementation calling the Gemini API was found in either the frontend code or the Python `ai-service`.

## 2. PostgreSQL
- **Purpose:** Primary relational data store.
- **Configuration:** Controlled via `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME`.
- **Current Status:** **Implemented**. The `auth-service` connects and manages users and audit logs.

## 3. Redis
- **Purpose:** Session store and potential caching layer.
- **Configuration:** Controlled via `REDIS_ADDR`, `REDIS_PORT`, `REDIS_PASS`.
- **Current Status:** **Implemented**. The `auth-service` uses Redis to store active session IDs for JWT validation.

## 4. RabbitMQ
- **Purpose:** Message broker for asynchronous task processing.
- **Configuration:** Controlled via `RABBITMQ_URL` (`amqp://guest:guest@rabbitmq:5672/`).
- **Current Status:** **Partially Implemented**. The `ai-service` configures a Celery worker to connect to RabbitMQ, but no messages are currently produced or consumed by the application code.

## 5. React Globe GL / Three.js
- **Purpose:** 3D interactive globe visualization in the frontend.
- **Configuration:** Relies on standard library imports.
- **Current Status:** **Implemented**. Used successfully in `GlobeView.tsx` and `App.tsx` to render the primary application interface.

## 6. USGS Earthquake API
- **Purpose:** Source for real-time global earthquake data.
- **Endpoint:** `https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_hour.geojson`
- **Format:** GeoJSON
- **Current Status:** **Implemented**. The `news-service` fetches this feed periodically using a custom connector. Event IDs are mapped to `ExternalID` for deduplication. Extracted details include magnitude, timestamp (`OccurredAt`), exact coordinates (Longitude/Latitude/Depth), and human-readable place description. Magnitude acts as a source constraint on severity calculation. Events are persisted to PostgreSQL as `natural_disaster` categories and `earthquake` event types.

## 7. GDELT Project API
- **Purpose:** Media monitoring and geopolitical event detection.
- **Endpoint:** `https://api.gdeltproject.org/api/v2/doc/doc`
- **Current Status:** **Implemented**. Extracts news signals. Uses article URL as `ExternalID`. Handles missing physical coordinates safely by passing `NULL` to PostgreSQL instead of fake fallbacks like `(0,0)`. Distinguishes `DetectedAt` (source publication time) from `OccurredAt` (actual event time, which is left safely unknown).

## 8. ReliefWeb API
- **Purpose:** Humanitarian reports and updates.
- **Endpoint:** `https://api.reliefweb.int/v1/reports`
- **Current Status:** **Implemented**. Extracts humanitarian updates. Uses ReliefWeb report ID as `ExternalID`. Preserves source provenance by accurately tagging records as `humanitarian_report` rather than fabricating confirmed physical events.

## Data Ingestion & Deduplication Architecture
- **Idempotency Strategy:** All connectors map external records to a source-scoped identity (`source_id`, `external_id`).
- **Persistence Safety:** Inserts are protected against race conditions and duplicates via PostgreSQL `ON CONFLICT (source_id, external_id) DO NOTHING`. Database uniqueness remains authoritative.
- **Normalization Semantics:** Connectors strictly enforce canonical invariants (e.g., coordinate bounds `[-90, 90]`/`[-180, 180]`). Unknown coordinates are represented strictly as `NULL / nil` (stored as SQL `NULL`). The coordinate `(0,0)` is recognized as a valid geographic coordinate near the Gulf of Guinea and is never used as a sentinel for missing location. Connectors cleanly separate publication timestamps (`DetectedAt`) from actual event occurrence times (`OccurredAt`), which remain `NULL/unknown` when absent from source data.
- **Enrichment Pipeline & Authority Hierarchy:** An extensible enrichment layer executes post-normalization and pre-persistence. It obeys a strict data authority hierarchy: `authoritative source data > reliable structured enrichment > approximate enrichment > heuristic inference > unknown`. Higher-authority source properties (such as USGS coordinates) are never overwritten by lower-authority heuristics.
- **Location Confidence Model:** Enrichment categorizes location confidence into `exact` (verified authoritative coordinates within geographic bounds), `approximate` (approximate geolocations), `country` (resolved only to country/national level without exact coordinates), and `unknown` (missing or unresolved location).
- **Source Context vs. Event Location:** Reporting context (such as GDELT's `sourcecountry`) represents the location of the media outlet, not the event location. It is recorded in provenance metadata and is never conflated with the event's geographical occurrence.
- **Idempotency & Non-Fatal Execution:** All enrichers (`GeoEnricher`, `MetadataEnricher`, `ProvenanceEnricher`) are strictly idempotent, deterministic in output ordering, non-blocking, and never fabricate missing event timestamps with `time.Now()`.
- **Reliability & Observability:** The ingestion scheduler runs connectors in isolated, panic-recovered goroutines with overlap protection. The HTTP client implements exponential backoff with jitter and respects HTTP 429 `Retry-After` headers. Operations emit structured logs and provide real-time telemetry (successes, failures, duplication rates) via a `/health/ingestion` JSON endpoint.

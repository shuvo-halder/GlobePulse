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

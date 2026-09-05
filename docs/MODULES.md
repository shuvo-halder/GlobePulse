# Business Modules

This document lists the major business modules within the GlobePulse AI project. Most modules are currently planned but stubbed.

## 1. Authentication & Identity
- **Purpose:** Manage user accounts, sessions, and roles.
- **Main Service:** `auth-service`
- **Database Entities:** `users`, `audit_logs`
- **Current Status:** **Implemented**. Supports registration, login, JWT issuance, and Redis-backed session invalidation.
- **Business Rules:** Passwords must be hashed via bcrypt; JWTs are validated alongside active Redis sessions.

## 2. AI Threat Analysis
- **Purpose:** Analyze geopolitical data and news to assess threat levels.
- **Main Service:** `ai-service`
- **Current Status:** **Stubbed**. The service is scaffolded using FastAPI and Celery. A Celery worker is configured to consume tasks from a RabbitMQ queue (`ai_analysis_queue`), but no actual AI models, prompts, or logic are implemented.

## 3. Global News Aggregation
- **Purpose:** Collect and normalize global news and events.
- **Main Service:** `news-service`
- **Current Status:** **Stubbed**. Exposes a `/health` endpoint only.

## 4. Country Intelligence
- **Purpose:** Maintain geographical, political, and statistical data for specific regions.
- **Main Service:** `country-service`
- **Current Status:** **Stubbed**. Exposes a `/health` endpoint only.

## 5. Analytics & Dashboard Data
- **Purpose:** Aggregate intelligence data to serve to the frontend visualizations.
- **Main Service:** `analytics-service`
- **Current Status:** **Stubbed**. Exposes a `/health` endpoint only.

## 6. Frontend Visualization (Globe Dashboard)
- **Purpose:** Render a 3D globe visualizing active threat events and intelligence metrics.
- **Main Location:** `src/App.tsx`, `src/components/GlobeView.tsx`
- **Current Status:** **Partially Implemented (Mocked)**. The UI is complete and interactive but is powered entirely by static mock data rather than connecting to the `analytics-service` or `news-service`.

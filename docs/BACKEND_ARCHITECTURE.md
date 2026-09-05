# Backend Architecture

This document describes the actual implemented backend architecture of the GlobePulse AI project.

## Architecture Type
The backend follows a **Microservices Architecture**. Services are containerized and orchestrated via Docker Compose, communicating over HTTP (REST) and potentially AMQP (RabbitMQ) for asynchronous tasks, though message broker integration is not yet fully implemented across all services.

## Services Overview

### 1. Auth Service (`services/auth-service`)
**Status:** Implemented
- **Language:** Go 1.21
- **Responsibility:** Handles user registration, login, profile management, and session management.
- **Database:** PostgreSQL (persistent user data, audit logs), Redis (session tokens).
- **Architecture:** Clean Architecture (Domain -> Repository -> Service -> HTTP Handler).

### 2. AI Service (`services/ai-service`)
**Status:** Partially Implemented (Scaffolded)
- **Language:** Python 3.11
- **Framework:** FastAPI, Celery
- **Responsibility:** Intended for asynchronous AI analysis tasks.
- **Current State:** Exposes a basic `/health` endpoint. Celery worker is configured to connect to RabbitMQ and Redis, but no actual AI processing logic is currently implemented.

### 3. News Service (`services/news-service`)
**Status:** Stubbed
- **Language:** Go 1.21
- **Responsibility:** Intended to fetch, aggregate, or serve global news events.
- **Current State:** Only exposes a `/health` endpoint. No database connections or domain logic.

### 4. Country Service (`services/country-service`)
**Status:** Stubbed
- **Language:** Go 1.21
- **Responsibility:** Intended to manage country-specific metadata or statistics.
- **Current State:** Only exposes a `/health` endpoint. No database connections or domain logic.

### 5. Analytics Service (`services/analytics-service`)
**Status:** Stubbed
- **Language:** Go 1.21
- **Responsibility:** Intended to provide data aggregation and analytics for the frontend dashboard.
- **Current State:** Only exposes a `/health` endpoint. No database connections or domain logic.

## Request Lifecycle (Implemented Example: Auth Service)

The actual request lifecycle for the `auth-service` operates as follows:

1. **HTTP Request:** Client sends a REST request to `/api/v1/auth/*`.
2. **Router (`router.go`):** Gin router intercepts the request.
3. **Middleware (`auth_middleware.go`):** If the route is protected, the JWT middleware verifies the `Authorization` header and checks Redis for session validity.
4. **Controller (`auth_handler.go`):** Binds the incoming JSON to a DTO (Data Transfer Object) and validates it.
5. **Service (`auth_service.go`):** Orchestrates business rules (e.g., password hashing, JWT generation).
6. **Repository (`user_repo.go`, `session_repo.go`):** Interacts with PostgreSQL or Redis using `sqlx` and `go-redis`.
7. **Response:** A JSON response is returned to the client.

## Shared Packages

To ensure consistency across the Go microservices, the project uses a `packages/` directory containing local Go modules:
- `shared-auth`: Intended for shared authentication utilities.
- `shared-types`: Intended for shared domain models and interfaces.

*Note: These packages are currently only imported by the `auth-service` via `replace` directives.*

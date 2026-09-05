# Deployment Architecture

This document describes how the GlobePulse AI application is orchestrated and deployed based on the repository configurations.

## Development & Orchestration
The primary deployment methodology uses **Docker Compose** (`docker-compose.yml`) which orchestrates all backend microservices, data stores, and the frontend.

### Infrastructure Services
- **Postgres:** `postgres:15-alpine` - Exposes `5432`. Uses a named volume `pgdata` for persistent storage.
- **Redis:** `redis:7-alpine` - Exposes `6379`.
- **RabbitMQ:** `rabbitmq:3-management-alpine` - Exposes `5672` (AMQP) and `15672` (Management UI).

### Application Services
- **Auth Service (`auth-service`):** Built from `services/auth-service/Dockerfile` (Go 1.21 multi-stage build). Exposes `8081`.
- **News Service (`news-service`):** Built from `services/news-service/Dockerfile`. Exposes `8080`.
- **Country Service (`country-service`):** Built from `services/country-service/Dockerfile`. Exposes `8082`.
- **Analytics Service (`analytics-service`):** Built from `services/analytics-service/Dockerfile`. Exposes `8084`.
- **AI Service (`ai-service`):** Built from `services/ai-service/Dockerfile` (Python 3.11). Exposes `8083` (Uvicorn).
- **AI Worker (`ai-worker`):** Built from the *same* `ai-service` Dockerfile, but overrides the command to run a Celery worker: `celery -A app.core.celery_app worker --loglevel=info -Q ai_analysis_queue`.

### Frontend Service
The React application is built via a multi-stage Dockerfile located at the repository root (`Dockerfile`).
- **Build Stage:** `node:20-alpine`. Installs dependencies and runs `npm run build`.
- **Serve Stage:** `nginx:alpine`. Copies the built static files (`dist/`) into `/usr/share/nginx/html`.
- **Execution:** Runs standard Nginx exposing port `80`. Mapped to port `3000` on the host machine.

## Build Process
A helper script `recreate.sh` is provided in the root to facilitate rebuilding and restarting the entire Docker Compose stack.

## Missing Elements
- There are no Kubernetes manifests (`.yaml`), Helm charts, or CI/CD pipelines (e.g., GitHub Actions, GitLab CI) configured in the repository.
- SSL/TLS termination is not configured inside the Nginx container, assuming this is handled by an upstream reverse proxy or the cloud provider (e.g., Google Cloud Run).
- The database migrations are included in the container images but there is no explicit init script or automation to run the migrations against the PostgreSQL container upon startup.

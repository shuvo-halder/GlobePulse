# Environment Variables & Configuration

This document lists the critical environment variables used to configure the services across the deployment.

## Global Environment Variables
These variables are defined in `.env.example` and are typically injected by the host environment or AI Studio platform.

| Variable | Purpose | Required |
|----------|---------|----------|
| `GEMINI_API_KEY` | Authentication key for Google Gemini AI integrations. | Planned |
| `APP_URL` | The public-facing URL where the application is hosted. | Planned |

## Backend Services Configuration
These variables are passed to the individual microservices via `docker-compose.yml`.

| Variable | Purpose | Required | Example/Default |
|----------|---------|----------|-----------------|
| `APP_ENV` | Application environment (e.g., `development`, `production`). | Yes | `development` |
| `PORT` | The internal port the HTTP server binds to. | Yes | `8081` |
| `DB_HOST` | PostgreSQL hostname. | Yes | `postgres` |
| `DB_PORT` | PostgreSQL port. | Yes | `5432` |
| `DB_USER` | PostgreSQL username. | Yes | `REDACTED` |
| `DB_PASS` | PostgreSQL password. | Yes | `REDACTED` |
| `DB_NAME` | PostgreSQL database name. | Yes | `globepulse` |
| `REDIS_ADDR` | Redis hostname and port (or just hostname for Python). | Yes | `redis:6379` |
| `REDIS_PORT` | Redis port (used by Python). | Yes | `6379` |
| `REDIS_PASS` | Redis password (often empty in dev). | No | `""` |
| `JWT_SECRET` | Secret key used to sign and verify JWT tokens. | Yes | `REDACTED` |
| `RABBITMQ_URL` | Full AMQP connection string for RabbitMQ. | Yes | `amqp://guest:guest@rabbitmq:5672/` |

## Frontend Configuration
The frontend uses Vite, which supports `VITE_` prefixed environment variables. However, no specific environment variables are currently being consumed in the React application code (`src/`).

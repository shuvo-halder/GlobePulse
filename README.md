# GlobePulse AI

## Overview
GlobePulse AI is a global threat intelligence platform. Currently, the repository contains a highly polished 3D interactive dashboard (Frontend) and a structural microservices backbone (Backend). The application visualizes simulated geopolitical events on a 3D globe and provides a fully functional authentication service.

## Architecture
The project utilizes a hybrid microservices architecture orchestrated via Docker Compose.
- **Frontend:** React SPA (Vite)
- **Backend:** Independent Go services, Python AI service, PostgreSQL, Redis, RabbitMQ.

For an in-depth architectural breakdown, please read the [Master Architecture Document](docs/ARCHITECTURE.md).

## Tech Stack
| Domain | Technology |
|---|---|
| **Frontend** | React 19, TypeScript, Vite, Tailwind CSS v4, react-globe.gl, Recharts, Lucide |
| **Backend (Auth)** | Go 1.21, Gin, sqlx, bcrypt, JWT |
| **Backend (AI)** | Python 3.11, FastAPI, Celery |
| **Database** | PostgreSQL 15 |
| **Caching/Session** | Redis 7 |
| **Messaging** | RabbitMQ 3 |

## Project Structure
A detailed breakdown of the codebase can be found in [Project Structure](docs/PROJECT_STRUCTURE.md).

## Main Modules
- **Frontend Dashboard:** 3D visualization and real-time alerts UI (Mocked Data).
- **Auth Service:** Complete user registration, login, and JWT/Redis session management.
- **AI Service:** Scaffolded Celery worker and FastAPI application (AI logic pending).
- **News, Country, Analytics Services:** Currently stubbed (`/health` only).

## Requirements
To run this project locally, you need:
- Docker and Docker Compose (v2)
- Node.js (v20+)
- Go (1.21+) (for local backend development)

## Installation & Development
1. **Clone the repository:**
   ```bash
   git clone <repo-url>
   cd globepulse
   ```
2. **Install frontend dependencies:**
   ```bash
   npm install
   ```
3. **Start the frontend development server:**
   ```bash
   npm run dev
   ```

## Environment Configuration
Copy `.env.example` to `.env` if needed by the frontend. The backend environment variables are injected automatically via `docker-compose.yml`. No secrets should be committed to version control. Read more in [Configuration](docs/CONFIGURATION.md).

## Database Setup
The PostgreSQL database is spun up automatically via Docker Compose. Schema migrations for the `auth-service` are located in `services/auth-service/migrations/`.
*Note: Currently, you must apply the `.sql` schema manually to the Postgres container as there is no automated init script.*

## Build
To build the frontend statically:
```bash
npm run build
```

## Deployment
The entire stack can be launched using Docker Compose. A helper script is provided:
```bash
./recreate.sh
```
This will build all service images, start PostgreSQL, Redis, RabbitMQ, all Go/Python microservices, and serve the frontend via Nginx on port `3000`.

For architectural details, please see [Deployment Architecture](docs/DEPLOYMENT.md).

## API
Detailed API endpoints (specifically for the `auth-service`) are documented in [API.md](docs/API.md).

## Authentication
Authentication uses bcrypt-hashed passwords and stateless JWTs paired with stateful Redis sessions. Read the deep dive in [Authentication & Authorization](docs/AUTHENTICATION_AUTHORIZATION.md).

## Documentation
All reverse-engineered architecture documentation is available in the `docs/` directory:
- [Documentation Index](docs/README.md)

## Testing
There are currently no tests implemented in the repository. Please review [Testing Architecture](docs/TESTING.md) for recommendations.

## Troubleshooting
- **Frontend not loading data:** The frontend is currently hardcoded to use mock data in `App.tsx` and does not connect to the backend services.
- **Service failing to connect to DB:** Ensure the PostgreSQL container is fully healthy before the Go services start.
- **Celery Worker Idle:** The Python `ai-worker` runs successfully but does not currently consume any tasks as no producer is implemented.

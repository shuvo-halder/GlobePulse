# Project Structure

This document outlines the directory and code structure of the GlobePulse AI project as it exists in the repository.

## Root Structure

```text
/
├── assets/                  # Static assets for the frontend
├── docs/                    # Architecture and project documentation
├── go/                      # Go workspace configuration
├── packages/                # Shared Go packages
│   ├── shared-auth/         # Shared authentication Go logic
│   └── shared-types/        # Shared Go types and interfaces
├── services/                # Backend microservices
│   ├── ai-service/          # Python/FastAPI service for AI tasks
│   ├── analytics-service/   # Go service (currently stubbed)
│   ├── auth-service/        # Go service for user auth & identity
│   ├── country-service/     # Go service (currently stubbed)
│   └── news-service/        # Go service (currently stubbed)
├── src/                     # React frontend source code
│   ├── components/          # Reusable React components (e.g., GlobeView)
│   ├── lib/                 # Frontend utilities and mock data
│   ├── App.tsx              # Main React application component
│   ├── main.tsx             # React DOM entry point
│   └── index.css            # Tailwind CSS entry point
├── Dockerfile               # Dockerfile for the frontend (Nginx)
├── docker-compose.yml       # Docker Compose configuration for all services
├── package.json             # Node.js dependencies for frontend
├── tsconfig.json            # TypeScript configuration
└── vite.config.ts           # Vite bundler configuration
```

## Important Directories

### `services/auth-service`
The most mature backend service in the repository. It implements a layered clean architecture:
- `cmd/api/`: Application entry point (`main.go`).
- `internal/config/`: Configuration loading.
- `internal/domain/`: Core business models and interfaces (User, Auth, Audit).
- `internal/handler/http/`: HTTP routing, controllers, and middleware.
- `internal/repository/`: Data access implementations (PostgreSQL, Redis).
- `internal/service/`: Business logic implementations.
- `migrations/`: Database schema definitions (`.sql` files).
- `pkg/`: Reusable utilities (JWT, password hashing, logging).

### `src/`
The frontend React application.
- Uses Vite, React 19, Tailwind CSS v4.
- Contains the main application dashboard (`App.tsx`) rendering a 3D globe using `react-globe.gl`.
- Uses static mock data (`lib/data.ts`) rather than connecting to the backend services.

### `packages/`
Contains shared Go modules intended to be used across multiple Go microservices to maintain type safety and avoid code duplication. Currently referenced primarily by the `auth-service` via `replace` directives in `go.mod`.

## Configuration Files

- `docker-compose.yml`: Defines the orchestration for all microservices, databases (PostgreSQL, Redis), message broker (RabbitMQ), and frontend.
- `package.json`: Manages frontend dependencies and scripts (e.g., `dev`, `build`, `lint`).
- `.env.example`: Template for required environment variables (e.g., `GEMINI_API_KEY`).
- `go.work` / `go.work.sum`: Go workspace configuration to manage multi-module local development.

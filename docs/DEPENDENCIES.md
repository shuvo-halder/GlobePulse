# Dependencies Audit

This document lists the primary dependencies utilized across the repository.

## Frontend (`package.json`)

### UI Framework & Libraries
- `react` / `react-dom` (^19.0.1)
- `lucide-react` (^1.27.0): Iconography.
- `motion` (^12.43.0): Animation library.
- `react-globe.gl` (^2.38.0) & `three` (^0.185.1): Core 3D visualization components.
- `recharts` (^3.10.1): Charting library.
- `clsx` (^2.1.1) & `tailwind-merge` (^3.6.0): CSS class utility merging.

### Build & Styling
- `vite` (^6.2.3): Bundler and dev server.
- `@vitejs/plugin-react` (^5.0.4)
- `tailwindcss` (^4.1.14) & `@tailwindcss/vite` (^4.1.14)
- `typescript` (~5.8.2)

### Third-Party / AI
- `@google/genai` (^2.4.0): Google Gemini SDK (Currently unused in actual implementation).

## Backend: Go Services (`services/auth-service/go.mod`)

### Framework & API
- `github.com/gin-gonic/gin` (v1.9.1): HTTP Web framework.

### Database & Storage
- `github.com/lib/pq` (v1.10.9): PostgreSQL driver.
- `github.com/jmoiron/sqlx` (v1.3.5): SQL extensions (struct scanning).
- `github.com/go-redis/redis/v8` (v8.11.5): Redis client.

### Security & Utilities
- `github.com/golang-jwt/jwt/v5` (v5.0.0): JWT generation and validation.
- `golang.org/x/crypto` (v0.14.0): Provides `bcrypt` for password hashing.
- `github.com/google/uuid` (v1.3.1): UUID generation.
- `github.com/spf13/viper` (v1.16.0): Configuration management.
- `go.uber.org/zap` (v1.25.0): Structured logging.

*Note: Other Go services have no external dependencies defined in their respective `go.mod` files.*

## Backend: Python Service (`services/ai-service/requirements.txt`)

### Framework & Async
- `fastapi` (==0.104.1): Web framework.
- `uvicorn` (==0.24.0.post1): ASGI server.
- `celery` (==5.3.6): Distributed task queue.
- `pydantic` (==2.5.2): Data validation.

## Suspicious / Unused Dependencies
- `express` (^4.21.2) is listed in the frontend `package.json`. Given the application is an SPA served by Nginx and the backend uses Go/Python, Express is likely unused/dead weight.
- `dotenv` (^17.2.3) is in the frontend dependencies, but Vite handles `.env` files natively.
- `@google/genai` is in the frontend `package.json` but is unused. The AI processing is intended to run in the backend Python service.

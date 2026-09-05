# Master Architecture Document

## 1. Project Overview
GlobePulse AI is designed as a global threat intelligence platform. It features a 3D interactive globe dashboard (Frontend) backed by a microservices architecture (Backend). The application intends to aggregate global news, analyze events using AI (Google Gemini), and visualize threat assessments in real-time.

## 2. System Architecture
The repository employs a **hybrid Microservices Architecture** with a decoupled Single Page Application (SPA) frontend.

- **Frontend:** React SPA built with Vite.
- **Backend Services:** Independent microservices written primarily in Go, with Python used for AI processing.
- **Orchestration:** Docker Compose manages all services and infrastructure dependencies.

```mermaid
graph TD
    Client[Browser / Client] --> Frontend[Nginx - React SPA]
    Client --> Auth[Auth Service :8081]
    Client -.-> News[News Service :8080 (Stub)]
    Client -.-> Country[Country Service :8082 (Stub)]
    Client -.-> Analytics[Analytics Service :8084 (Stub)]
    
    Auth --> Postgres[(PostgreSQL)]
    Auth --> Redis[(Redis)]
    
    News -.-> RabbitMQ[RabbitMQ]
    
    RabbitMQ --> AIWorker[Celery AI Worker]
    AIWorker --> AIService[AI Service :8083]
    AIService -.-> Gemini[External: Gemini AI]
```

## 3. Current Implementation Reality
*Crucial Note: The architecture described above represents the structural intent. The current implemented reality is drastically different:*
- Only the **Frontend UI** and the **Auth Service** are fully implemented.
- The Frontend uses **mock data** and does not connect to the backend APIs.
- The `news`, `country`, `analytics`, and `ai` services are **stubs** that only return `200 OK` on `/health`.

## 4. Component Summary

### Frontend (React/Vite)
- A highly polished dashboard using Tailwind CSS, `react-globe.gl` (Three.js), and `recharts`.
- Contains no API integration logic.

### Auth Service (Go)
- **Architecture:** Clean Architecture (Handlers -> Services -> Repositories -> DB).
- **Features:** Registration, Login, Logout, Profile, Password Reset.
- **Security:** bcrypt password hashing, JWT stateless tokens, Redis stateful session validation.

### Database (PostgreSQL)
- Schema contains `users` and `audit_logs` tables.
- No ORM; utilizes `database/sql` and `sqlx`.

## 5. Major Missing Workflows
- **AI Processing:** No code exists connecting the application to the Gemini AI API.
- **Data Ingestion:** No logic exists to fetch or parse news/geopolitical events.
- **Frontend Hydration:** The UI does not fetch real data.

## 6. Deployment
Deployments are handled locally via `docker-compose.yml`.
- The frontend is served statically via Nginx.
- Backend services map their respective internal ports to host ports.

## 7. Known Limitations & Technical Debt
Please see [TECHNICAL_DEBT.md](./TECHNICAL_DEBT.md) for a comprehensive list. The most critical issue is an insecure `/reset-password` endpoint in the `auth-service` that lacks verification logic.

## 8. Directory Index
For more detailed insights into specific architectural domains, refer to the following documents:
- [API Documentation](./API.md)
- [Authentication & Authorization](./AUTHENTICATION_AUTHORIZATION.md)
- [Backend Architecture](./BACKEND_ARCHITECTURE.md)
- [Frontend Architecture](./FRONTEND_ARCHITECTURE.md)
- [Database Architecture](./DATABASE.md)
- [Deployment](./DEPLOYMENT.md)
- [Integrations](./INTEGRATIONS.md)
- [Security](./SECURITY.md)
- [Testing](./TESTING.md)
- [Project Structure](./PROJECT_STRUCTURE.md)

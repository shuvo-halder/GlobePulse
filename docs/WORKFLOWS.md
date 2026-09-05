# Important Business Workflows

This document traces the significant implemented business workflows through the application code.

## 1. Authentication Lifecycle

```mermaid
sequenceDiagram
    participant Client
    participant AuthHandler (API)
    participant AuthService (Logic)
    participant Postgres (Users)
    participant Redis (Sessions)

    %% Registration
    Client->>AuthHandler (API): POST /api/v1/auth/register
    AuthHandler (API)->>AuthService (Logic): Register(email, password, ...)
    AuthService (Logic)->>AuthService (Logic): Hash password (bcrypt)
    AuthService (Logic)->>Postgres (Users): Insert new user record
    Postgres (Users)-->>AuthService (Logic): Success
    AuthService (Logic)-->>AuthHandler (API): User object
    AuthHandler (API)-->>Client: 201 Created

    %% Login
    Client->>AuthHandler (API): POST /api/v1/auth/login
    AuthHandler (API)->>AuthService (Logic): Login(email, password)
    AuthService (Logic)->>Postgres (Users): Fetch user by email
    Postgres (Users)-->>AuthService (Logic): User record
    AuthService (Logic)->>AuthService (Logic): Verify bcrypt password
    AuthService (Logic)->>Redis (Sessions): Create session_id with expiry
    AuthService (Logic)->>AuthService (Logic): Generate JWT (user_id, session_id)
    AuthService (Logic)-->>AuthHandler (API): Token, Session ID
    AuthHandler (API)-->>Client: 200 OK (Token)

    %% Protected Request
    Client->>AuthHandler (API): GET /api/v1/auth/profile (Bearer Token)
    AuthHandler (API)->>AuthHandler (API): AuthMiddleware Validates JWT Signature
    AuthHandler (API)->>Redis (Sessions): Verify session_id exists
    Redis (Sessions)-->>AuthHandler (API): Session valid
    AuthHandler (API)->>AuthService (Logic): GetProfile(user_id)
    AuthService (Logic)->>Postgres (Users): Fetch user by ID
    Postgres (Users)-->>AuthService (Logic): User record
    AuthService (Logic)-->>AuthHandler (API): User object
    AuthHandler (API)-->>Client: 200 OK (User object)

    %% Logout
    Client->>AuthHandler (API): POST /api/v1/auth/logout
    AuthHandler (API)->>AuthService (Logic): Logout(session_id)
    AuthService (Logic)->>Redis (Sessions): Delete session_id key
    Redis (Sessions)-->>AuthService (Logic): Success
    AuthService (Logic)-->>AuthHandler (API): Success
    AuthHandler (API)-->>Client: 200 OK
```

## 2. Inferred / Planned Workflows

### AI Threat Analysis
Based on the configuration in the `ai-service` and `docker-compose.yml`, the intended workflow is:
1. A service (e.g., `news-service`) publishes an event to RabbitMQ.
2. The `ai-worker` (Celery) consumes the message from the `ai_analysis_queue`.
3. The Python code runs AI threat assessment (likely using the Gemini API based on `.env.example`).
4. Results are stored in the database or published back to RabbitMQ.
*Note: This flow is not actually implemented yet.*

### Frontend Data Hydration
The frontend (`App.tsx`) is intended to fetch live data from the backend services (presumably `analytics-service` or `news-service`), but it currently uses static arrays (`mockEvents`, `mockChartData`).

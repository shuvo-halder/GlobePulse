# Technical Debt & Architectural Issues

This document lists the technical debt, missing implementations, and architectural inconsistencies identified during the codebase audit.

## Critical Issues

### 1. Insecure Password Reset (Authentication Vulnerability)
- **Description:** The `/api/v1/auth/reset-password` endpoint updates a user's password based solely on the provided email and new password. There is no verification mechanism (e.g., OTP, email link, or requiring the old password).
- **Consequence:** Anyone can overwrite any user's password if they know their email.

### 2. Disconnected Frontend
- **Description:** The React frontend (`App.tsx`) is entirely driven by static mock data (`mockEvents`, `mockChartData`). It makes zero API calls to the backend services.
- **Consequence:** The application appears functional visually, but acts purely as a UI prototype.

### 3. Stubbed Microservices
- **Description:** `news-service`, `country-service`, and `analytics-service` exist merely as structural folders with a `/health` endpoint. They have no domain logic, database connections, or API definitions.
- **Consequence:** The intended microservice architecture is largely vaporware beyond the `auth-service`.

## High Issues

### 4. Missing Complete AI Implementation
- **Description:** The `ai-service` configures a Celery worker, but contains no actual AI processing logic, prompt construction, or integration with the Gemini SDK. The frontend also installs the `@google/genai` SDK unnecessarily.
- **Consequence:** The core value proposition of the app ("AI Threat Intelligence") is missing.

### 5. No Testing Infrastructure
- **Description:** Zero unit, integration, or E2E tests exist in the repository across all languages.
- **Consequence:** Refactoring or extending the codebase carries high risk.

## Medium Issues

### 6. Unenforced RBAC (Role-Based Access Control)
- **Description:** While a `role` field exists in the DB and JWT, there is no middleware actually checking these roles to restrict access to endpoints.

### 7. Unused Dependencies
- **Description:** The frontend `package.json` includes `express` and `dotenv`, which are unused in a Vite-based SPA deployed via Nginx.

### 8. Database Migrations Not Automated
- **Description:** While `.sql` migration files exist for the `auth-service`, there is no automated mechanism (e.g., init container, Flyway, golang-migrate script) running these migrations against the PostgreSQL container on startup.

## Informational Issues

### 9. Misaligned Responsibilities
- **Description:** The `shared-packages` architecture in Go is a known anti-pattern for microservices if it creates tight coupling. Currently, it is only used by the `auth-service`, but scaling this to other services may introduce versioning headaches.
- **Description:** The frontend includes logic and state for features that should likely be managed by the backend (e.g., computing threat levels).

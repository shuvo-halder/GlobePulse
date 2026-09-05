# Security Architecture

This document outlines the security mechanisms and potential vulnerabilities identified in the codebase. This is a passive audit based on the existing repository state.

## Authentication & Authorization
- **Implementation:** Custom stateless JWT paired with Redis-backed session checking.
- **Passwords:** Hashed securely using `bcrypt` (`golang.org/x/crypto/bcrypt`).
- **Token Management:** The pairing of JWTs with a stateful Redis store effectively solves the "JWT invalidation" problem, allowing for immediate session termination on logout.
- **RBAC:** A `role` column exists in the database and is embedded in the JWT. However, there are no middleware or route guards enforcing role-based access.

## Data Security & Input Validation
- **Input Validation:** The Go `auth-service` uses Gin's `ShouldBindJSON`, which provides basic struct binding validation. There are no advanced sanitization rules applied to inputs like `first_name` or `last_name`.
- **SQL Injection:** The `auth-service` uses standard parameterized queries via `sqlx` (e.g., `db.GetContext`, `db.ExecContext`), which protects against SQL injection.
- **CSRF Protection:** No explicit CSRF protection (e.g., CSRF tokens) is implemented. The API assumes clients will pass the token via the `Authorization` header, which is generally immune to traditional CSRF compared to cookie-based auth, provided the token is not stored in an insecure location like `localStorage` where XSS could steal it.

## Network & Configuration Security
- **CORS:** There is no Cross-Origin Resource Sharing (CORS) middleware configured in the `auth-service` or frontend Nginx configuration. This may cause issues if the frontend and backend are hosted on different domains/ports.
- **Secrets Management:** Secrets like `JWT_SECRET` and database passwords are passed via environment variables, which is standard practice.
- **HTTP Security Headers:** The Nginx frontend container and the Go backend services do not inject security headers (e.g., `Strict-Transport-Security`, `Content-Security-Policy`, `X-Frame-Options`).

## Vulnerabilities & Areas for Improvement
- **Reset Password Endpoint:** The `/api/v1/auth/reset-password` endpoint does not require an old password, OTP, or email verification token. Anyone who knows a user's email can reset their password, which is a **critical vulnerability**.
- **Audit Logs:** While an `audit_logs` table exists, there is no strict Foreign Key constraint enforcing referential integrity with the `users` table. Furthermore, there is no code actively inserting records into this table within the `auth-service`.

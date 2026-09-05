# Authentication & Authorization

This document outlines the authentication and authorization mechanisms implemented within the codebase. The logic resides exclusively within the `auth-service`.

## Authentication Flow

The application uses a **stateless JWT** paired with **stateful Redis sessions** for authentication.

1. **Login:**
   - User submits email and password.
   - `auth-service` retrieves the user record from PostgreSQL.
   - The password is verified against a bcrypt hash (`golang.org/x/crypto/bcrypt`).
   - A unique `session_id` is generated using UUID.
   - The session is stored in Redis (with an expiration time).
   - A JWT is generated containing the `user_id`, `role`, and `session_id`.
   - The JWT and session ID are returned to the client.

2. **Middleware Verification (`AuthMiddleware`):**
   - Extracts the Bearer token from the `Authorization` header.
   - Validates the JWT signature using the configured `JWT_SECRET`.
   - Extracts the `session_id` from the JWT claims.
   - Checks Redis to ensure the session still exists and is valid.
   - If valid, injects `user_id`, `role`, and `session_id` into the Gin context for downstream handlers.

3. **Logout:**
   - The user calls the protected logout endpoint.
   - The `auth-service` deletes the corresponding `session_id` from Redis, immediately invalidating the token.

## Authorization (RBAC)

- The PostgreSQL `users` table includes a `role` column (default: `'user'`).
- The role is injected into the JWT claims and subsequently into the Gin request context by the middleware.
- **Status:** Partially Implemented. While the infrastructure for roles exists (database schema, JWT claims, context injection), there is currently **no middleware or logic enforcing role-based access control (e.g., Admin-only routes)**.

## Security Observations

- **Password Hashing:** Passwords are securely hashed using bcrypt (`pkg/hash/password.go`).
- **Token Invalidation:** Because JWTs are paired with a Redis session check, tokens can be actively revoked upon logout.
- **Missing Features:** 
  - There is no email verification flow, though an `is_verified` boolean exists in the database schema.
  - The `/reset-password` endpoint allows resetting a password simply by providing the email and a new password, lacking a secure challenge (OTP or reset link).

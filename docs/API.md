# API Documentation

This document describes the REST API endpoints currently implemented in the backend microservices. 

> **Note:** Only the `auth-service` has functional business endpoints. All other services currently only expose a `/health` endpoint.

## Auth Service (`:8081`)

### 1. Register User
- **Method:** `POST`
- **Endpoint:** `/api/v1/auth/register`
- **Auth Required:** Public
- **Purpose:** Create a new user account.
- **Request Body (JSON):**
  - `email` (string, required): User's email address.
  - `password` (string, required): User's password (min 6 chars).
  - `first_name` (string, required): User's first name.
  - `last_name` (string, required): User's last name.
- **Success Response:** `201 Created` returning the user object.
- **Error Responses:** `400 Bad Request` (validation failed), `500 Internal Server Error`.

### 2. Login
- **Method:** `POST`
- **Endpoint:** `/api/v1/auth/login`
- **Auth Required:** Public
- **Purpose:** Authenticate a user and generate a session token.
- **Request Body (JSON):**
  - `email` (string, required): User's email address.
  - `password` (string, required): User's password.
- **Success Response:** `200 OK`
  ```json
  {
    "token": "eyJhbG...",
    "session_id": "uuid-string"
  }
  ```
- **Error Responses:** `400 Bad Request` (validation failed), `401 Unauthorized` (invalid credentials).

### 3. Reset Password
- **Method:** `POST`
- **Endpoint:** `/api/v1/auth/reset-password`
- **Auth Required:** Public
- **Purpose:** Update a user's password.
- **Request Body (JSON):**
  - `email` (string, required): User's email.
  - `new_password` (string, required): The new password to set.
- **Success Response:** `200 OK`
- **Error Responses:** `400 Bad Request`, `500 Internal Server Error`.
- *(Note: In a production system, this endpoint lacks an email verification or OTP check before resetting the password.)*

### 4. Get Profile
- **Method:** `GET`
- **Endpoint:** `/api/v1/auth/profile`
- **Auth Required:** Yes (Bearer Token)
- **Purpose:** Retrieve the currently authenticated user's profile.
- **Headers:** `Authorization: Bearer <token>`
- **Success Response:** `200 OK` returning the user object.
- **Error Responses:** `401 Unauthorized`, `404 Not Found`.

### 5. Logout
- **Method:** `POST`
- **Endpoint:** `/api/v1/auth/logout`
- **Auth Required:** Yes (Bearer Token)
- **Purpose:** Invalidate the current session.
- **Headers:** `Authorization: Bearer <token>`
- **Success Response:** `200 OK`
- **Error Responses:** `401 Unauthorized`, `500 Internal Server Error`.

---

## Health Endpoints

The following endpoints are implemented across the microservices to facilitate container health checks:

| Service | Endpoint | Port | Response |
|---------|----------|------|----------|
| Auth | `GET /health` | (Not explicitly defined in router, inferred standard) | N/A |
| News | `GET /health` | `8080` | `200 OK` `{"status": "ok"}` |
| Country | `GET /health` | `8082` | `200 OK` `{"status": "ok"}` |
| Analytics | `GET /health` | `8084` | `200 OK` `{"status": "ok"}` |
| AI | `GET /health` | `8083` | `200 OK` `{"status": "ok"}` |

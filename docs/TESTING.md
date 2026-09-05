# Testing Architecture

This document describes the testing approach within the repository.

## Current State

**Testing is completely missing from the repository.**

- **Backend (Go/Python):** 
  - There are no `_test.go` files anywhere in the `services/` or `packages/` directories.
  - There are no test scripts defined in the Python `ai-service`.
- **Frontend (React):** 
  - There are no test files (e.g., `.test.ts`, `.spec.tsx`) in the `src/` directory.
  - No testing libraries (e.g., Vitest, Jest, React Testing Library) are included in the `package.json` dependencies.
- **E2E/Integration:** 
  - There are no End-to-End testing frameworks (e.g., Cypress, Playwright) configured.

## Missing Components
To establish a healthy testing architecture, the following should be implemented:
1. **Go Unit Tests:** Table-driven tests for services, repositories, and handlers (e.g., using `testify`).
2. **Python Unit Tests:** Pytest setup for the FastAPI/Celery service.
3. **Frontend Tests:** Vitest + React Testing Library for component unit testing.
4. **Integration Tests:** Dockerized database testing to verify repository interactions.

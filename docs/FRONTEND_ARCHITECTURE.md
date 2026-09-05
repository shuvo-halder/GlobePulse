# Frontend Architecture

This document describes the implemented frontend architecture of the GlobePulse AI project.

## Tech Stack
- **Framework:** React 19
- **Build Tool:** Vite
- **Language:** TypeScript
- **Styling:** Tailwind CSS v4
- **Visualization:** `react-globe.gl`, `recharts`
- **Icons:** `lucide-react`
- **Animation:** `motion`

## Architecture Type
The frontend is a **Single Page Application (SPA)** that acts as a global intelligence dashboard. It is served natively via Nginx when deployed using the provided Dockerfile.

## Code Structure

The source code resides in the `src/` directory:

- `main.tsx`: The React DOM entry point.
- `App.tsx`: The primary layout and composition component. It manages the global state for the dashboard and renders the primary UI shell.
- `index.css`: Global stylesheet injecting Tailwind CSS.
- `components/`: Contains specialized UI components.
  - `GlobeView.tsx`: Wrapper for `react-globe.gl` handling the 3D globe rendering and interaction.
- `lib/`: Contains utilities and data.
  - `data.ts`: Contains hardcoded mock data (`mockEvents`) currently used to populate the UI.

## Current Implementation Status

### User Interface
The UI is well-developed, featuring a dark-themed, glassmorphic layout overlaid on a 3D interactive globe. It includes:
- A top navigation bar with a search and filter mock interface.
- A side panel displaying a feed of real-time alerts and events.
- A right-side analytics panel showing AI threat assessments and metric charts (using `recharts`).
- A bottom scrolling ticker for latest alerts.

### Data Flow & API Integration
**Status: Not Implemented / Mocked**
- The frontend currently relies entirely on static mock data defined in `src/lib/data.ts` and hardcoded inside `App.tsx` (e.g., `mockChartData`).
- There are no active API clients (e.g., `fetch`, `axios`) communicating with the backend microservices.
- There is no client-side routing (e.g., `react-router-dom`); the entire application exists on a single view.
- State management relies purely on local React `useState` hooks within `App.tsx`.

## Deployment
The frontend is built statically via `npm run build` and served using an Nginx Alpine container, as defined in the root `Dockerfile`.

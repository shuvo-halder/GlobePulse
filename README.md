# GlobePulse AI

A next-generation global news intelligence platform that leverages artificial intelligence to provide real-time, spatial, and analytical insights into world events.

## Overview

GlobePulse AI reimagines how we consume and understand international news. Instead of a flat list of headlines, GlobePulse AI visualizes the world's information pulse on an interactive 3D globe. It is a comprehensive intelligence platform designed to track, analyze, and synthesize global events in real-time.

Traditional news portals present fragmented information. GlobePulse AI unifies data streams from around the world, applying advanced Natural Language Processing (NLP) and machine learning models to instantly surface sentiment, extract key topics, and highlight breaking global events. It transforms raw news into actionable intelligence.

## Features

GlobePulse AI offers a comprehensive suite of tools for global monitoring:

- **Interactive 3D Globe**: Navigate world news spatially. Spin the globe and click on any region to instantly access localized intelligence.
- **Live News Tracking**: Continuous, real-time ingestion of global news sources, ensuring you never miss a critical update.
- **Country Based News**: Deep dive into specific nations, exploring their unique news landscape, trending topics, and overall sentiment.
- **Global News Heatmap**: Visually identify hotspots of global activity, crisis zones, and trending regions at a glance.
- **Breaking News Feed**: A dedicated, low-latency stream of the most critical and impactful global events.
- **AI Summaries**: Instantly grasp the core message of complex articles and evolving situations with AI-generated concise summaries.
- **Sentiment Analysis**: Understand the emotional tone of news coverage, categorized by country, topic, or specific event.
- **Topic Detection**: Automatic categorization and clustering of news articles into relevant global themes (e.g., Geopolitics, Economics, Technology).
- **Real-Time Analytics**: Dashboards providing live metrics on global news volume, threat assessments, and regional stability indicators.
- **Search & Filtering**: Powerful, granular search capabilities across the entire global news database.
- **Watchlists**: Curate personalized lists of countries, topics, or ongoing events for focused monitoring.
- **Bookmarks**: Save critical articles, analyses, and intelligence reports for later review.
- **Notifications**: Configurable real-time alerts for breaking news in your watched regions or topics.

## How It Works

1. **User opens the platform**: You are immediately presented with a high-level overview of global activity via the interactive 3D globe and live ticker.
2. **User interacts with the globe**: Rotate, zoom, and explore the planet. Heatmap layers reveal areas with high news volume or critical alerts.
3. **User selects a country**: Clicking a nation transitions the interface to a focused dashboard for that specific region.
4. **Related news loads**: The latest articles, categorized by relevance and urgency, are instantly pulled for the selected area.
5. **AI generates insights**: Our backend AI models analyze the incoming stream, providing instant summaries, extracting key entities, and calculating sentiment scores.
6. **Analytics are displayed**: Real-time charts and metrics update, showing activity volume, threat levels, and historical trends for the region.

## User Guide

GlobePulse AI is designed to be intuitive for both casual readers and intensive researchers.

- **View global news**: Start on the homepage to see the big picture. Use the globe to see where news is happening right now.
- **Track a country**: Click on a country on the globe or search for it to view its dedicated intelligence page.
- **Follow breaking events**: Keep an eye on the "Latest Alerts" ticker and the dedicated Breaking News feed for immediate updates.
- **Save bookmarks**: Click the bookmark icon on any article or summary to save it to your personal library for deep reading later.
- **Create watchlists**: Build custom monitors for specific nations or global topics (e.g., "Middle East Energy Markets") to receive curated feeds and alerts.

## Country Explorer

The Country Explorer provides a dedicated intelligence hub for every nation:

- **How country pages work**: Each page aggregates all relevant data for a single nation, providing a unified view of its current state.
- **What information is available**: Access local news feeds, historical data, key political and economic indicators derived from news volume, and regional analytics.
- **Country rankings**: View how a country ranks globally in terms of news volume, positive/negative sentiment, and specific topic mentions.
- **Country sentiment**: A live gauge of the overall media sentiment surrounding the nation, helping identify periods of crisis or stability.
- **Trending topics**: Discover what subjects are currently dominating the national discourse.

## News Intelligence

Our core AI engine provides deep analytical capabilities:

- **AI summaries**: Long-form reports and complex geopolitical events are distilled into instantly readable bullet points.
- **Sentiment scoring**: Every article is evaluated for tone, allowing users to track the emotional trajectory of ongoing events.
- **Entity extraction**: Automatic identification of key people, organizations, and locations involved in the news.
- **Topic classification**: Articles are intelligently sorted into macro-categories, enabling powerful filtering.
- **News clustering**: Similar articles from different sources are grouped together to provide a comprehensive view of an event and reduce noise.

## Real-Time Global Monitoring

The interactive globe is the heart of the platform's monitoring capability:

- **Country activity indicators**: Visual cues on the map highlight nations with surging news volume or critical alerts.
- **Event markers**: Specific, high-impact events are plotted geographically for spatial context.
- **News flow visualization**: See how stories spread and impact different regions simultaneously.
- **Global event tracking**: Monitor ongoing, multi-national events as they unfold across the globe.

## Use Cases

### Journalists
Monitor emerging global stories, track sentiment shifts in regions of interest, and quickly understand complex situations using AI summaries to accelerate reporting.

### Researchers
Analyze historical news trends, track the media presence of specific entities over time, and gather comprehensive data on global events for academic or institutional studies.

### Students
Visualize global affairs, understand the geographical context of international relations, and stay informed on world events with easily digestible summaries.

### Investors
Track geopolitical risks, monitor economic sentiment across different markets, and receive real-time alerts on events that could impact global supply chains or financial stability.

### Government Agencies
Maintain high-level situational awareness of global hotspots, monitor international sentiment regarding specific policies, and track emerging regional crises.

### Security Analysts
Identify elevated threat levels through sentiment and topic analysis, monitor regions for instability, and track the flow of information during critical events.

### General Readers
Escape the echo chamber. Explore news spatially, discover what is happening outside your immediate region, and gain a broader, more objective understanding of the world.

## Screens and Pages

- **Homepage**: The central hub featuring the 3D globe, live threat assessments, and the global news ticker.
- **Globe View**: A fullscreen, interactive spatial visualization of global news density and events.
- **Country Page**: A detailed dashboard focusing on the news, sentiment, and analytics of a specific nation.
- **Search Page**: A powerful, faceted search interface to query the entire historical news database.
- **Analytics Dashboard**: Comprehensive charts and metrics tracking global trends, sentiment shifts, and topic popularity over time.
- **User Profile**: Manage your account details, API keys, and subscription preferences.
- **Settings**: Configure notification preferences, manage watchlists, and customize the platform interface.

## Developer Guide

GlobePulse AI is built on a modern, microservices-based architecture for scalability and resilience.

### Repository Structure

```text
.
├── src/                  # React/Vite frontend application
├── services/             # Backend microservices
│   ├── auth-service/     # Handles authentication and authorization
│   ├── news-service/     # Ingests, processes, and serves news data
│   ├── country-service/  # Manages country metadata and regional data
│   ├── analytics-service/# Processes metrics and aggregates data
│   └── ai-service/       # Python-based NLP and machine learning engine
├── packages/             # Shared libraries and internal tools
├── docker-compose.yml    # Local orchestration
└── package.json          # Frontend dependencies and scripts
```

### Services Overview

- **auth-service**: Written in Go. Manages user registration, login, JWT issuance, and session management using PostgreSQL and Redis.
- **news-service**: Written in Go. The core data ingestion engine. It fetches, sanitizes, and stores news articles, serving them to the frontend via REST APIs.
- **country-service**: Written in Go. Provides geospatial data, country metadata, and handles region-specific queries.
- **analytics-service**: Written in Go. Aggregates data from the news-service to generate time-series metrics, heatmaps, and platform-wide statistics. It consumes events via RabbitMQ.
- **ai-service**: Written in Python (FastAPI/Celery). Handles the heavy lifting for Natural Language Processing. It performs sentiment analysis, entity extraction, summarization, and topic modeling on incoming news articles, communicating results back via RabbitMQ.

## Deployment

### Local Development

To run the entire stack locally for development:

1. Ensure Docker and Docker Compose are installed.
2. Clone the repository.
3. Run `docker-compose up -d` to start the backend services and databases.
4. Run `npm install` and `npm run dev` in the root directory to start the Vite development server.

### Docker Deployment

The `docker-compose.yml` file provides a complete, containerized environment suitable for testing and staging. It spins up all microservices, PostgreSQL, Redis, and RabbitMQ.

### Production Deployment

For production, the microservices are designed to be deployed to a Kubernetes cluster.
- Frontend: Served via CDN (e.g., Cloudflare) and standard static hosting.
- Backend Services: Containerized and orchestrated via Kubernetes (GKE, EKS, etc.) for auto-scaling and high availability.
- Databases: Managed cloud databases (e.g., Cloud SQL for PostgreSQL, Memorystore for Redis) are recommended.
- Message Queue: Managed RabbitMQ or standard message brokers.

## Future Roadmap

- **Multi-lingual Support**: Native translation and analysis of news sources in languages beyond English.
- **Predictive Analytics**: Using historical data to forecast potential geopolitical hotspots or market shifts.
- **Custom Data Ingestion**: Allowing enterprise users to connect their own internal data feeds or specific RSS sources for analysis alongside global news.
- **Advanced Network Graphs**: Visualizing the relationships between extracted entities (people, organizations) to map influence networks.
- **Mobile Application**: Dedicated native iOS and Android apps for on-the-go global monitoring.

## License

Copyright © 2024 GlobePulse AI. All rights reserved.

## Contributing

We welcome contributions! Please see our `CONTRIBUTING.md` file for guidelines on how to submit pull requests, report issues, and suggest features.

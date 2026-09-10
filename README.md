# DataLens — Real-time Data Analytics Platform

A full-stack data analytics platform built to demonstrate proficiency in Java, Go, Elasticsearch, PostgreSQL, Spark, React, Redux, and cloud infrastructure.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        FRONTEND                             │
│  React + TypeScript + Redux + Tailwind + Recharts           │
└──────────────────────┬──────────────────────────────────────┘
                       │ REST
┌──────────────────────▼──────────────────────────────────────┐
│                     API GATEWAY (Java)                       │
│  Spring Boot 3 · JWT Auth · Rate Limiting · CORS            │
└──┬──────────────┬───────────────────┬───────────────────────┘
   │              │                   │
   ▼              ▼                   ▼
┌────────┐  ┌───────────┐  ┌──────────────────┐
│  Go    │  │Elasticsearch│  │   Apache Spark   │
│Ingest  │  │  (Search)   │  │  (Batch Analytics)│
│Service │  │             │  │                  │
└───┬────┘  └─────────────┘  └────────┬─────────┘
    │                                  │
    ▼                                  ▼
┌──────────────────────────────────────────────┐
│              PostgreSQL                       │
│  (Dataset metadata, user accounts)           │
└──────────────────────────────────────────────┘
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | React 18, TypeScript, Redux Toolkit, Tailwind CSS, React Router, Recharts |
| API Gateway | Java 17, Spring Boot 3.2, Gradle, JWT (jjwt), BCrypt |
| Ingestion | Go 1.26, net/http, lib/pq, Elasticsearch client |
| Search | Elasticsearch 8 |
| Storage | PostgreSQL 16 |
| Analytics | Apache Spark (PySpark) |
| Build | Gradle (Java), Vite (Frontend), go build (Go) |
| Testing | JUnit 5, Mockito, Vitest, React Testing Library |

## Prerequisites

- Java 17+
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Python 3.9+ (for Spark jobs)

## Quick Start

### 1. Start Infrastructure

```bash
docker-compose up -d
```

This starts:
- PostgreSQL on port 5432
- Elasticsearch on port 9200

### 2. Start API Gateway (Java)

```bash
cd api-gateway
./gradlew bootRun
```

API runs on port 8080.

### 3. Start Ingestion Service (Go)

```bash
cd ingestion-service
go run .
```

Service runs on port 8081.

### 4. Start Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend runs on port 3000.

## API Endpoints

### Authentication
- `POST /api/auth/register` — Create account
- `POST /api/auth/login` — Get JWT token

### Datasets
- `POST /api/datasets/upload` — Upload CSV/JSON
- `GET /api/datasets` — List user datasets
- `GET /api/datasets/:id` — Get dataset details
- `DELETE /api/datasets/:id` — Delete dataset

### Search
- `POST /search` — Full-text search (Go service, port 8081)

### Health
- `GET /api/health` — Service health check

## Running Tests

### Java Tests
```bash
cd api-gateway
./gradlew test
```

### Frontend Tests
```bash
cd frontend
npm test
npm run test:coverage
```

### Go Tests
```bash
cd ingestion-service
go test ./...
```

### Spark Analytics
```bash
cd spark-jobs
pip install -r requirements.txt
python analytics_job.py /path/to/dataset.csv 1
```

## Project Structure

```
datalens/
├── frontend/              # React + TypeScript + Redux + Tailwind
├── api-gateway/           # Java Spring Boot (Gradle)
├── ingestion-service/     # Go microservice
├── spark-jobs/            # PySpark analytics
├── docker-compose.yml     # PostgreSQL + Elasticsearch
└── README.md
```

## Design Decisions

1. **PostgreSQL over Cassandra** — Simpler for MVP, relational model fits dataset metadata
2. **Go for ingestion** — High-throughput CSV parsing, concurrent ES indexing
3. **Java for API gateway** — Spring Security for JWT, Palantir's primary backend language
4. **Redux Toolkit** — Predictable state for complex dashboard interactions
5. **Tailwind CSS** — Rapid UI development without custom CSS
6. **Elasticsearch** — Full-text search across datasets with autocomplete

## Future Improvements

- [ ] WebSocket streaming for real-time data updates
- [ ] Cassandra for time-series event storage
- [ ] Spark ML for anomaly detection
- [ ] AWS deployment (ECS Fargate + S3)
- [ ] CI/CD with GitHub Actions
- [ ] Data visualization dashboard with D3.js

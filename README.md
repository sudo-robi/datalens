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
| API Gateway | Java 21, Spring Boot 3.2, Gradle, JWT (jjwt), BCrypt |
| Ingestion | Go 1.22, net/http, lib/pq, Elasticsearch client |
| Search | Elasticsearch 8 |
| Storage | PostgreSQL 16 |
| Analytics | Apache Spark (PySpark) |
| Build | Gradle (Java), Vite (Frontend), go build (Go) |
| Testing | JUnit 5, Mockito, Vitest, React Testing Library |

## Prerequisites

- Java 21+ (the build uses a Java 21 toolchain; set `JAVA_HOME` if your default JDK is newer)
- Go 1.22+
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

## Running Tests & Coverage

### Java Tests (JUnit 5 + JaCoCo)
```bash
cd api-gateway
./gradlew test                    # Run tests
./gradlew test jacocoTestReport   # Run + generate coverage report
```
Coverage report: `api-gateway/build/reports/jacoco/test/html/index.html`

### Frontend Tests (Vitest + React Testing Library)
```bash
cd frontend
npm test                          # Run tests in watch mode
npx vitest run                    # Single run
npx vitest run --coverage         # Run with coverage report
```

### Go Tests (go test)
```bash
cd ingestion-service
go test ./...                     # Run tests
make test-coverage                # Run + generate coverage report
```
Coverage report: `ingestion-service/coverage.html`

### Spark Analytics
```bash
cd spark-jobs
pip install -r requirements.txt
python analytics_job.py /path/to/dataset.csv 1
```

## Test Coverage

| Service | Language | Tests | Line Coverage | Notes |
|---------|----------|-------|---------------|-------|
| API Gateway | Java | 16 | **50.9%** | 100% on SecurityConfig, 61% on security, 37% on service |
| Ingestion Service | Go | 50 | **60.0%** | 87.8% on ES client, 100% on CSV parsing |
| Frontend | TypeScript | 59 | **59%** statements | 100% on auth slice, App, Layout, Dashboard |
| **Total** | | **125** | | |

## Project Structure

```
datalens/
├── frontend/              # React + TypeScript + Redux + Tailwind
├── api-gateway/           # Java Spring Boot (Gradle)
├── ingestion-service/     # Go microservice
├── spark-jobs/            # PySpark analytics
├── deploy/                # Cloud Run deployment configs
│   ├── cloud-run.sh       # Interactive deploy script
│   └── cloudbuild.yaml    # CI/CD pipeline
├── docker-compose.yml     # PostgreSQL + Elasticsearch
└── README.md
```

## Design Decisions

### 1. PostgreSQL over Cassandra

Chose PostgreSQL for the MVP because the data model is inherently relational — users own datasets, datasets have metadata, and queries are primarily user-scoped lookups. Cassandra excels at write-heavy time-series workloads with partition-key routing, but adds operational complexity (compaction tuning, materialized views, no JOINs) that isn't justified at this scale. PostgreSQL gives us ACID transactions, foreign keys, and a familiar SQL interface, making it trivial to add features like dataset sharing or audit logs later. The plan is to evaluate Cassandra specifically for time-series event storage (page 4 of the future improvements) once real-time streaming is implemented.

### 2. Go for the Ingestion Service

CSV ingestion is I/O-bound (reading files, writing to ES) and benefits from Go's lightweight goroutines for concurrent batch indexing. The Go standard library's `net/http` and `encoding/csv` are production-grade — no need for a framework. Go compiles to a single static binary (~15MB), making the Docker image tiny and cold starts fast on Cloud Run. The alternative would be adding this to the Spring Boot service, but that would couple ingestion latency to API request handling and make it harder to scale independently.

### 3. Java (Spring Boot) for the API Gateway

Spring Security provides battle-tested JWT authentication, BCrypt password hashing, and CORS configuration out of the box. The API gateway handles auth, request routing, and validation — exactly the kind of structured, boilerplate-heavy work where a framework earns its keep. Java 21's virtual threads (available but not yet used) would let us add WebSocket streaming without an event-loop rewrite. Choosing Java also aligns with Palantir's backend stack, making this project directly relevant to the target role.

### 4. Elasticsearch for Full-Text Search

Datasets contain heterogeneous column data (strings, numbers, dates) that users need to search across freely. Elasticsearch's `multi_match` query with wildcard fields handles this naturally. The alternative — PostgreSQL full-text search with `tsvector` — works but requires manual index tuning per column type and doesn't support fuzzy matching or autocomplete. ES also gives us a path to analytics dashboards (Kibana) without additional infrastructure.

### 5. Redux Toolkit for State Management

The dashboard has three interconnected state domains: auth (token + user), datasets (list + detail), and search (query + results). Redux Toolkit's `createSlice` and `createAsyncThunk` eliminate the boilerplate of manual action types and reducers while keeping state predictable and debuggable. The `configureStore` setup integrates DevTools for time-travel debugging during development. For a app this size, Context API would work, but Redux makes the data flow explicit and testable — which is valuable for a portfolio project where code readability matters.

### 6. Microservice Decomposition (Java + Go)

Splitting the API gateway (Java) from the ingestion service (Go) demonstrates understanding of polyglot microservice architecture. Each service has a clear bounded context: the gateway handles auth and CRUD, while ingestion handles file processing and ES indexing. This lets us deploy, scale, and iterate on each independently. The Go service can be restarted or scaled during bulk CSV imports without affecting API availability. In production, this boundary would also let us swap the ingestion pipeline (e.g., to Apache Kafka consumers) without touching the API layer.

### 7. Multi-Stage Docker Builds

Every service uses multi-stage builds: a build stage with the full SDK (JDK, Go, Node) and a runtime stage with only the compiled artifact. This keeps production images small (50-100MB instead of 500MB+) and reduces the attack surface. The Go service runs as a non-root `app` user, and the frontend Nginx container runs as the `nginx` user — both with minimal OS packages installed.

### 8. Tailwind CSS over Component Libraries

Tailwind gives us complete control over the visual design without fighting a library's opinionated defaults. For a portfolio project, this matters — the UI needs to look intentional, not like every other Bootstrap/Material app. The utility-first approach also means the CSS bundle is small (purged unused classes) and there's no component library version to maintain.

## Deployment (Google Cloud Run)

### Prerequisites
- Google Cloud account with billing enabled
- `gcloud` CLI installed and authenticated (`gcloud auth login`)
- Cloud SQL (PostgreSQL) and Elasticsearch instances provisioned

### Quick Deploy

```bash
# Set your GCP project
export PROJECT_ID=your-gcp-project-id

# Run the deployment script
./deploy/cloud-run.sh
```

The script will prompt for:
- Cloud SQL connection name (optional — for Cloud SQL Proxy)
- PostgreSQL connection string
- Elasticsearch URL
- JWT signing secret

### Architecture on Cloud Run

- **API Gateway** → Cloud Run (auto-scales to 0, free tier covers idle)
- **Ingestion Service** → Cloud Run
- **Frontend** → Cloud Run + Nginx
- **PostgreSQL** → Cloud SQL (or free-tier Supabase/Neon)
- **Elasticsearch** → Elastic Cloud free trial or self-hosted

### Cost Estimate (Free Tier)

| Service | Free Tier | Cost |
|---------|-----------|------|
| Cloud Run | 240,000 vCPU-seconds/month | $0 (idle + light usage) |
| Cloud SQL | db-f1-micro instance | ~$7.67/month |
| Alternative | Supabase / Neon (free PostgreSQL) | $0 |

### CI/CD with Cloud Build

```bash
# Trigger a build manually
gcloud builds submit --config deploy/cloudbuild.yaml .

# Or set up a trigger on the main branch
gcloud builds triggers create github \
  --repo-name=datalens \
  --repo-owner=your-github-user \
  --branch-pattern=^main$ \
  --build-config=deploy/cloudbuild.yaml
```

### Manual Service Management

```bash
# List running services
gcloud run services list --region=us-central1

# View logs
gcloud logs read --service=datalens-api-gateway --limit=50

# Update a single service
gcloud run deploy datalens-api-gateway \
  --image=gcr.io/$PROJECT_ID/datalens-api-gateway:latest \
  --region=us-central1

# Delete a service
gcloud run services delete datalens-frontend --region=us-central1
```

## Future Improvements

- [ ] WebSocket streaming for real-time data updates
- [ ] Cassandra for time-series event storage
- [ ] Spark ML for anomaly detection
- [x] ~~AWS deployment (ECS Fargate + S3)~~ → Deployed to Google Cloud Run
- [x] ~~CI/CD with GitHub Actions~~ → Cloud Build pipeline in `deploy/cloudbuild.yaml`
- [ ] Data visualization dashboard with D3.js

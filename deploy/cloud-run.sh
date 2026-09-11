#!/usr/bin/env bash
# =============================================================================
# DataLens — Google Cloud Run Deployment Script
#
# Deploys all three services (API Gateway, Ingestion Service, Frontend) to
# Google Cloud Run. Requires gcloud CLI authenticated with a project.
#
# Usage:
#   export PROJECT_ID=your-gcp-project-id
#   ./deploy/cloud-run.sh
# =============================================================================

set -euo pipefail

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
REGION="${REGION:-us-central1}"
API_GATEWAY_SERVICE="datalens-api-gateway"
INGESTION_SERVICE="datalens-ingestion-service"
FRONTEND_SERVICE="datalens-frontend"
GCR_PREFIX="gcr.io"

# Resource limits (free-tier friendly)
MIN_INSTANCES="0"          # Scale to zero when idle (saves cost)
MAX_INSTANCES="3"          # Cap for free tier
CPU="0.5"                  # Smallest allocation
MEMORY="512Mi"            # Smallest allocation
TIMEOUT="300s"             # 5 minute request timeout

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
log()   { echo -e "${BLUE}[INFO]${NC}  $*"; }
ok()    { echo -e "${GREEN}[OK]${NC}    $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
err()   { echo -e "${RED}[ERROR]${NC} $*" >&2; }
die()   { err "$@"; exit 1; }

# ---------------------------------------------------------------------------
# Pre-flight checks
# ---------------------------------------------------------------------------
preflight() {
    log "Running pre-flight checks..."

    # gcloud installed?
    if ! command -v gcloud &>/dev/null; then
        die "gcloud CLI not found. Install it: https://cloud.google.com/sdk/docs/install"
    fi

    # gcloud authenticated?
    if ! gcloud auth print-access-token &>/dev/null; then
        die "gcloud not authenticated. Run: gcloud auth login"
    fi

    # PROJECT_ID set?
    if [[ -z "${PROJECT_ID:-}" ]]; then
        echo -n "Enter your GCP Project ID: "
        read -r PROJECT_ID
        export PROJECT_ID
    fi

    # Validate project exists
    if ! gcloud projects describe "$PROJECT_ID" &>/dev/null; then
        die "Project '$PROJECT_ID' not found or not accessible."
    fi

    log "Using project: ${PROJECT_ID}"
    gcloud config set project "$PROJECT_ID"
    gcloud config set run/region "$REGION"

    # Enable required APIs
    log "Enabling required GCP APIs..."
    gcloud services enable \
        run.googleapis.com \
        artifactregistry.googleapis.com \
        cloudbuild.googleapis.com \
        sqladmin.googleapis.com \
        --project="$PROJECT_ID"

    ok "Pre-flight checks passed"
}

# ---------------------------------------------------------------------------
# Collect configuration from user
# ---------------------------------------------------------------------------
collect_config() {
    log "Collecting deployment configuration..."

    # Cloud SQL connection name (optional)
    if [[ -z "${CLOUD_SQL_CONNECTION:-}" ]]; then
        echo -n "Cloud SQL connection name (leave empty to skip, format: PROJECT:REGION:INSTANCE): "
        read -r CLOUD_SQL_CONNECTION
    fi

    # Database URL
    if [[ -z "${DB_URL:-}" ]]; then
        echo -n "PostgreSQL connection string (e.g., postgres://user:pass@host:5432/dbname): "
        read -r DB_URL
    fi

    # Elasticsearch URL
    if [[ -z "${ELASTICSEARCH_URL:-}" ]]; then
        echo -n "Elasticsearch URL (e.g., https://your-instance.es.io): "
        read -r ELASTICSEARCH_URL
    fi

    # JWT Secret
    if [[ -z "${JWT_SECRET:-}" ]]; then
        echo -n "JWT signing secret (will be base64-encoded for Cloud Run): "
        read -rs JWT_SECRET
        echo
    fi

    # Validate required vars
    [[ -z "${DB_URL:-}" ]] && die "DB_URL is required"
    [[ -z "${ELASTICSEARCH_URL:-}" ]] && die "ELASTICSEARCH_URL is required"
    [[ -z "${JWT_SECRET:-}" ]] && die "JWT_SECRET is required"

    ok "Configuration collected"
}

# ---------------------------------------------------------------------------
# Deploy: API Gateway (Java Spring Boot)
# ---------------------------------------------------------------------------
deploy_api_gateway() {
    log "Deploying API Gateway..."

    local image="${GCR_PREFIX}/${PROJECT_ID}/${API_GATEWAY_SERVICE}:latest"

    # Build and push
    log "Building API Gateway image..."
    gcloud builds submit \
        --tag "$image" \
        --project="$PROJECT_ID" \
        --quiet \
        api-gateway/

    # Build env vars
    local env_vars="DB_URL=${DB_URL},ELASTICSEARCH_URL=${ELASTICSEARCH_URL},JWT_SECRET=${JWT_SECRET}"

    # Build cloud-sql-connections if configured
    local add_cloud_sql_flags=()
    if [[ -n "${CLOUD_SQL_CONNECTION:-}" ]]; then
        add_cloud_sql_flags=(--add-cloudsql-instances "${CLOUD_SQL_CONNECTION}")
        env_vars="${env_vars},CLOUD_SQL_CONNECTION=${CLOUD_SQL_CONNECTION}"
    fi

    # Deploy to Cloud Run
    gcloud run deploy "$API_GATEWAY_SERVICE" \
        --image "$image" \
        --platform managed \
        --region "$REGION" \
        --allow-unauthenticated \
        --min-instances "$MIN_INSTANCES" \
        --max-instances "$MAX_INSTANCES" \
        --cpu "$CPU" \
        --memory "$MEMORY" \
        --timeout "$TIMEOUT" \
        --port 8080 \
        --set-env-vars "$env_vars" \
        "${add_cloud_sql_flags[@]}" \
        --project="$PROJECT_ID" \
        --quiet

    local url
    url=$(gcloud run services describe "$API_GATEWAY_SERVICE" \
        --region "$REGION" \
        --format 'value(status.url)' \
        --project="$PROJECT_ID")

    ok "API Gateway deployed: ${url}"
    echo "$url"
}

# ---------------------------------------------------------------------------
# Deploy: Ingestion Service (Go)
# ---------------------------------------------------------------------------
deploy_ingestion() {
    log "Deploying Ingestion Service..."

    local image="${GCR_PREFIX}/${PROJECT_ID}/${INGESTION_SERVICE}:latest"

    # Build and push
    log "Building Ingestion Service image..."
    gcloud builds submit \
        --tag "$image" \
        --project="$PROJECT_ID" \
        --quiet \
        ingestion-service/

    # Build env vars
    local env_vars="DB_URL=${DB_URL},ELASTICSEARCH_URL=${ELASTICSEARCH_URL}"

    # Build cloud-sql-connections if configured
    local add_cloud_sql_flags=()
    if [[ -n "${CLOUD_SQL_CONNECTION:-}" ]]; then
        add_cloud_sql_flags=(--add-cloudsql-instances "${CLOUD_SQL_CONNECTION}")
        env_vars="${env_vars},CLOUD_SQL_CONNECTION=${CLOUD_SQL_CONNECTION}"
    fi

    # Deploy to Cloud Run
    gcloud run deploy "$INGESTION_SERVICE" \
        --image "$image" \
        --platform managed \
        --region "$REGION" \
        --allow-unauthenticated \
        --min-instances "$MIN_INSTANCES" \
        --max-instances "$MAX_INSTANCES" \
        --cpu "$CPU" \
        --memory "$MEMORY" \
        --timeout "$TIMEOUT" \
        --port 8081 \
        --set-env-vars "$env_vars" \
        "${add_cloud_sql_flags[@]}" \
        --project="$PROJECT_ID" \
        --quiet

    local url
    url=$(gcloud run services describe "$INGESTION_SERVICE" \
        --region "$REGION" \
        --format 'value(status.url)' \
        --project="$PROJECT_ID")

    ok "Ingestion Service deployed: ${url}"
    echo "$url"
}

# ---------------------------------------------------------------------------
# Deploy: Frontend (React + Nginx)
# ---------------------------------------------------------------------------
deploy_frontend() {
    local api_gateway_url="$1"
    log "Deploying Frontend..."

    local image="${GCR_PREFIX}/${PROJECT_ID}/${FRONTEND_SERVICE}:latest"

    # Build and push
    log "Building Frontend image..."
    gcloud builds submit \
        --tag "$image" \
        --project="$PROJECT_ID" \
        --quiet \
        frontend/

    # Replace API_GATEWAY_URL placeholder in nginx.conf
    # We use a sed command to rewrite the proxy_pass before deploying
    log "Patching nginx.conf with API Gateway URL: ${api_gateway_url}"
    local tmp_dir
    tmp_dir=$(mktemp -d)
    cp frontend/nginx.conf "${tmp_dir}/nginx.conf"
    sed -i "s|http://API_GATEWAY_URL:8080|${api_gateway_url}|g" "${tmp_dir}/nginx.conf"

    # Deploy to Cloud Run (no special env vars needed — baked into nginx.conf)
    gcloud run deploy "$FRONTEND_SERVICE" \
        --image "$image" \
        --platform managed \
        --region "$REGION" \
        --allow-unauthenticated \
        --min-instances "$MIN_INSTANCES" \
        --max-instances "$MAX_INSTANCES" \
        --cpu "$CPU" \
        --memory "$MEMORY" \
        --timeout "$TIMEOUT" \
        --port 80 \
        --project="$PROJECT_ID" \
        --quiet

    local url
    url=$(gcloud run services describe "$FRONTEND_SERVICE" \
        --region "$REGION" \
        --format 'value(status.url)' \
        --project="$PROJECT_ID")

    rm -rf "$tmp_dir"

    ok "Frontend deployed: ${url}"
    echo "$url"
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
main() {
    echo ""
    echo "============================================="
    echo "  DataLens — Cloud Run Deployment"
    echo "============================================="
    echo ""

    preflight
    collect_config

    echo ""
    log "Starting deployment to Cloud Run (${REGION})..."
    echo ""

    # Deploy in order (frontend depends on API Gateway URL)
    local api_url ingestion_url frontend_url

    api_url=$(deploy_api_gateway)
    echo ""

    ingestion_url=$(deploy_ingestion)
    echo ""

    frontend_url=$(deploy_frontend "$api_url")
    echo ""

    # Print summary
    echo "============================================="
    echo "  Deployment Complete!"
    echo "============================================="
    echo ""
    echo "  API Gateway:    ${api_url}"
    echo "  Ingestion Svc:  ${ingestion_url}"
    echo "  Frontend:       ${frontend_url}"
    echo ""
    echo "  Region:         ${REGION}"
    echo "  Project:        ${PROJECT_ID}"
    echo ""
    echo "  Cloud SQL:      ${CLOUD_SQL_CONNECTION:-Not configured}"
    echo ""
    echo "  Note: Services may take 30-60s to respond"
    echo "        on first request (cold start)."
    echo "============================================="
}

main "$@"

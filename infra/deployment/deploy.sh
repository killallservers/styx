#!/bin/bash
# Allora Deployment Script
# Called by GitHub Actions CI/CD to deploy images to production server
# Usage: ./deploy.sh <image-tag>

set -e

IMAGE_TAG="${1:-latest}"
COMPOSE_DIR="/opt/allora/compose"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-ghcr.io/killallservers}"

echo "=== Allora Deployment Starting ==="
echo "Image tag: ${IMAGE_TAG}"
echo "Registry: ${DOCKER_REGISTRY}"
echo "Compose directory: ${COMPOSE_DIR}"

# Check if compose directory exists
if [ ! -d "${COMPOSE_DIR}" ]; then
  echo "ERROR: Compose directory not found at ${COMPOSE_DIR}"
  exit 1
fi

cd "${COMPOSE_DIR}"

# Load environment
if [ -f .env ]; then
  export $(cat .env | grep -v '^#' | xargs)
else
  echo "ERROR: .env file not found at ${COMPOSE_DIR}/.env"
  exit 1
fi

# Authenticate to GHCR if token provided
if [ -n "$GITHUB_TOKEN" ]; then
  echo "Authenticating to GitHub Container Registry..."
  echo "$GITHUB_TOKEN" | docker login ghcr.io -u oauth2accesstoken --password-stdin
fi

# Pull latest images
echo "Pulling Docker images..."
export IMAGE_TAG="${IMAGE_TAG}"
docker compose -f compose.prod.yml pull

# Check if models exist, if not try to download them
echo "Checking for model cache at /opt/models..."
if [ ! -d "/opt/models" ] || [ -z "$(ls -A /opt/models 2>/dev/null)" ]; then
  echo "⚠️  Model cache not found or empty"
  echo ""
  echo "Attempting to download models (this may take 10-15 minutes)..."

  # Check if download script exists
  if [ -f "/opt/allora/deployment/download-models.sh" ]; then
    echo "Running model download script..."
    bash /opt/allora/deployment/download-models.sh || {
      echo "⚠️  Model download failed, but continuing deployment"
      echo "    vLLM will attempt to download models on container startup"
      echo "    (This may cause health checks to timeout)"
    }
  else
    echo "⚠️  Model download script not found at /opt/allora/deployment/download-models.sh"
    echo "    Please run the 'Initialize Server' and 'Download Models' workflows first"
    echo "    Or manually run: bash /opt/allora/deployment/download-models.sh"
  fi
  echo ""
else
  echo "✓ Model cache found at /opt/models"
  du -sh /opt/models | awk '{print "  Size: " $1}'
  echo ""
fi

# Stop old containers gracefully
echo "Stopping old services..."
docker compose -f compose.prod.yml down --timeout 30 || true

# Wait for services to fully stop
sleep 5

# Start new containers
echo "Starting services..."
export IMAGE_TAG="${IMAGE_TAG}"
docker compose -f compose.prod.yml up -d

# Wait for services to be healthy
# Note: vLLM containers (llm, embeddings) need up to 600s to start (model initialization)
# So we need to wait longer for dependent services (API, Web) to become healthy
echo "Waiting for services to become healthy..."
echo "  (This may take 10-15 minutes on first deployment)"
max_attempts=180  # 180 × 10s = 30 minutes total (sufficient for vLLM model init)
attempt=1

while [ $attempt -le $max_attempts ]; do
  echo "Health check attempt $attempt/$max_attempts ($(( ($attempt * 10) / 60 )) min)..."

  # Check container states and detailed info
  llm_state=$(docker compose -f compose.prod.yml ps llm --format='{{.State}}' 2>/dev/null || echo "unknown")
  embeddings_state=$(docker compose -f compose.prod.yml ps embeddings --format='{{.State}}' 2>/dev/null || echo "unknown")

  # Check restart count for embeddings
  embeddings_restart=$(docker inspect allora-embeddings 2>/dev/null | grep -A 1 '"RestartCount"' | grep -o '[0-9]*' | head -1 || echo "0")

  echo "  LLM state: $llm_state | Embeddings state: $embeddings_state (restarts: $embeddings_restart)"

  # If embeddings is restarting repeatedly, show its logs immediately
  if [ "$embeddings_state" = "restarting" ] && [ "$embeddings_restart" -gt 2 ]; then
    echo ""
    echo "⚠️  Embeddings container is crashing (restarted $embeddings_restart times)"
    echo "Last logs from embeddings:"
    docker compose -f compose.prod.yml logs embeddings --tail=30 2>/dev/null | head -30
    echo ""
  fi

  # Check if vLLM processes are running
  if [ "$llm_state" = "running" ]; then
    llm_procs=$(docker compose -f compose.prod.yml exec -T llm ps aux 2>/dev/null | grep -c "python.*vllm" || echo "0")
    echo "  LLM processes running: $llm_procs"
  fi

  if [ "$embeddings_state" = "running" ]; then
    embeddings_procs=$(docker compose -f compose.prod.yml exec -T embeddings ps aux 2>/dev/null | grep -c "python.*vllm" || echo "0")
    echo "  Embeddings processes running: $embeddings_procs"
  fi

  # Check Caddy
  if docker compose -f compose.prod.yml exec -T caddy curl -f http://localhost:2019/status > /dev/null 2>&1; then
    echo "  ✓ Caddy healthy"
  else
    echo "  ✗ Caddy not ready"
  fi

  # Check API
  if docker compose -f compose.prod.yml exec -T api curl -f http://localhost:7500/health > /dev/null 2>&1; then
    echo "  ✓ API healthy"
  else
    echo "  ✗ API not ready"
  fi

  # Check Web
  if docker compose -f compose.prod.yml exec -T web curl -f http://localhost:7501/health > /dev/null 2>&1; then
    echo "  ✓ Web healthy"
  else
    echo "  ✗ Web not ready"
  fi

  # If all checks pass, we're done
  if docker compose -f compose.prod.yml exec -T caddy curl -f http://localhost:2019/status > /dev/null 2>&1 && \
     docker compose -f compose.prod.yml exec -T api curl -f http://localhost:7500/health > /dev/null 2>&1 && \
     docker compose -f compose.prod.yml exec -T web curl -f http://localhost:7501/health > /dev/null 2>&1; then
    echo "✓ All services healthy"
    break
  fi

  # Early exit if embeddings keeps crashing (clear sign of a real error)
  if [ "$embeddings_restart" -gt 5 ]; then
    echo ""
    echo "❌ Embeddings container crashed 6+ times - there's a real error"
    break
  fi

  attempt=$((attempt + 1))
  sleep 10
done

# If we timed out or hit max restarts OR embeddings kept crashing, show detailed logs
if [ $attempt -gt $max_attempts ] || [ "$embeddings_restart" -gt 5 ]; then
  echo ""
  echo "❌ Services failed to become healthy after 30 minutes"
  echo ""

  # Test vLLM endpoints directly
  echo "=== Direct vLLM Health Endpoint Tests ==="
  echo "Testing LLM health endpoint..."
  docker compose -f compose.prod.yml exec -T llm curl -v http://localhost:8000/health 2>&1 | head -20
  echo ""
  echo "Testing Embeddings health endpoint..."
  docker compose -f compose.prod.yml exec -T embeddings curl -v http://localhost:8000/health 2>&1 | head -20
  echo ""

  echo "=== vLLM (LLM) Container Logs ==="
  docker compose -f compose.prod.yml logs llm --tail=100
  echo ""
  echo "=== vLLM (Embeddings) Container Logs ==="
  docker compose -f compose.prod.yml logs embeddings --tail=100
  echo ""
  echo "=== Model Cache Check ==="
  if [ -d /opt/models ]; then
    echo "✓ /opt/models exists"
    du -sh /opt/models
    echo ""
    echo "Model directories:"
    ls -ld /opt/models/models--* 2>/dev/null || echo "  ✗ No model directories found!"
    echo ""

    # Check if Ministral model exists
    if [ -d "/opt/models/models--mistralai--Ministral-3-3B-Instruct-2512" ]; then
      echo "✓ Ministral-3-3B model directory exists"
    else
      echo "✗ Ministral-3-3B model directory NOT FOUND!"
    fi

    # Check if Qwen model specifically exists
    if [ -d "/opt/models/models--Qwen--Qwen3-VL-Embedding-2B" ]; then
      echo "✓ Qwen3 model directory exists"
    else
      echo "✗ Qwen3 model directory NOT FOUND!"
      ls -la /opt/models/ | grep -i qwen || echo "  No Qwen directories in /opt/models"
    fi
  else
    echo "✗ /opt/models NOT FOUND - models not downloaded!"
    echo "  Run 'Initialize Server' and 'Download Models' workflows first"
  fi
  echo ""
  echo "=== Check /opt/models Permissions ==="
  ls -ld /opt/models
  echo ""
  echo "=== Embeddings Container Logs (Last 50 lines) ==="
  docker compose -f compose.prod.yml logs embeddings --tail=50 || echo "Could not retrieve logs"
  echo ""
  echo "=== API Container Logs ==="
  docker compose -f compose.prod.yml logs api --tail=20
  echo ""
  echo "=== Web Container Logs ==="
  docker compose -f compose.prod.yml logs web --tail=20
  echo ""
  echo "=== Full Container Status ==="
  docker compose -f compose.prod.yml ps
  echo ""
  echo "❌ Deployment FAILED"
  exit 1
fi

# Only reload Caddy if all services are healthy (not after failures)
echo "✓ All services healthy"
echo "Reloading Caddy configuration..."
if docker compose -f compose.prod.yml exec -T caddy caddy reload 2>/dev/null; then
  echo "✓ Caddy configuration reloaded"
else
  echo "⚠️  Caddy reload failed (may be normal if container just started)"
fi

# Cleanup old images
echo "Cleaning up old images..."
docker image prune -f --filter "label!=keep=true"

# Log deployment
echo "Deployment completed at $(date)" >> /var/log/allora-deployments.log
echo "Image tag: ${IMAGE_TAG}" >> /var/log/allora-deployments.log

# Show running containers
echo ""
echo "=== Running Containers ==="
docker compose -f compose.prod.yml ps

echo ""
echo "=== Deployment Complete ==="
echo "Verify at:"
echo "  API: curl https://${DOMAIN}/api/health"
echo "  Web: https://${DOMAIN}/"
echo ""

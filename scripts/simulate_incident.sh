#!/usr/bin/env bash
# =============================================================================
# BetKZ Incident Simulation Script
# Simulates a misconfigured Order Service (Assignment 4)
# =============================================================================

set -euo pipefail

COMPOSE="docker compose -f deployments/docker-compose.yml"
SERVICE="order-service"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

log()  { echo -e "${GREEN}[$(date +%T)] $*${NC}"; }
warn() { echo -e "${YELLOW}[$(date +%T)] $*${NC}"; }
err()  { echo -e "${RED}[$(date +%T)] $*${NC}"; }

# ── STEP 1: Verify healthy state ──────────────────────────────────────────────
log "=== INCIDENT SIMULATION: BetKZ Order Service ==="
log "Step 1: Verifying healthy state..."
curl -sf http://localhost:8003/health | python3 -m json.tool || warn "Health check failed (expected if not running)"

# ── STEP 2: Inject fault ──────────────────────────────────────────────────────
warn ""
warn "Step 2: INJECTING FAULT — setting wrong BACKEND_URL for order-service..."
warn "This simulates a misconfigured database/backend connection string."

# Override INJECT_FAULT env and recreate only the order-service container
INJECT_FAULT=true $COMPOSE up -d --no-deps --force-recreate \
  -e INJECT_FAULT=true \
  $SERVICE 2>/dev/null || \
  docker compose -f deployments/docker-compose.yml \
    run --rm -d \
    -e INJECT_FAULT=true \
    -e BACKEND_URL=http://backend-broken:9999 \
    --name betkz-order-service-faulty \
    -p 8003:8003 \
    $SERVICE uvicorn main:app --host 0.0.0.0 --port 8003 || true

# Alternative: just restart with env override
docker stop betkz-order-service 2>/dev/null || true
docker run -d --rm \
  --name betkz-order-service \
  --network betkz_betkz-net \
  -p 8003:8003 \
  -e PORT=8003 \
  -e INJECT_FAULT=true \
  -e BACKEND_URL=http://backend-broken:9999 \
  betkz-main-order-service \
  uvicorn main:app --host 0.0.0.0 --port 8003 || warn "Could not inject - run manually"

sleep 3

# ── STEP 3: Detect incident ───────────────────────────────────────────────────
warn ""
warn "Step 3: DETECTING INCIDENT — checking health endpoint..."
HEALTH=$(curl -sf http://localhost:8003/health 2>/dev/null || echo '{"status":"unreachable"}')
echo "$HEALTH" | python3 -m json.tool 2>/dev/null || echo "$HEALTH"

warn "Attempting to place a bet (should fail with 503)..."
curl -s -w "\nHTTP STATUS: %{http_code}\n" \
  -X POST http://localhost:8003/api/orders/bets \
  -H "Content-Type: application/json" \
  -d '{"stake":10,"selections":[{"odds_id":"test"}]}' || true

err ""
err "INCIDENT DETECTED: Order Service is failing to reach backend!"
err "Prometheus will show: order_service_failures_total increasing"
err "Grafana alert should trigger within 15s"

# ── STEP 4: Resolve incident ──────────────────────────────────────────────────
echo ""
log "Step 4: RESOLVING — restoring correct configuration..."
docker stop betkz-order-service 2>/dev/null || true
$COMPOSE up -d --no-deps --force-recreate $SERVICE
sleep 5

# ── STEP 5: Verify resolution ─────────────────────────────────────────────────
log "Step 5: VERIFYING resolution..."
curl -sf http://localhost:8003/health | python3 -m json.tool
log "Order Service restored! Incident resolved."

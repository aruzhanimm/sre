#!/usr/bin/env bash
# =============================================================================
# BetKZ Load Test Script — Assignment 6: Capacity Planning
# Simulates concurrent user traffic to observe system behaviour under load.
# Usage: ./scripts/load_test.sh [--duration 60] [--concurrency 10]
# =============================================================================

set -euo pipefail

# ── Config ───────────────────────────────────────────────────────────────────
DURATION=${DURATION:-60}       # seconds to run
CONCURRENCY=${CONCURRENCY:-10} # parallel workers
BASE_URL=${BASE_URL:-http://localhost}
ORDER_URL=${ORDER_URL:-http://localhost:8003}
AUTH_URL=${AUTH_URL:-http://localhost:8001}
USER_URL=${USER_URL:-http://localhost:8002}
RESULTS_DIR="./load_test_results"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
log()  { echo -e "${GREEN}[$(date +%T)]${NC} $*"; }
warn() { echo -e "${YELLOW}[$(date +%T)]${NC} $*"; }
err()  { echo -e "${RED}[$(date +%T)]${NC} $*"; }

mkdir -p "$RESULTS_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULT_FILE="$RESULTS_DIR/load_test_${TIMESTAMP}.txt"

log "============================================================"
log "  BetKZ Load Test — Assignment 6 Capacity Planning"
log "  Duration: ${DURATION}s | Concurrency: ${CONCURRENCY} workers"
log "  Results: $RESULT_FILE"
log "============================================================"

# ── Pre-flight health check ──────────────────────────────────────────────────
log "Pre-flight: verifying all services are healthy..."
SERVICES=(
  "Frontend:${BASE_URL}"
  "Auth-Service:${AUTH_URL}/health"
  "User-Service:${USER_URL}/health"
  "Order-Service:${ORDER_URL}/health"
)
ALL_UP=true
for svc in "${SERVICES[@]}"; do
  NAME="${svc%%:*}"
  URL="${svc#*:}"
  CODE=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 "$URL" || echo "000")
  if [[ "$CODE" == "200" ]]; then
    log "  ✓ $NAME — HTTP $CODE"
  else
    warn "  ✗ $NAME — HTTP $CODE (may affect results)"
    ALL_UP=false
  fi
done
if [[ "$ALL_UP" == false ]]; then
  warn "Some services are not responding. Results may be incomplete."
fi

# ── Counters ─────────────────────────────────────────────────────────────────
TOTAL_REQUESTS=0
SUCCESS=0
FAILURES=0

# ── Worker function ──────────────────────────────────────────────────────────
run_worker() {
  local worker_id=$1
  local end_time=$(($(date +%s) + DURATION))
  local local_total=0
  local local_ok=0
  local local_fail=0

  while [[ $(date +%s) -lt $end_time ]]; do
    # Rotate through endpoints to simulate realistic traffic
    ENDPOINT_IDX=$((local_total % 4))
    case $ENDPOINT_IDX in
      0)
        CODE=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 \
          "${AUTH_URL}/health" || echo "000")
        LABEL="auth-health"
        ;;
      1)
        CODE=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 \
          "${ORDER_URL}/health" || echo "000")
        LABEL="order-health"
        ;;
      2)
        CODE=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 \
          -X POST "${ORDER_URL}/api/orders/bets" \
          -H "Content-Type: application/json" \
          -d '{"stake":10,"selections":[{"odds_id":"load-test-'${worker_id}'"}]}' \
          || echo "000")
        LABEL="order-place-bet"
        ;;
      3)
        CODE=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 \
          "${USER_URL}/health" || echo "000")
        LABEL="user-health"
        ;;
    esac

    local_total=$((local_total + 1))
    if [[ "$CODE" == "200" || "$CODE" == "401" || "$CODE" == "422" ]]; then
      local_ok=$((local_ok + 1))
    else
      local_fail=$((local_fail + 1))
    fi

    echo "worker=${worker_id} req=${local_total} endpoint=${LABEL} status=${CODE}" >> "$RESULT_FILE"
  done

  # Write summary line for aggregation
  echo "WORKER_SUMMARY worker=${worker_id} total=${local_total} ok=${local_ok} fail=${local_fail}" >> "$RESULT_FILE"
}

# ── Launch workers ────────────────────────────────────────────────────────────
log "Starting $CONCURRENCY workers for ${DURATION}s..."
echo "# BetKZ Load Test — $(date)" > "$RESULT_FILE"
echo "# Duration=${DURATION}s  Concurrency=${CONCURRENCY}" >> "$RESULT_FILE"
echo "" >> "$RESULT_FILE"

PIDS=()
for i in $(seq 1 "$CONCURRENCY"); do
  run_worker "$i" &
  PIDS+=($!)
done

# Progress bar
log "Running... (Ctrl+C to abort)"
END_TIME=$(($(date +%s) + DURATION))
while [[ $(date +%s) -lt $END_TIME ]]; do
  REMAINING=$((END_TIME - $(date +%s)))
  DONE=$((DURATION - REMAINING))
  PCT=$((DONE * 100 / DURATION))
  BARS=$((PCT / 5))
  BAR=$(printf '#%.0s' $(seq 1 $BARS 2>/dev/null) 2>/dev/null || echo "##########")
  printf "\r  [%-20s] %3d%% (%ds remaining)" "$BAR" "$PCT" "$REMAINING"
  sleep 2
done
echo ""

# Wait for all workers
for PID in "${PIDS[@]}"; do
  wait "$PID" 2>/dev/null || true
done

# ── Aggregate results ────────────────────────────────────────────────────────
log "============================================================"
log "  LOAD TEST RESULTS"
log "============================================================"

TOTAL_REQ=$(grep -c "^worker=" "$RESULT_FILE" 2>/dev/null || echo 0)
TOTAL_OK=$(grep "^worker=" "$RESULT_FILE" 2>/dev/null | grep -c "status=200\|status=401\|status=422" || echo 0)
TOTAL_FAIL=$(grep "^worker=" "$RESULT_FILE" 2>/dev/null | grep -c "status=0\|status=503\|status=502\|status=500" || echo 0)

if [[ "$TOTAL_REQ" -gt 0 ]]; then
  SUCCESS_PCT=$((TOTAL_OK * 100 / TOTAL_REQ))
  FAIL_PCT=$(( TOTAL_REQ > 0 ? TOTAL_FAIL * 100 / TOTAL_REQ : 0 ))
  RPS=$((TOTAL_REQ / DURATION))
else
  SUCCESS_PCT=0; FAIL_PCT=0; RPS=0
fi

log "  Total Requests : $TOTAL_REQ"
log "  Success        : $TOTAL_OK ($SUCCESS_PCT%)"
log "  Failures       : $TOTAL_FAIL ($FAIL_PCT%)"
log "  Avg RPS        : $RPS req/s"
log "  Workers        : $CONCURRENCY"
log "  Duration       : ${DURATION}s"
log ""
log "  Endpoint breakdown:"
for ep in auth-health order-health order-place-bet user-health; do
  COUNT=$(grep "endpoint=${ep}" "$RESULT_FILE" 2>/dev/null | wc -l || echo 0)
  log "    $ep: $COUNT requests"
done

log ""
log "  Full results saved to: $RESULT_FILE"
log "  Check Grafana at http://localhost:3000 to see load in dashboards"
log "  Check Prometheus at http://localhost:9090 for raw metrics"

# ── Capacity assessment ──────────────────────────────────────────────────────
log ""
log "  === CAPACITY ASSESSMENT ==="
if [[ "$RPS" -ge 30 ]]; then
  warn "  HIGH LOAD: System is processing $RPS RPS."
  warn "  Recommendation: Consider horizontal scaling of order-service."
  warn "  → docker compose up --scale order-service=3"
elif [[ "$RPS" -ge 10 ]]; then
  log "  MODERATE LOAD: $RPS RPS — system is stable."
  log "  Monitor order_service_failures_total in Prometheus."
else
  log "  LOW LOAD: $RPS RPS — well within capacity."
fi

if [[ "$FAIL_PCT" -gt 10 ]]; then
  err "  ERROR RATE IS HIGH: $FAIL_PCT% failures!"
  err "  Check: docker compose -f deployments/docker-compose.yml logs order-service"
fi

log "============================================================"

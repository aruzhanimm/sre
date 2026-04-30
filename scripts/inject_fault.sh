#!/usr/bin/env bash
# Manually inject fault into Order Service for incident simulation
# Run from project root: bash scripts/inject_fault.sh

echo "=== INJECTING FAULT: Order Service misconfiguration ==="
docker stop betkz-order-service

docker run -d \
  --name betkz-order-service \
  --network betkz_betkz-net \
  -p 8003:8003 \
  -e PORT=8003 \
  -e INJECT_FAULT=true \
  betkz-main-order-service \
  uvicorn main:app --host 0.0.0.0 --port 8003

echo "Fault injected. Check:"
echo "  curl http://localhost:8003/health"
echo "  curl -X POST http://localhost:8003/api/orders/bets -H 'Content-Type: application/json' -d '{\"stake\":10}'"
echo ""
echo "To RESTORE: bash scripts/restore_service.sh"

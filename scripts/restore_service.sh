#!/usr/bin/env bash
# Restore Order Service after incident simulation
echo "=== RESTORING Order Service ==="
docker stop betkz-order-service 2>/dev/null || true
docker rm betkz-order-service 2>/dev/null || true
docker compose -f deployments/docker-compose.yml up -d --no-deps --force-recreate order-service
sleep 3
curl http://localhost:8003/health
echo ""
echo "Order Service restored!"

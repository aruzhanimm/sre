# BetKZ Deployment Guide

## Local container deployment

1. Open terminal in repository root:
   ```bash
   cd /Users/aruzanimka/Downloads/BetKZ-main
   ```
2. Start services with Docker Compose:
   ```bash
   docker-compose -f deployments/docker-compose.yml up -d --build
   ```
3. Verify containers:
   ```bash
   docker ps --filter "name=betkz"
   ```
4. Open services in browser:
   - Frontend: http://localhost
   - Backend health: http://localhost:8081/health
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000 (admin/admin)

## Verification

After deployment, verify all services are running:

```bash
# Check all containers are up
docker ps --filter "name=betkz"

# Verify backend health
curl -s http://localhost:8081/health | jq .status
# Expected: "ok"

# Verify database health
curl -s http://localhost:8081/health/db | jq .status
# Expected: "ok"

# Check metrics are exposed
curl -s http://localhost:8081/metrics | grep 'betkz_' | wc -l
# Expected: > 0

# Verify monitoring services
curl -s http://localhost:9090/-/ready
# Expected: "Prometheus Server is Ready."

curl -s http://localhost:3000/api/health | jq .database
# Expected: "ok"
```

## Monitoring

- Backend metrics endpoint: http://localhost:8081/metrics
- Prometheus scrapes the backend service from inside Docker.
- Grafana is provisioned with a BetKZ dashboard using Prometheus as default data source.

## Notes

- The frontend container uses Nginx to proxy `/api` and `/ws` requests to the backend service.
- PostgreSQL is exposed on host port `5433` to avoid local port conflicts.
- Redis uses default port `6379`.

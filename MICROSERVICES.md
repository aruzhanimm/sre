# BetKZ Microservices Architecture

## Overview

The project is split into **5 independent microservices** + the core Go backend:

| Service | Port | Language | Responsibility |
|---|---|---|---|
| **backend** (Go) | 8080 (internal) / 8081 | Go/Gin | Core logic: bets, events, odds, auth DB layer |
| **auth-service** | 8001 | Python/FastAPI | Authentication & JWT proxy |
| **user-service** | 8002 | Python/FastAPI | User profiles & balance management |
| **order-service** | 8003 | Python/FastAPI | Bet placement & order management |
| **chat-service** | 8004 | Python/FastAPI | Real-time WebSocket chat between users |

---

## Quick Start

```bash
cd deployments
docker compose up --build -d
```

Wait ~60s for all containers to start, then verify:

```bash
# Check all services healthy
curl http://localhost:8001/health   # auth-service
curl http://localhost:8002/health   # user-service
curl http://localhost:8003/health   # order-service
curl http://localhost:8004/health   # chat-service
curl http://localhost:8081/health   # backend (Go)
```

---

## Service Endpoints

### Auth Service (`:8001`)
```
POST /api/auth/register
POST /api/auth/login
POST /api/auth/refresh
GET  /api/auth/profile     (requires JWT)
POST /api/auth/logout      (requires JWT)
GET  /metrics
```

### User Service (`:8002`)
```
GET  /api/users/profile           (requires JWT)
GET  /api/users/transactions      (requires JWT)
POST /api/users/deposit           (admin)
POST /api/users/{id}/deposit      (admin)
GET  /metrics
```

### Order Service (`:8003`)
```
POST /api/orders/bets             (requires JWT)
GET  /api/orders/bets             (requires JWT)
GET  /api/orders/bets/{id}        (requires JWT)
GET  /api/orders/admin/bets       (admin)
POST /api/orders/admin/settle     (admin)
GET  /api/orders/admin/stats      (admin)
GET  /metrics
```

### Chat Service (`:8004`)
```
GET  /api/chat/rooms
GET  /api/chat/rooms/{room_id}/messages
WS   /ws/chat/{room_id}?username=YourName
GET  /metrics
```

---

## Monitoring

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

Prometheus scrapes `/metrics` from all 5 services.

---

## Incident Simulation (Assignment 4)

The **Order Service** supports fault injection via environment variable.

### Step 1 — Inject fault
```bash
# Stop the healthy container
docker stop betkz-order-service

# Start with broken backend URL (simulates DB misconfiguration)
docker run -d \
  --name betkz-order-service \
  --network betkz_betkz-net \
  -p 8003:8003 \
  -e PORT=8003 \
  -e INJECT_FAULT=true \
  betkz-main-order-service \
  uvicorn main:app --host 0.0.0.0 --port 8003
```

### Step 2 — Detect via health check & Prometheus
```bash
curl http://localhost:8003/health
# Returns: {"status": "FAULT_INJECTED", "backend_url": "http://backend-broken:9999", ...}

# Metrics will show order_service_failures_total increasing
curl http://localhost:8003/metrics | grep order_service_failures
```

### Step 3 — Restore service
```bash
bash scripts/restore_service.sh
```

---

## Project Structure

```
BetKZ-main/
├── backend/                    # Go monolith (untouched)
├── frontend/                   # React/Vite frontend
├── services/
│   ├── auth-service/           # Microservice 1
│   ├── user-service/           # Microservice 2
│   ├── order-service/          # Microservice 3
│   └── chat-service/           # Microservice 4
├── deployments/
│   ├── docker-compose.yml      # All services
│   ├── prometheus/
│   │   └── prometheus.yml      # Scrapes all 5 services
│   └── grafana/
├── terraform/                  # IaC (Assignment 5)
├── scripts/
│   ├── inject_fault.sh         # Incident simulation
│   └── restore_service.sh      # Incident resolution
└── docs/
```

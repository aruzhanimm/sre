"""
Order Service (Bet Service) - Microservice responsible for bet/order placement and settlement.
Proxies order operations to the main backend. This service is used in the Incident Simulation.
"""

import os
import time
import httpx
from fastapi import FastAPI, Request, Response, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from prometheus_client import Counter, Histogram, generate_latest, CONTENT_TYPE_LATEST, CollectorRegistry

BACKEND_URL = os.getenv("BACKEND_URL", "http://backend:8080")
PORT        = int(os.getenv("PORT", 8003))

# ── Fault injection (for incident simulation) ────────────────────────────────
# Set INJECT_FAULT=true to simulate a misconfigured backend URL (incident demo)
INJECT_FAULT = os.getenv("INJECT_FAULT", "false").lower() == "true"
if INJECT_FAULT:
    BACKEND_URL = os.getenv("FAULTY_BACKEND_URL", "http://backend-broken:9999")

app = FastAPI(title="BetKZ Order Service", version="1.0.0")
app.add_middleware(CORSMiddleware, allow_origins=["*"], allow_methods=["*"], allow_headers=["*"])

registry = CollectorRegistry()
REQUEST_COUNT   = Counter("order_service_requests_total",   "Total requests",   ["method", "endpoint", "status"], registry=registry)
REQUEST_LATENCY = Histogram("order_service_request_duration_seconds", "Latency", ["endpoint"], registry=registry)
FAILED_ORDERS   = Counter("order_service_failures_total", "Total failed orders", registry=registry)

@app.get("/health")
def health():
    fault_status = "FAULT_INJECTED" if INJECT_FAULT else "ok"
    return {"service": "order-service", "status": fault_status, "backend_url": BACKEND_URL, "timestamp": time.time()}

@app.get("/metrics")
def metrics():
    return Response(generate_latest(registry), media_type=CONTENT_TYPE_LATEST)

# Place bet
@app.post("/api/orders/bets")
async def place_bet(request: Request):
    return await _proxy(request, "/api/bets", "POST", "/bets")

# List bets
@app.get("/api/orders/bets")
async def list_bets(request: Request):
    return await _proxy(request, "/api/bets", "GET", "/bets")

# Get single bet
@app.get("/api/orders/bets/{bet_id}")
async def get_bet(bet_id: str, request: Request):
    return await _proxy(request, f"/api/bets/{bet_id}", "GET", "/bets/:id")

# Admin: list all bets
@app.get("/api/orders/admin/bets")
async def admin_list_bets(request: Request):
    return await _proxy(request, "/api/admin/bets", "GET", "/admin/bets")

# Admin: settle market
@app.post("/api/orders/admin/settle")
async def settle(request: Request):
    return await _proxy(request, "/api/admin/settle", "POST", "/admin/settle")

# Admin: stats
@app.get("/api/orders/admin/stats")
async def stats(request: Request):
    return await _proxy(request, "/api/admin/stats", "GET", "/admin/stats")

async def _proxy(request: Request, backend_path: str, method: str, label: str):
    start = time.time()
    headers = dict(request.headers)
    headers.pop("host", None)
    params = str(request.url.query)
    url = f"{BACKEND_URL}{backend_path}"
    if params:
        url += f"?{params}"

    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            body = await request.body()
            resp = await client.request(method, url, content=body, headers=headers)
        REQUEST_COUNT.labels(method=method, endpoint=label, status=str(resp.status_code)).inc()
        REQUEST_LATENCY.labels(endpoint=label).observe(time.time() - start)
        return Response(
            content=resp.content,
            status_code=resp.status_code,
            media_type=resp.headers.get("content-type", "application/json"),
        )
    except httpx.RequestError as e:
        FAILED_ORDERS.inc()
        REQUEST_COUNT.labels(method=method, endpoint=label, status="503").inc()
        REQUEST_LATENCY.labels(endpoint=label).observe(time.time() - start)
        raise HTTPException(status_code=503, detail=f"Order Service: backend unreachable ({BACKEND_URL}): {e}")

"""
User Service - Microservice responsible for user profile and balance management.
Proxies user operations to the main backend and exposes its own /metrics endpoint.
"""

import os
import time
import httpx
from fastapi import FastAPI, Request, Response, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from prometheus_client import Counter, Histogram, Gauge, generate_latest, CONTENT_TYPE_LATEST, CollectorRegistry

BACKEND_URL = os.getenv("BACKEND_URL", "http://backend:8080")
PORT        = int(os.getenv("PORT", 8002))

app = FastAPI(title="BetKZ User Service", version="1.0.0")
app.add_middleware(CORSMiddleware, allow_origins=["*"], allow_methods=["*"], allow_headers=["*"])

registry = CollectorRegistry()
REQUEST_COUNT   = Counter("user_service_requests_total", "Total requests", ["method", "endpoint", "status"], registry=registry)
REQUEST_LATENCY = Histogram("user_service_request_duration_seconds", "Request latency", ["endpoint"], registry=registry)
ACTIVE_USERS    = Gauge("user_service_active_users", "Simulated active users gauge", registry=registry)
ACTIVE_USERS.set(0)

@app.get("/health")
def health():
    return {"service": "user-service", "status": "ok", "timestamp": time.time()}

@app.get("/metrics")
def metrics():
    return Response(generate_latest(registry), media_type=CONTENT_TYPE_LATEST)

# Profile
@app.get("/api/users/profile")
async def get_profile(request: Request):
    return await _proxy(request, "/api/auth/profile", "GET", "/profile")

# Transactions
@app.get("/api/users/transactions")
async def list_transactions(request: Request):
    return await _proxy(request, "/api/transactions", "GET", "/transactions")

# Admin: deposit by email
@app.post("/api/users/deposit")
async def deposit(request: Request):
    return await _proxy(request, "/api/admin/deposit-by-email", "POST", "/deposit")

# Admin: deposit by id
@app.post("/api/users/{user_id}/deposit")
async def deposit_by_id(user_id: str, request: Request):
    return await _proxy(request, f"/api/admin/users/{user_id}/deposit", "POST", "/deposit-by-id")

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
        REQUEST_COUNT.labels(method=method, endpoint=label, status="503").inc()
        raise HTTPException(status_code=503, detail=f"Backend unreachable: {e}")

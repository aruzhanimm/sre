"""
Auth Service - Microservice responsible for authentication and authorization.
Proxies auth operations to the main backend and exposes its own /metrics endpoint.
"""

import os
import time
import httpx
from fastapi import FastAPI, Request, Response, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from prometheus_client import Counter, Histogram, generate_latest, CONTENT_TYPE_LATEST, CollectorRegistry

# ── Config ──────────────────────────────────────────────────────────────────
BACKEND_URL = os.getenv("BACKEND_URL", "http://backend:8080")
PORT        = int(os.getenv("PORT", 8001))

app = FastAPI(title="BetKZ Auth Service", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

# ── Prometheus metrics ───────────────────────────────────────────────────────
registry = CollectorRegistry()
REQUEST_COUNT   = Counter("auth_service_requests_total",   "Total requests",        ["method", "endpoint", "status"], registry=registry)
REQUEST_LATENCY = Histogram("auth_service_request_duration_seconds", "Request latency", ["endpoint"], registry=registry)

# ── Health ───────────────────────────────────────────────────────────────────
@app.get("/health")
def health():
    return {"service": "auth-service", "status": "ok", "timestamp": time.time()}

# ── Metrics ──────────────────────────────────────────────────────────────────
@app.get("/metrics")
def metrics():
    return Response(generate_latest(registry), media_type=CONTENT_TYPE_LATEST)

# ── Auth endpoints (proxy to main backend) ───────────────────────────────────
@app.post("/api/auth/register")
async def register(request: Request):
    return await _proxy(request, "/api/auth/register", "POST", "/register")

@app.post("/api/auth/login")
async def login(request: Request):
    return await _proxy(request, "/api/auth/login", "POST", "/login")

@app.post("/api/auth/refresh")
async def refresh(request: Request):
    return await _proxy(request, "/api/auth/refresh", "POST", "/refresh")

@app.get("/api/auth/profile")
async def profile(request: Request):
    return await _proxy(request, "/api/auth/profile", "GET", "/profile")

@app.post("/api/auth/logout")
async def logout(request: Request):
    return await _proxy(request, "/api/auth/logout", "POST", "/logout")

# ── Proxy helper ─────────────────────────────────────────────────────────────
async def _proxy(request: Request, backend_path: str, method: str, label: str):
    start = time.time()
    headers = dict(request.headers)
    headers.pop("host", None)

    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            body = await request.body()
            resp = await client.request(
                method,
                f"{BACKEND_URL}{backend_path}",
                content=body,
                headers=headers,
            )
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

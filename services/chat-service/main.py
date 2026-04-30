"""
Chat Service - Standalone microservice for real-time user-to-user chat.
Uses WebSockets for real-time communication with in-memory message storage.
"""

import os
import time
import json
import asyncio
from typing import Dict, Set
from fastapi import FastAPI, WebSocket, WebSocketDisconnect, Request, Response
from fastapi.middleware.cors import CORSMiddleware
from prometheus_client import Counter, Gauge, generate_latest, CONTENT_TYPE_LATEST, CollectorRegistry

PORT = int(os.getenv("PORT", 8004))

app = FastAPI(title="BetKZ Chat Service", version="1.0.0")
app.add_middleware(CORSMiddleware, allow_origins=["*"], allow_methods=["*"], allow_headers=["*"])

registry = CollectorRegistry()
MESSAGES_SENT     = Counter("chat_service_messages_total",   "Total chat messages sent",       registry=registry)
ACTIVE_CONNECTIONS = Gauge("chat_service_active_connections", "Active WebSocket connections",   registry=registry)
ROOMS_ACTIVE      = Gauge("chat_service_active_rooms",       "Active chat rooms",               registry=registry)

# ── In-memory state ───────────────────────────────────────────────────────────
rooms: Dict[str, Set[WebSocket]] = {}         # room_id -> set of websockets
history: Dict[str, list] = {}                 # room_id -> last 50 messages

# ── Health & metrics ──────────────────────────────────────────────────────────
@app.get("/health")
def health():
    return {
        "service": "chat-service",
        "status": "ok",
        "active_connections": sum(len(v) for v in rooms.values()),
        "active_rooms": len(rooms),
        "timestamp": time.time(),
    }

@app.get("/metrics")
def metrics():
    return Response(generate_latest(registry), media_type=CONTENT_TYPE_LATEST)

# ── REST: fetch room history ──────────────────────────────────────────────────
@app.get("/api/chat/rooms/{room_id}/messages")
def get_history(room_id: str):
    return {"room_id": room_id, "messages": history.get(room_id, [])}

# ── REST: list active rooms ───────────────────────────────────────────────────
@app.get("/api/chat/rooms")
def list_rooms():
    return {"rooms": [{"id": k, "users": len(v)} for k, v in rooms.items()]}

# ── WebSocket endpoint ────────────────────────────────────────────────────────
@app.websocket("/ws/chat/{room_id}")
async def chat_endpoint(websocket: WebSocket, room_id: str):
    await websocket.accept()
    username = websocket.query_params.get("username", "anonymous")

    # Join room
    if room_id not in rooms:
        rooms[room_id] = set()
        history[room_id] = []
    rooms[room_id].add(websocket)
    ACTIVE_CONNECTIONS.inc()
    ROOMS_ACTIVE.set(len(rooms))

    join_msg = _make_msg("system", f"{username} joined the room", room_id)
    await _broadcast(room_id, join_msg, exclude=None)

    # Send history to new connection
    for msg in history[room_id][-20:]:
        await websocket.send_json(msg)

    try:
        while True:
            data = await websocket.receive_text()
            try:
                payload = json.loads(data)
                text = payload.get("message", "")[:500]   # cap at 500 chars
            except json.JSONDecodeError:
                text = data[:500]

            msg = _make_msg(username, text, room_id)
            history[room_id].append(msg)
            history[room_id] = history[room_id][-50:]    # keep last 50
            await _broadcast(room_id, msg, exclude=None)
            MESSAGES_SENT.inc()

    except WebSocketDisconnect:
        rooms[room_id].discard(websocket)
        ACTIVE_CONNECTIONS.dec()
        if not rooms[room_id]:
            del rooms[room_id]
            del history[room_id]
        ROOMS_ACTIVE.set(len(rooms))

        leave_msg = _make_msg("system", f"{username} left the room", room_id)
        await _broadcast(room_id, leave_msg, exclude=None)

# ── Helpers ───────────────────────────────────────────────────────────────────
def _make_msg(user: str, text: str, room: str) -> dict:
    return {"user": user, "message": text, "room": room, "ts": time.time()}

async def _broadcast(room_id: str, msg: dict, exclude: WebSocket | None):
    if room_id not in rooms:
        return
    dead = set()
    for ws in list(rooms[room_id]):
        if ws is exclude:
            continue
        try:
            await ws.send_json(msg)
        except Exception:
            dead.add(ws)
    for ws in dead:
        rooms[room_id].discard(ws)

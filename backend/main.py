from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.middleware.cors import CORSMiddleware
from backend.routers import server, system, tools
from backend.paths import get_static_dir
import os

app = FastAPI(title="Llama Admin Pro")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(server.router, prefix="/api/server")
app.include_router(system.router, prefix="/api/system")
app.include_router(tools.router, prefix="/api/tools")

# Ensure static dir exists before mounting
static_dir = get_static_dir()
os.makedirs(static_dir, exist_ok=True)

# Important: This MUST be at the end, so it doesn't shadow the API routes
app.mount("/", StaticFiles(directory=static_dir, html=True), name="static")

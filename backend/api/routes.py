from fastapi import APIRouter, Request
from pydantic import BaseModel
from typing import Optional
import os
from backend.services.process_manager import manager

router = APIRouter()

class StartRequest(BaseModel):
    model: str
    host: Optional[str] = "127.0.0.1"
    port: Optional[int] = 8080
    ctx: Optional[str] = ""
    ngl: Optional[str] = ""
    threads: Optional[str] = ""
    auto: Optional[bool] = False

@router.get("/status")
def get_status():
    return {
        "running": manager.is_running(),
        "pid": manager.pid,
        "health": manager.health
    }

from backend.paths import get_app_dir

@router.post("/start")
def start_server(req: StartRequest):
    if not req.model:
        return {"ok": False, "error": "Selecciona un modelo"}
        
    exe_path = os.path.join(get_app_dir(), "bin", "llama-server.exe")
    args = ["-m", req.model, "--port", str(req.port), "--host", req.host]
    if req.ctx: args.extend(["-c", req.ctx])
    if req.ngl: args.extend(["-ngl", req.ngl])
    if req.threads: args.extend(["-t", req.threads])
    
    ok, msg = manager.start_process(exe_path, args, req.port)
    return {"ok": ok, "message": msg}

@router.post("/stop")
def stop_server():
    ok, msg = manager.stop_process()
    return {"ok": ok, "message": msg}

from fastapi import WebSocket, WebSocketDisconnect
import asyncio

@router.websocket("/ws/logs")
async def websocket_logs(websocket: WebSocket):
    await websocket.accept()
    last_idx = 0
    try:
        while True:
            lines = manager._logs[last_idx:]
            if lines:
                last_idx += len(lines)
                await websocket.send_json({"lines": lines})
            await asyncio.sleep(0.5)
    except WebSocketDisconnect:
        pass


from backend.services.model_manager import model_manager, load_config, save_config
from backend.services.hardware import hardware, get_auto_config

@router.get("/config")
def get_config():
    return load_config()

@router.get("/hardware")
def get_hardware():
    return hardware.get()

@router.get("/auto-config")
def auto_config_endpoint(model: str = "", lang: str = "es"):
    if not model or not os.path.isfile(model):
        return {"ok": False, "error": "Modelo inválido"}
    return {"ok": True, **get_auto_config(model, lang)}

@router.post("/config")
async def update_config(request: Request):
    body = await request.json()
    action = body.get("action")
    cfg = load_config()
    
    if action == "save_preset":
        name = (body.get("name") or "").strip()
        settings = body.get("settings")
        if not name or not isinstance(settings, dict):
            return {"ok": False, "error": "Nombre o configuración inválidos"}
        if "presets" not in cfg: cfg["presets"] = {}
        cfg["presets"][name] = settings
        save_config(cfg)
        return {"ok": True, "presets": cfg["presets"]}
        
    if action == "delete_preset":
        name = body.get("name")
        if "presets" in cfg:
            cfg["presets"].pop(name, None)
            save_config(cfg)
        return {"ok": True, "presets": cfg.get("presets", {})}
        
    if action == "save_last" or action == "save_settings":
        settings = body.get("settings")
        if isinstance(settings, dict):
            cfg["last_settings"] = settings
            save_config(cfg)
        return {"ok": True}
        
    if action == "set_scan_dirs":
        dirs = body.get("dirs")
        if isinstance(dirs, list):
            cfg["scan_dirs"] = [d.strip() for d in dirs if d.strip()]
            save_config(cfg)
            model_manager.init_watchers()
        return {"ok": True}
        
    return {"ok": False, "error": "acción desconocida"}

@router.get("/models")
def get_models(fresh: str = "0"):
    if fresh == "1":
        model_manager.init_watchers()
        
    return {
        "models": model_manager.get_models(),
        "scanning": False, 
        "last_scan": model_manager.last_scan,
        "dirs": model_manager.get_monitored_dirs()
    }

@router.get("/tools")
def get_tools():
    tools = []
    bin_dir = os.path.join(get_app_dir(), "bin")
    server_ok = False
    if os.path.isdir(bin_dir):
        server_ok = os.path.isfile(os.path.join(bin_dir, "llama-server.exe"))
        for f in sorted(os.listdir(bin_dir)):
            if f.lower().endswith(".exe"):
                path = os.path.join(bin_dir, f)
                try:
                    size = os.path.getsize(path)
                    tools.append({"name": f, "size_mb": round(size / (1024 ** 2), 1)})
                except OSError:
                    pass
    return {"tools": tools, "bin_dir": bin_dir, "server_ok": server_ok}

@router.get("/logs")
def get_logs(since: int = 0):
    with manager._lock:
        lines = manager._logs[since:]
        next_idx = len(manager._logs)
    return {"lines": lines, "next": next_idx, "running": manager.is_running()}

@router.get("/tool/status")
def get_tool_status():
    return {"running": False, "tool": None}

@router.get("/tool/logs")
def get_tool_logs(since: str = "0"):
    return {"lines": [], "running": False, "next": 0}

from fastapi import APIRouter
from backend.paths import get_data_dir
from backend.services.system_info import get_system_specs
import os

router = APIRouter()

@router.get("/auto-config")
async def get_auto_config():
    return get_system_specs()

@router.get("/logs")
async def get_logs(since: int = 0):
    log_path = os.path.join(get_data_dir(), "server.log")
    if not os.path.exists(log_path):
        return {"lines": "", "next": 0}
        
    try:
        with open(log_path, "r", encoding="utf-8") as f:
            f.seek(since)
            lines = f.read()
            next_pos = f.tell()
        return {"lines": lines, "next": next_pos}
    except Exception:
        return {"lines": "", "next": since}

@router.post("/scan_models")
async def scan_models(dirs: list[str]):
    models = []
    # Always scan the default models directory
    default_models = os.path.join(get_data_dir(), "models")
    if os.path.exists(default_models):
        for f in os.listdir(default_models):
            if f.endswith(".gguf"):
                models.append(os.path.join(default_models, f))
                
    for d in dirs:
        if os.path.exists(d) and os.path.isdir(d):
            for f in os.listdir(d):
                if f.endswith(".gguf"):
                    models.append(os.path.join(d, f))
                    
    return {"models": list(set(models))}

from backend.services.installer import install_local_async, install_system_async
from backend.services.binary_resolver import is_binary_available
from fastapi import HTTPException

@router.get("/check-binaries")
async def check_binaries(strategy: str = "auto"):
    # Comprobamos si llama-server está disponible con la estrategia actual
    available = is_binary_available("llama-server", strategy)
    return {"installed": available}

@router.get("/check-updates")
async def check_updates_endpoint(strategy: str = "auto"):
    from backend.services.installer import check_updates_async
    return await check_updates_async(strategy)

@router.post("/install/local")
async def install_local_endpoint():
    try:
        await install_local_async()
        return {"status": "success"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/install/global")
async def install_global_endpoint():
    try:
        output = await install_system_async()
        return {"status": "success", "output": output}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

from fastapi import APIRouter, HTTPException
from backend.services.tool_manager import tool_manager
from backend.paths import get_data_dir
from pydantic import BaseModel
import os

router = APIRouter()

class BenchmarkRequest(BaseModel):
    model: str
    threads: int
    ngl: int
    prompt_tokens: int
    gen_tokens: int
    binary_strategy: str = "auto"

@router.post("/benchmark/start")
async def start_benchmark(req: BenchmarkRequest):
    if not req.model:
        raise HTTPException(400, "Modelo requerido")
    
    success = await tool_manager.run_benchmark(
        req.model, req.threads, req.ngl, req.prompt_tokens, req.gen_tokens, req.binary_strategy
    )
    if success:
        return {"status": "started"}
    raise HTTPException(500, "Benchmark ya en ejecución o hubo un error")

@router.post("/benchmark/stop")
async def stop_benchmark():
    await tool_manager.stop_benchmark()
    return {"status": "stopped"}

@router.get("/benchmark/logs")
async def get_benchmark_logs(since: int = 0):
    log_path = os.path.join(get_data_dir(), "bench.log")
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

class QuantizeRequest(BaseModel):
    input_model: str
    output_model: str
    method: str
    binary_strategy: str = "auto"

@router.post("/quantize/start")
async def start_quantize(req: QuantizeRequest):
    if not req.input_model or not req.output_model or not req.method:
        raise HTTPException(400, "Faltan parámetros requeridos para la cuantización")
    
    success = await tool_manager.run_quantize(
        req.input_model, req.output_model, req.method, req.binary_strategy
    )
    if success:
        return {"status": "started"}
    raise HTTPException(500, "Cuantización ya en ejecución o hubo un error")

@router.post("/quantize/stop")
async def stop_quantize():
    await tool_manager.stop_quantize()
    return {"status": "stopped"}

@router.get("/quantize/logs")
async def get_quantize_logs(since: int = 0):
    log_path = os.path.join(get_data_dir(), "quantize.log")
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

class ConvertRequest(BaseModel):
    model_dir: str
    outtype: str
    output_path: str = ""

@router.post("/convert/start")
async def start_convert(req: ConvertRequest):
    if not req.model_dir or not req.outtype:
        raise HTTPException(400, "Faltan parámetros requeridos para la conversión")
    
    success = await tool_manager.run_convert(
        req.model_dir, req.outtype, req.output_path
    )
    if success:
        return {"status": "started"}
    raise HTTPException(500, "Conversión ya en ejecución o hubo un error")

@router.post("/convert/stop")
async def stop_convert():
    await tool_manager.stop_convert()
    return {"status": "stopped"}

@router.get("/convert/logs")
async def get_convert_logs(since: int = 0):
    log_path = os.path.join(get_data_dir(), "convert.log")
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

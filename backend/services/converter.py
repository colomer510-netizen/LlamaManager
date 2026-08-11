import os
import sys
import asyncio
import urllib.request
from backend.paths import get_data_dir

CONVERT_SCRIPT_URL = "https://raw.githubusercontent.com/ggerganov/llama.cpp/master/convert_hf_to_gguf.py"

async def ensure_converter_dependencies():
    """Descarga el script de conversión y asegura las dependencias de pip."""
    script_path = os.path.join(get_data_dir(), "convert_hf_to_gguf.py")
    
    if not os.path.exists(script_path):
        try:
            urllib.request.urlretrieve(CONVERT_SCRIPT_URL, script_path)
        except Exception as e:
            raise Exception(f"Error descargando el script de conversión: {str(e)}")
            
    # Instalar dependencias necesarias para la conversión
    cmd = [sys.executable, "-m", "pip", "install", "torch", "transformers", "gguf", "sentencepiece", "protobuf", "accelerate"]
    process = await asyncio.create_subprocess_exec(
        *cmd,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.STDOUT
    )
    await process.communicate() # Esperamos a que termine
    
    return script_path

async def start_conversion(model_dir: str, outtype: str, output_path: str = ""):
    script_path = await ensure_converter_dependencies()
    
    cmd = [sys.executable, script_path, model_dir, "--outtype", outtype]
    if output_path:
        cmd.extend(["--outfile", output_path])
        
    process = await asyncio.create_subprocess_exec(
        *cmd,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.STDOUT,
        cwd=get_data_dir()
    )
    
    return process

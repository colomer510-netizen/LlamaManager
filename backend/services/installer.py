import os
import urllib.request
import urllib.error
import json
import zipfile
import platform
import asyncio
from backend.paths import get_bin_dir

async def install_local_async():
    """Descarga e instala la última versión de llama.cpp en bin/."""
    await asyncio.to_thread(install_local)

def install_local():
    """Descarga e instala la última versión (Windows CPU) en la carpeta local bin/."""
    if platform.system() != "Windows":
        raise Exception("Instalador local actualmente optimizado para Windows.")
        
    api_url = "https://api.github.com/repos/ggerganov/llama.cpp/releases/latest"
    
    req = urllib.request.Request(api_url, headers={'User-Agent': 'Mozilla/5.0'})
    try:
        with urllib.request.urlopen(req) as response:
            data = json.loads(response.read().decode())
    except Exception as e:
        raise Exception(f"Fallo al contactar GitHub API: {e}")

    assets = data.get("assets", [])
    download_url = None
    
    for asset in assets:
        name = asset["name"].lower()
        if "win" in name and name.endswith(".zip") and "cublas" not in name and "vulkan" not in name:
            download_url = asset["browser_download_url"]
            break
            
    if not download_url:
        for asset in assets:
            if "win" in asset["name"].lower() and asset["name"].endswith(".zip"):
                download_url = asset["browser_download_url"]
                break

    if not download_url:
        raise Exception("No se encontró ningún binario para Windows en el último release.")

    zip_path = os.path.join(get_bin_dir(), "temp_llama.zip")
    
    try:
        urllib.request.urlretrieve(download_url, zip_path)
        with zipfile.ZipFile(zip_path, 'r') as zip_ref:
            # En releases recientes, puede que los binarios estén en una subcarpeta
            for member in zip_ref.namelist():
                filename = os.path.basename(member)
                if not filename:
                    continue
                source = zip_ref.open(member)
                target = open(os.path.join(get_bin_dir(), filename), "wb")
                with source, target:
                    import shutil
                    shutil.copyfileobj(source, target)
    finally:
        if os.path.exists(zip_path):
            os.remove(zip_path)

async def install_system_async():
    """Ejecuta el script oficial de instalación global en PowerShell."""
    if platform.system() != "Windows":
        raise Exception("Este comando es para PowerShell en Windows.")
        
    cmd = ["powershell", "-NoProfile", "-Command", "irm https://llama.app/install.ps1 | iex"]
    
    process = await asyncio.create_subprocess_exec(
        *cmd,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.PIPE
    )
    
    stdout, stderr = await process.communicate()
    
    if process.returncode != 0:
        err_msg = stderr.decode('utf-8', errors='ignore')
        raise Exception(f"Error al instalar globalmente: {err_msg}")
        
    return stdout.decode('utf-8', errors='ignore')

def get_latest_github_tag():
    api_url = "https://api.github.com/repos/ggerganov/llama.cpp/releases/latest"
    req = urllib.request.Request(api_url, headers={'User-Agent': 'Mozilla/5.0'})
    try:
        with urllib.request.urlopen(req) as response:
            data = json.loads(response.read().decode())
            return data.get("tag_name", "")
    except Exception:
        return ""

async def get_local_version(strategy: str):
    from backend.services.binary_resolver import get_binary
    import re
    
    bin_path = get_binary("llama-server", strategy)
    if not bin_path:
        return ""
        
    try:
        process = await asyncio.create_subprocess_exec(
            bin_path, "--version",
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        stdout, _ = await process.communicate()
        out_str = stdout.decode('utf-8', errors='ignore')
        
        match = re.search(r'(?:version:|build)\s*(\d+)', out_str, re.IGNORECASE)
        if match:
            return match.group(1)
        return ""
    except Exception:
        return ""

async def check_updates_async(strategy: str):
    import re
    latest_tag = await asyncio.to_thread(get_latest_github_tag)
    local_ver = await get_local_version(strategy)
    
    latest_ver = ""
    match = re.search(r'b?(\d+)', latest_tag, re.IGNORECASE)
    if match:
        latest_ver = match.group(1)
        
    has_update = False
    if local_ver and latest_ver:
        try:
            has_update = int(latest_ver) > int(local_ver)
        except:
            pass
            
    return {
        "installed": bool(local_ver),
        "local_version": local_ver,
        "latest_version": latest_ver,
        "has_update": has_update
    }

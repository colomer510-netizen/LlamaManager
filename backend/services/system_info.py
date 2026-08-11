import os
import subprocess
import shutil

def get_system_specs():
    specs = {
        "cpu_cores": os.cpu_count() or 4,
        "gpu_name": "Ninguna",
        "has_nvidia": False,
        "has_amd": False,
        "recommended_threads": 4,
        "recommended_ngl": 0
    }
    
    # 1. CPU
    # Recommended threads: physical cores, or total cores - 2 (to leave some for OS)
    # os.cpu_count() returns logical cores.
    logical_cores = specs["cpu_cores"]
    specs["recommended_threads"] = max(1, logical_cores - 2) if logical_cores > 2 else logical_cores
    
    # 2. GPU
    try:
        if os.name == 'nt':
            output = subprocess.check_output(
                ["wmic", "path", "win32_VideoController", "get", "name"], 
                text=True, creationflags=subprocess.CREATE_NO_WINDOW
            )
            gpus = [line.strip() for line in output.split('\n') if line.strip() and "Name" not in line]
            if gpus:
                # Prefer dedicated GPUs over integrated ones if multiple
                dedicated = [g for g in gpus if "NVIDIA" in g.upper() or "AMD" in g.upper() or "RADEON" in g.upper()]
                specs["gpu_name"] = dedicated[0] if dedicated else gpus[0]
                
                name_upper = specs["gpu_name"].upper()
                specs["has_nvidia"] = "NVIDIA" in name_upper
                specs["has_amd"] = "AMD" in name_upper or "RADEON" in name_upper
                
                # If they have a dedicated GPU, offload layers
                # Note: This is an estimation. A real implementation would check VRAM.
                # For simplicity we'll recommend offloading everything (99 layers usually caps it) 
                # if a dedicated GPU is found, or 0 if only integrated.
                if specs["has_nvidia"] or specs["has_amd"]:
                    specs["recommended_ngl"] = 99
    except Exception as e:
        print("Error detecting GPU:", e)
        
    return specs

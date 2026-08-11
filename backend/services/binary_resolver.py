import os
import platform
import shutil
from backend.paths import get_bin_dir

def get_executable_name(base_name: str) -> str:
    """Returns the executable name with .exe appended if on Windows."""
    return f"{base_name}.exe" if platform.system() == "Windows" else base_name

def get_binary_path(base_name: str, strategy: str = "auto") -> str:
    """Resolves the absolute path to a binary based on the selected strategy."""
    exe_name = get_executable_name(base_name)
    local_path = os.path.join(get_bin_dir(), exe_name)
    system_path = shutil.which(exe_name)
    
    if strategy == "local":
        return local_path
    elif strategy == "system":
        # Fallback to local_path so it throws the "not found in local" error if system is also missing,
        # or we could return system_path and let os.path.exists fail.
        return system_path if system_path else local_path
        
    # Auto strategy (Priority 1: Local, Priority 2: System)
    if os.path.isfile(local_path):
        return local_path
    if system_path:
        return system_path
        
    return local_path

def is_binary_available(base_name: str, strategy: str = "auto") -> bool:
    """Checks if a binary exists and is an executable."""
    path = get_binary_path(base_name, strategy)
    return os.path.isfile(path) if path else False

import os
import sys

def get_base_dir() -> str:
    """Returns the base directory of the application."""
    if getattr(sys, 'frozen', False):
        if hasattr(sys, '_MEIPASS'):
            return sys._MEIPASS
        return os.path.dirname(sys.executable)
    return os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

def get_data_dir() -> str:
    """Returns the directory for user data (config, models, bin)."""
    if getattr(sys, 'frozen', False):
        return os.path.dirname(sys.executable)
    return get_base_dir()

def get_bin_dir() -> str:
    return os.path.join(get_data_dir(), "bin")

def get_config_dir() -> str:
    return os.path.join(get_data_dir(), "config")

def get_static_dir() -> str:
    return os.path.join(get_base_dir(), "static")

# Ensure required directories exist
os.makedirs(get_bin_dir(), exist_ok=True)
os.makedirs(get_config_dir(), exist_ok=True)

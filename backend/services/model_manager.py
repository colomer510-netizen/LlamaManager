import os
import time
import json
import threading
from typing import Dict, Any, List
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

from backend.paths import get_base_dir

def load_config() -> Dict[str, Any]:
    cfg_path = os.path.join(get_base_dir(), "config.json")
    try:
        with open(cfg_path, "r", encoding="utf-8") as f:
            return json.load(f)
    except:
        return {"presets": {}, "scan_dirs": []}

def save_config(data: Dict[str, Any]):
    cfg_path = os.path.join(get_base_dir(), "config.json")
    with open(cfg_path, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=4)

class GGUFEventHandler(FileSystemEventHandler):
    def __init__(self, manager):
        self.manager = manager

    def on_created(self, event):
        if not event.is_directory and event.src_path.lower().endswith('.gguf'):
            self.manager.add_model(event.src_path)

    def on_deleted(self, event):
        if not event.is_directory and event.src_path.lower().endswith('.gguf'):
            self.manager.remove_model(event.src_path)

    def on_moved(self, event):
        if not event.is_directory:
            if event.src_path.lower().endswith('.gguf'):
                self.manager.remove_model(event.src_path)
            if event.dest_path.lower().endswith('.gguf'):
                self.manager.add_model(event.dest_path)

class ModelManager:
    def __init__(self):
        self.models_cache: Dict[str, Any] = {}
        self.lock = threading.Lock()
        self.last_scan = 0.0
        self.observer = None
        self.monitored_dirs = set()
        
        # Start initialization in background
        threading.Thread(target=self.init_watchers, daemon=True).start()

    def _get_model_info(self, path: str) -> Any:
        try:
            size = os.path.getsize(path)
            name = os.path.basename(path)
            return {
                "path": path,
                "name": name,
                "size_mb": round(size / (1024 ** 2), 1),
                "arch": "Desconocido",
                "quant": "Desconocido",
                "params": "?",
            }
        except OSError:
            return None

    def add_model(self, path: str):
        info = self._get_model_info(path)
        if info:
            with self.lock:
                self.models_cache[path] = info

    def remove_model(self, path: str):
        with self.lock:
            self.models_cache.pop(path, None)

    def init_watchers(self):
        cfg = load_config()
        dirs = cfg.get("scan_dirs", [])
        models_dir = os.path.join(get_base_dir(), "models")
        if models_dir not in dirs:
            dirs.append(models_dir)

        temp_cache = {}
        new_monitored_dirs = set()
        for d in dirs:
            if not os.path.isdir(d): continue
            new_monitored_dirs.add(d)
            try:
                for f in os.listdir(d):
                    if f.lower().endswith(".gguf"):
                        path = os.path.join(d, f)
                        info = self._get_model_info(path)
                        if info:
                            temp_cache[path] = info
            except OSError:
                pass

        with self.lock:
            self.models_cache = temp_cache
            self.last_scan = time.time()

        # Setup watchdog
        if self.observer:
            self.observer.stop()
            self.observer.join()

        self.observer = Observer()
        handler = GGUFEventHandler(self)
        for d in new_monitored_dirs:
            try:
                self.observer.schedule(handler, d, recursive=False)
            except Exception:
                pass
                
        self.monitored_dirs = new_monitored_dirs
        self.observer.start()

    def get_models(self) -> List[Any]:
        with self.lock:
            return sorted(self.models_cache.values(), key=lambda m: m["name"].lower())

    def get_monitored_dirs(self) -> List[str]:
        with self.lock:
            return list(self.monitored_dirs)

# Singleton instance
model_manager = ModelManager()

import psutil
import subprocess
import threading
import time
from typing import Optional, Dict

class ProcessManager:
    def __init__(self):
        self.proc: Optional[subprocess.Popen] = None
        self.pid: Optional[int] = None
        self.health: str = "off"
        self._logs: list = []
        self._lock = threading.Lock()

    def start_process(self, exe_path: str, args: list, port: int):
        if self.is_running():
            return False, "Ya está en ejecución."
        
        try:
            self.proc = subprocess.Popen(
                [exe_path] + args,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                stdin=subprocess.DEVNULL,
                text=True,
                encoding="utf-8",
                errors="replace"
            )
            self.pid = self.proc.pid
            self.health = "loading"
            
            # Start a thread to read stdout
            threading.Thread(target=self._read_stdout, daemon=True).start()
            return True, f"Proceso iniciado con PID {self.pid}"
        except Exception as e:
            return False, f"Error al iniciar: {e}"

    def stop_process(self):
        if not self.pid:
            return False, "No hay proceso en ejecución."
        
        try:
            parent = psutil.Process(self.pid)
            for child in parent.children(recursive=True):
                child.terminate()
            parent.terminate()
            
            gone, alive = psutil.wait_procs([parent], timeout=3)
            for p in alive:
                p.kill()
            
            self.proc = None
            self.pid = None
            self.health = "off"
            return True, "Proceso detenido limpiamente."
        except psutil.NoSuchProcess:
            self.proc = None
            self.pid = None
            self.health = "off"
            return True, "El proceso ya estaba muerto."
        except Exception as e:
            return False, f"Error al detener: {e}"
            
    def is_running(self) -> bool:
        if self.pid is None:
            return False
        try:
            p = psutil.Process(self.pid)
            return p.is_running() and p.status() != psutil.STATUS_ZOMBIE
        except psutil.NoSuchProcess:
            return False

    def _read_stdout(self):
        if not self.proc:
            return
        try:
            for line in self.proc.stdout:
                line = line.rstrip("\\n")
                if line:
                    with self._lock:
                        self._logs.append(line)
        except Exception:
            pass

manager = ProcessManager()

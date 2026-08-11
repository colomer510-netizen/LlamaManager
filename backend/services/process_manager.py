import subprocess
import threading
import time
from typing import Optional, List

class ProcessManager:
    def __init__(self):
        self.process: Optional[subprocess.Popen] = None
        self.log_file = None
        self.start_time = 0.0

    def start(self, cmd: List[str], log_path: str):
        if self.is_running():
            return
        
        self.log_file = open(log_path, "w", encoding="utf-8")
        self.start_time = time.time()
        
        # Cross-platform subprocess creation
        # On Windows we use CREATE_NO_WINDOW to hide the console if not explicitly needed
        creationflags = 0
        import platform
        if platform.system() == "Windows":
            creationflags = subprocess.CREATE_NO_WINDOW

        self.process = subprocess.Popen(
            cmd,
            stdout=self.log_file,
            stderr=subprocess.STDOUT,
            creationflags=creationflags
        )

    def stop(self):
        if self.process:
            self.process.terminate()
            try:
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.process.kill()
            self.process = None
        
        if self.log_file:
            self.log_file.close()
            self.log_file = None

    def is_running(self) -> bool:
        if self.process is None:
            return False
        if self.process.poll() is not None:
            # Process has finished
            return False
        return True

    def get_uptime(self) -> float:
        if self.is_running():
            return time.time() - self.start_time
        return 0.0

# Global instance
server_process = ProcessManager()

#!/usr/bin/env python3
"""Administrador web de llama.cpp.

Sirve una interfaz web local (http://127.0.0.1:8756) para iniciar/parar
llama-server, ver sus logs en vivo y chatear con el modelo.
Solo usa la librería estándar de Python. No requiere dependencias.

Uso: python app.py [puerto] [--browse]
"""
import json
import ctypes
import mimetypes
import os
import platform
import re
import shlex
import shutil
import struct
import subprocess
import sys
import threading
import time
import urllib.parse
import urllib.request
import webbrowser
import zipfile
from http.client import HTTPConnection
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
BIN_DIR = os.path.join(BASE_DIR, "bin")
STATIC_DIR = os.path.join(BASE_DIR, "static")
CONFIG_FILE = os.path.join(BASE_DIR, "config.json")
SERVER_EXE = os.path.join(BIN_DIR, "llama-server.exe")

MAX_LOG_LINES = 8000
CREATE_NO_WINDOW = 0x08000000 if os.name == "nt" else 0
CREATE_NEW_CONSOLE = 0x00000010 if os.name == "nt" else 0

DEFAULT_SCAN_DIRS = [
    os.path.expanduser("~"),
    os.path.join(os.path.expanduser("~"), "Downloads"),
    os.path.join(os.path.expanduser("~"), "Desktop"),
    os.path.join(os.path.expanduser("~"), "Documents"),
    os.path.join(os.path.expanduser("~"), ".lmstudio", "models"),
    os.path.join(os.path.expanduser("~"), ".cache"),
    os.path.join(os.path.expanduser("~"), ".local"),
    "C:\\models", "C:\\ai", "C:\\ia", "C:\\llm", "C:\\ollama",
    "D:\\", "E:\\", "F:\\",
]

SKIP_DIR_PARTS = ("$recycle.bin", "windows", "program files", "appdata",
                  "node_modules", ".git", ".cache", ".huggingface", ".npm",
                  "python", "site-packages", "lib\\site-packages", "venv",
                  ".venv", "system volume information", "programdata")


def default_config():
    return {
        "scan_dirs": [],
        "presets": {},
        "last_settings": {
            "host": "127.0.0.1",
            "port": "8080",
            "ctx": "4096",
            "ngl": "auto",
            "threads": "",
            "slots": "",
            "temp": "0.80",
            "top_p": "0.95",
            "top_k": "40",
            "repeat": "1.00",
            "seed": "-1",
            "flash": "auto",
            "api_key": "",
            "extra_args": "",
        },
    }


class Config:
    def __init__(self):
        self.data = default_config()
        self.lock = threading.Lock()
        self.load()

    def load(self):
        try:
            with open(CONFIG_FILE, "r", encoding="utf-8") as f:
                saved = json.load(f)
            self.data = default_config()
            for k, v in saved.items():
                if k in self.data:
                    self.data[k] = v
            if not isinstance(self.data["presets"], dict):
                self.data["presets"] = {}
            if not isinstance(self.data["scan_dirs"], list):
                self.data["scan_dirs"] = []
        except Exception:
            pass

    def save(self):
        with self.lock:
            try:
                with open(CONFIG_FILE, "w", encoding="utf-8") as f:
                    json.dump(self.data, f, ensure_ascii=False, indent=2)
            except Exception:
                pass


class ServerManager:
    def __init__(self):
        self.proc = None
        self.pid = None
        self.port = None
        self.host = None
        self.model = None
        self.started_at = 0.0
        self.health = "off"          # off | loading | ok | error
        self.model_info = None
        self._logs = []
        self._lock = threading.Lock()
        self._stop = threading.Event()
        self._dead_logged = False
        self._stopped_intentionally = False

    # ---------- logs ----------
    def _log(self, line):
        with self._lock:
            self._logs.append(line)
            if len(self._logs) > MAX_LOG_LINES:
                del self._logs[: len(self._logs) - MAX_LOG_LINES]

    def get_logs_since(self, idx):
        with self._lock:
            return self._logs[idx:], len(self._logs)

    # ---------- ciclo de vida ----------
    def start(self, args, model, host, port):
        if self.proc is not None and self.proc.poll() is None:
            return False, "El servidor ya está en ejecución"
        if not os.path.isfile(SERVER_EXE):
            return False, f"No se encontró {SERVER_EXE}"
        if not os.path.isfile(model):
            return False, f"No se encontró el modelo:\n{model}"
        self._stop = threading.Event()
        self._dead_logged = False
        self._stopped_intentionally = False
        cmd = [SERVER_EXE] + args
        try:
            self.proc = subprocess.Popen(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                stdin=subprocess.DEVNULL,
                text=True,
                encoding="utf-8",
                errors="replace",
                bufsize=1,
                creationflags=CREATE_NO_WINDOW,
            )
        except Exception as e:
            return False, f"Error al iniciar el proceso: {e}"
        self.pid = self.proc.pid
        self.port = port
        self.host = host
        self.model = model
        self.started_at = time.time()
        self.health = "loading"
        self.model_info = None
        self._log("")
        self._log("=== llama-server iniciado (PID %d, puerto %d) ===" % (self.pid, port))
        self._log("Modelo: %s" % model)
        self._log("Comando: %s" % self._redact(" ".join(cmd)))
        threading.Thread(target=self._read_stdout, daemon=True).start()
        threading.Thread(target=self._health_loop, daemon=True).start()
        return True, "Servidor iniciado (PID %d)" % self.pid

    @staticmethod
    def _redact(cmdline):
        parts = cmdline.split(" ")
        for i, p in enumerate(parts):
            if p in ("--api-key", "--hf-token") and i + 1 < len(parts):
                parts[i + 1] = "****"
        return " ".join(parts)

    def _read_stdout(self):
        try:
            for line in self.proc.stdout:
                line = line.rstrip("\n")
                if line:
                    self._log(line)
        except Exception:
            pass

    def _health_loop(self):
        while not self._stop.is_set():
            time.sleep(2.5)
            if self.proc is None or self.proc.poll() is not None:
                self._on_dead()
                break
            try:
                conn = HTTPConnection(self.host if self.host not in ("", "0.0.0.0") else "127.0.0.1",
                                      self.port, timeout=3)
                conn.request("GET", "/health")
                resp = conn.getresponse()
                data = resp.read()
                conn.close()
                if resp.status == 200:
                    if self.health != "ok":
                        self._log("Servidor listo (health OK)")
                    self.health = "ok"
                    if not self.model_info:
                        self._fetch_model_info()
                else:
                    self.health = "loading"
            except Exception:
                self.health = "loading"

    def _fetch_model_info(self):
        try:
            conn = HTTPConnection("127.0.0.1", self.port, timeout=3)
            conn.request("GET", "/v1/models")
            resp = conn.getresponse()
            data = json.loads(resp.read().decode("utf-8", "replace"))
            conn.close()
            if data.get("data"):
                self.model_info = data["data"][0]
        except Exception:
            pass

    def _on_dead(self):
        if self._stopped_intentionally:
            self.health = "off"
            return
        code = self.proc.poll() if self.proc else None
        if not self._dead_logged:
            self._dead_logged = True
            if code in (None, 0):
                self._log("=== proceso terminado ===")
            else:
                self._log("=== proceso terminado con código de salida %s ===" % code)
        self.health = "error"

    def stop(self):
        self._stopped_intentionally = True
        pid = self.pid
        if pid is None and self.port:
            pid = self._find_pid_on_port(self.port)
        if pid is None:
            return False, "No hay servidor en ejecución"
        self._stop.set()
        try:
            subprocess.run(["taskkill", "/PID", str(pid), "/T", "/F"],
                           capture_output=True, timeout=10,
                           creationflags=CREATE_NO_WINDOW)
        except Exception:
            try:
                if self.proc:
                    self.proc.terminate()
            except Exception:
                pass
        self.proc = None
        self.pid = None
        self.health = "off"
        self._log("=== servidor detenido ===")
        return True, "Servidor detenido"

    @staticmethod
    def _find_pid_on_port(port):
        try:
            out = subprocess.run(["netstat", "-ano"], capture_output=True,
                                 text=True, timeout=15,
                                 creationflags=CREATE_NO_WINDOW).stdout
            for line in out.splitlines():
                parts = line.split()
                if (len(parts) >= 5 and parts[0].lower() == "tcp"
                        and parts[1].endswith(":%d" % port) and parts[3] == "LISTENING"):
                    try:
                        return int(parts[4])
                    except ValueError:
                        continue
        except Exception:
            pass
        return None

    def status(self):
        running = self.proc is not None and self.proc.poll() is None
        if not running and self.port:
            # el admin puede haberse reiniciado: detectar un llama-server vivo en el puerto
            try:
                conn = HTTPConnection("127.0.0.1", self.port, timeout=1.5)
                conn.request("GET", "/health")
                resp = conn.getresponse()
                resp.read()
                conn.close()
                if resp.status == 200:
                    running = True
                    self.health = "ok"
                    if not self.model_info:
                        self._fetch_model_info()
                    if self.model is None:
                        self.model = self.cfg.data.get("last_settings", {}).get("model")
            except Exception:
                if self.health != "off":
                    self.health = "error"
        return {
            "running": running,
            "pid": self.pid,
            "port": self.port,
            "host": self.host,
            "model": os.path.basename(self.model) if self.model else None,
            "health": self.health,
            "uptime": round(time.time() - self.started_at) if running and self.pid else 0,
            "model_info": self.model_info,
        }


class ToolManager:
    """Ejecuta una herramienta de bin/ (llama-bench, llama-quantize, llama-cli)
    capturando su salida en un buffer compartido con la interfaz web."""

    def __init__(self):
        self.proc = None
        self.pid = None
        self.tool = None
        self._logs = []
        self._lock = threading.Lock()

    def _log(self, line):
        with self._lock:
            self._logs.append(line)
            if len(self._logs) > MAX_LOG_LINES:
                del self._logs[: len(self._logs) - MAX_LOG_LINES]

    def clear_logs(self):
        with self._lock:
            self._logs.clear()

    def get_logs_since(self, idx):
        with self._lock:
            return self._logs[idx:], len(self._logs)

    @staticmethod
    def _redact(cmdline):
        parts = cmdline.split(" ")
        for i, p in enumerate(parts):
            if p in ("--api-key", "--hf-token") and i + 1 < len(parts):
                parts[i + 1] = "****"
        return " ".join(parts)

    def run(self, exe_name, args, label):
        if self.proc is not None and self.proc.poll() is None:
            return False, "Ya hay una herramienta en ejecución"
        exe = os.path.join(BIN_DIR, exe_name)
        if not os.path.isfile(exe):
            return False, "No se encontró %s" % exe
        try:
            self.proc = subprocess.Popen(
                [exe] + list(args),
                stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                stdin=subprocess.DEVNULL, text=True, encoding="utf-8",
                errors="replace", bufsize=1, creationflags=CREATE_NO_WINDOW,
                cwd=BIN_DIR,
            )
        except Exception as e:
            return False, "Error al iniciar el proceso: %s" % e
        self.pid = self.proc.pid
        self.tool = label
        self._log("")
        self._log("=== %s iniciado (PID %d) ===" % (label, self.pid))
        self._log("Comando: %s" % self._redact(" ".join([exe] + list(args))))
        threading.Thread(target=self._read, daemon=True).start()
        return True, "%s iniciado (PID %d)" % (label, self.pid)

    def _read(self):
        try:
            for line in self.proc.stdout:
                line = line.rstrip("\n")
                if line:
                    self._log(line)
        except Exception:
            pass
        try:
            code = self.proc.wait(timeout=2)
        except Exception:
            code = None
        if code in (None, 0):
            self._log("=== %s terminado ===" % self.tool)
        else:
            self._log("=== %s terminado con código de salida %s ===" % (self.tool, code))

    def stop(self):
        pid = self.pid
        if pid is None:
            return False, "No hay herramienta en ejecución"
        try:
            subprocess.run(["taskkill", "/PID", str(pid), "/T", "/F"],
                           capture_output=True, timeout=10,
                           creationflags=CREATE_NO_WINDOW)
        except Exception:
            try:
                if self.proc:
                    self.proc.terminate()
            except Exception:
                pass
        self.proc = None
        self.pid = None
        self._log("=== herramienta detenida por el usuario ===")
        return True, "Herramienta detenida"

    def status(self):
        running = self.proc is not None and self.proc.poll() is None
        return {"running": running, "pid": self.pid, "tool": self.tool}


TOOL_EXES = {
    "bench": "llama-bench.exe",
    "quantize": "llama-quantize.exe",
    "cli": "llama-cli.exe",
}
TOOL_LABELS = {
    "bench": "llama-bench",
    "quantize": "llama-quantize",
    "cli": "llama-cli",
}


def parse_tokenize_output(text):
    """Parsea la salida de llama-tokenize (formato: '  123 -> 'token'')."""
    tokens = []
    lines = text.splitlines()
    i = 0
    while i < len(lines):
        m = re.match(r"^\s*(\d+)\s*->\s*'", lines[i])
        if m:
            tok_id = int(m.group(1))
            rest = lines[i].split("'", 1)[1]
            if rest.endswith("'") and not rest.endswith("\\'"):
                content = rest[:-1]
            else:
                buf = rest
                i += 1
                while i < len(lines):
                    buf += "\n" + lines[i]
                    if lines[i].rstrip().endswith("'"):
                        content = buf[:-1]
                        break
                    i += 1
                else:
                    content = buf
            tokens.append({"id": tok_id, "token": content.replace("\n", "⏎")})
        i += 1
    m2 = re.search(r"total number of tokens:\s*(\d+)", text, re.IGNORECASE)
    count = int(m2.group(1)) if m2 else len(tokens)
    return tokens, count


# ---------------------------------------------------------------------------
# Detección de hardware y configuración automática
# ---------------------------------------------------------------------------

def total_ram_gb():
    """RAM total en GB usando GlobalMemoryStatusEx (sin procesos externos)."""
    if os.name != "nt":
        return None
    try:
        class MEMORYSTATUSEX(ctypes.Structure):
            _fields_ = [
                ("dwLength", ctypes.c_ulong),
                ("dwMemoryLoad", ctypes.c_ulong),
                ("ullTotalPhys", ctypes.c_ulonglong),
                ("ullAvailPhys", ctypes.c_ulonglong),
                ("ullTotalPageFile", ctypes.c_ulonglong),
                ("ullAvailPageFile", ctypes.c_ulonglong),
                ("ullTotalVirtual", ctypes.c_ulonglong),
                ("ullAvailVirtual", ctypes.c_ulonglong),
                ("ullAvailExtendedVirtual", ctypes.c_ulonglong),
            ]
        m = MEMORYSTATUSEX()
        m.dwLength = ctypes.sizeof(MEMORYSTATUSEX)
        if ctypes.windll.kernel32.GlobalMemoryStatusEx(ctypes.byref(m)):
            return m.ullTotalPhys / (1024.0 ** 3)
    except Exception:
        pass
    return None


class HardwareDetector:
    """Detecta CPU/RAM/GPU con caché de ~10 s."""

    def __init__(self):
        self.cache = {}
        self.lock = threading.Lock()

    def get(self, force=False):
        now = time.time()
        with self.lock:
            if not force and self.cache and now - self.cache.get("t", 0) < 10:
                return self.cache
        info = self._detect()
        info["t"] = now
        with self.lock:
            self.cache = info
        return info

    def _detect(self):
        cpu_threads = os.cpu_count() or 1
        cpu_name = os.environ.get("PROCESSOR_IDENTIFIER", platform.processor() or "desconocido")
        ram = total_ram_gb()
        gpus = self._detect_gpus()
        return {
            "cpu": {"name": cpu_name, "threads": cpu_threads},
            "ram_gb": round(ram, 1) if ram else None,
            "gpus": gpus,
        }

    def _detect_gpus(self):
        gpus = []
        # 1) nvidia-smi (VRAM exacta)
        try:
            out = subprocess.run(
                ["nvidia-smi", "--query-gpu=name,memory.total,memory.free",
                 "--format=csv,noheader,nounits"],
                capture_output=True, text=True, timeout=10,
                creationflags=CREATE_NO_WINDOW).stdout
            for line in out.splitlines():
                parts = [p.strip() for p in line.split(",")]
                if len(parts) >= 3:
                    try:
                        gpus.append({
                            "name": parts[0],
                            "vram_total_gb": round(int(parts[1]) / 1024, 1),
                            "vram_free_gb": round(int(parts[2]) / 1024, 1),
                            "usable": True,
                        })
                    except ValueError:
                        continue
            if gpus:
                return gpus
        except Exception:
            pass
        # 2) llama-server --list-devices (detecta CUDA/Metal que ve llama.cpp)
        try:
            out = subprocess.run([SERVER_EXE, "--list-devices"],
                                 capture_output=True, text=True, timeout=20,
                                 creationflags=CREATE_NO_WINDOW, cwd=BIN_DIR)
            text = out.stdout or out.stderr or ""
            for line in text.splitlines():
                m = re.match(r"^\s*(\w+\d*):\s+(.+?)\s*\(VRAM:\s*(\d+)\s*MiB", line)
                if m:
                    gpus.append({
                        "name": "%s: %s" % (m.group(1), m.group(2)),
                        "vram_total_gb": round(int(m.group(3)) / 1024, 1),
                        "vram_free_gb": round(int(m.group(3)) / 1024, 1),
                        "usable": True,
                    })
            if gpus:
                return gpus
        except Exception:
            pass
        # 3) Win32_VideoController (solo nombres, sin VRAM fiable)
        try:
            out = subprocess.run(
                ["powershell", "-NoProfile", "-Command",
                 "(Get-CimInstance Win32_VideoController).Name"],
                capture_output=True, text=True, timeout=15,
                creationflags=CREATE_NO_WINDOW)
            for line in out.stdout.splitlines():
                name = line.strip()
                if name and name.lower() not in ("name",) and "microsoft basic" not in name.lower():
                    gpus.append({"name": name, "vram_total_gb": None, "vram_free_gb": None,
                                 "usable": False})
        except Exception:
            pass
        return gpus


GGUF_VALUE_SIZES = {0: 1, 1: 1, 2: 2, 3: 2, 4: 4, 5: 4, 6: 4, 7: 1,
                    10: 8, 11: 8, 12: 8}


def read_gguf_meta(path):
    """Lee metadata básica de un GGUF (arquitectura, capas, ctx, parámetros)
    sin cargar tensores. Devuelve dict con valores o None si no se puede."""
    meta = {"arch": None, "n_layers": None, "ctx_train": None,
            "n_params": None, "ftype": None}
    try:
        with open(path, "rb") as f:
            if f.read(4) != b"GGUF":
                return meta
            f.read(4)          # version u32
            f.read(8)          # tensor_count u64
            kv_count = int.from_bytes(f.read(8), "little")

            def read_str():
                n = int.from_bytes(f.read(8), "little")
                return f.read(n).decode("utf-8", "replace")

            for _ in range(kv_count):
                try:
                    key = read_str()
                    vtype = int.from_bytes(f.read(4), "little")
                    if vtype == 8:
                        val = read_str()
                    elif vtype == 9:  # array
                        atype = int.from_bytes(f.read(4), "little")
                        alen = int.from_bytes(f.read(8), "little")
                        val = None
                        if atype == 8:
                            for _ in range(alen):
                                read_str()
                        else:
                            size = GGUF_VALUE_SIZES.get(atype, 1)
                            f.seek(size * alen, os.SEEK_CUR)
                    elif vtype in (10, 11):
                        val = int.from_bytes(f.read(8), "little", signed=(vtype == 11))
                    elif vtype in (4, 5):
                        val = int.from_bytes(f.read(4), "little", signed=(vtype == 5))
                    elif vtype == 6:
                        val = struct.unpack("<f", f.read(4))[0]
                    elif vtype == 7:
                        val = bool(f.read(1)[0])
                    elif vtype == 12:
                        val = struct.unpack("<d", f.read(8))[0]
                    elif vtype in GGUF_VALUE_SIZES:
                        f.read(GGUF_VALUE_SIZES[vtype])
                        val = None
                    else:
                        val = None
                except Exception:
                    break
                if key == "general.architecture" and isinstance(val, str):
                    meta["arch"] = val
                elif key == "general.parameter_count" and isinstance(val, int):
                    meta["n_params"] = val
                elif key == "general.file_type" and isinstance(val, int):
                    meta["ftype"] = val
                elif meta["arch"] and key == "%s.block_count" % meta["arch"] and isinstance(val, int):
                    meta["n_layers"] = val
                elif meta["arch"] and key == "%s.context_length" % meta["arch"] and isinstance(val, int):
                    meta["ctx_train"] = val
    except Exception:
        pass
    return meta


AUTO_TXT = {
    "es": {
        "threads_gpu": "Hilos: %(t)d (máx. 16 cuando se usa GPU)",
        "threads_cpu": "Hilos: %(t)d (núcleos detectados)",
        "ctx_ram": "Contexto: %(ctx)d (según RAM de %(ram).1f GB)",
        "ngl_all": "Capas en GPU: all (VRAM %(vram).1f GB libre para modelo de %(size).1f GB)",
        "ngl_partial": "Capas en GPU: %(ngl)d de %(n)d (VRAM %(vram).1f GB libre, %(per).2f GB por capa)",
        "ngl_zero_vram": "Capas en GPU: 0 (VRAM insuficiente: %(vram).1f GB libre, modelo de %(size).1f GB)",
        "ngl_no_gpu": "Capas en GPU: 0 (sin GPU usable detectada)",
        "flash_on": "Flash attention: on (GPU presente)",
        "flash_auto": "Flash attention: auto",
        "slots_1": "Slots: 1 (RAM limitada)",
        "slots_auto": "Slots: auto",
        "model_info": "Modelo: %(size).2f GB en disco",
        "unknown": "No se pudo leer la metadata del modelo",
    },
    "en": {
        "threads_gpu": "Threads: %(t)d (max 16 when using GPU)",
        "threads_cpu": "Threads: %(t)d (detected cores)",
        "ctx_ram": "Context: %(ctx)d (based on %(ram).1f GB RAM)",
        "ngl_all": "GPU layers: all (%(vram).1f GB VRAM free for a %(size).1f GB model)",
        "ngl_partial": "GPU layers: %(ngl)d of %(n)d (%(vram).1f GB VRAM free, %(per).2f GB per layer)",
        "ngl_zero_vram": "GPU layers: 0 (not enough VRAM: %(vram).1f GB free, model is %(size).1f GB)",
        "ngl_no_gpu": "GPU layers: 0 (no usable GPU detected)",
        "flash_on": "Flash attention: on (GPU present)",
        "flash_auto": "Flash attention: auto",
        "slots_1": "Slots: 1 (limited RAM)",
        "slots_auto": "Slots: auto",
        "model_info": "Model: %.2f GB on disk",
        "unknown": "Could not read model metadata",
    },
}


def _fmt(lang, key, **kw):
    d = AUTO_TXT.get(lang) or AUTO_TXT["es"]
    try:
        return d.get(key, "") % kw
    except (KeyError, TypeError, ValueError):
        return d.get(key, key)


def auto_config(model_path, lang="es"):
    """Calcula los argumentos óptimos para ejecutar el modelo en esta PC."""
    hw = hardware.get()
    meta = read_gguf_meta(model_path)
    args = []
    expl = []

    size_gb = None
    try:
        size_gb = os.path.getsize(model_path) / (1024.0 ** 3)
    except OSError:
        pass

    if not meta.get("n_layers"):
        expl.append(_fmt(lang, "unknown"))

    threads = hw["cpu"].get("threads") or 1
    usable_gpus = [g for g in hw["gpus"] if g.get("usable") and g.get("vram_free_gb") is not None]
    if usable_gpus:
        threads = min(threads, 16)
        expl.append(_fmt(lang, "threads_gpu", t=threads))
    else:
        expl.append(_fmt(lang, "threads_cpu", t=threads))
    args += ["-t", str(threads)]

    ram = hw.get("ram_gb") or 8.0
    if ram >= 32:
        ctx = 8192
    elif ram >= 12:
        ctx = 4096
    elif ram >= 8:
        ctx = 2048
    else:
        ctx = 1024
    if meta.get("ctx_train"):
        ctx = min(ctx, meta["ctx_train"])
    args += ["-c", str(ctx)]
    expl.append(_fmt(lang, "ctx_ram", ctx=ctx, ram=ram))

    ngl = 0
    flash = "auto"
    if usable_gpus and size_gb:
        vram_free = max(g.get("vram_free_gb") or 0 for g in usable_gpus)
        vram_total = max(g.get("vram_total_gb") or 0 for g in usable_gpus)
        n_layers = meta.get("n_layers") or 32
        if vram_total and vram_free and vram_free >= size_gb * 1.25:
            ngl = "all"
            flash = "on"
            expl.append(_fmt(lang, "ngl_all", vram=vram_free, size=size_gb))
        elif vram_free > 0.5:
            per_layer = size_gb / max(n_layers, 1)
            kv_gb = (ctx / 4096.0) * (n_layers / 32.0)
            usable = vram_free - kv_gb - 0.5
            ngl = max(0, int(usable / per_layer)) if per_layer > 0 else 0
            flash = "on" if ngl > 0 else "auto"
            if ngl > 0:
                expl.append(_fmt(lang, "ngl_partial", ngl=ngl, n=n_layers,
                                 vram=vram_free, per=per_layer))
            else:
                expl.append(_fmt(lang, "ngl_zero_vram", vram=vram_free, size=size_gb))
    else:
        expl.append(_fmt(lang, "ngl_no_gpu"))

    args += ["-ngl", str(ngl)]
    if flash == "on":
        args += ["--flash-attn", "on"]
        expl.append(_fmt(lang, "flash_on"))
    else:
        expl.append(_fmt(lang, "flash_auto"))

    if not usable_gpus and ram < 16:
        args += ["-np", "1"]
        expl.append(_fmt(lang, "slots_1"))
    else:
        expl.append(_fmt(lang, "slots_auto"))

    return {
        "args": args,
        "explicacion": expl,
        "hardware": {k: v for k, v in hw.items() if k != "t"},
        "model_meta": meta,
        "model_size_gb": round(size_gb, 2) if size_gb else None,
    }


class ModelScanner:
    def __init__(self, cfg):
        self.cfg = cfg
        self.models = []
        self.scanning = False
        self.last_scan = 0.0
        self.lock = threading.Lock()

    def scan_dirs(self):
        dirs = list(self.cfg.data.get("scan_dirs", []))
        for d in DEFAULT_SCAN_DIRS:
            if d not in dirs:
                dirs.append(d)
        return [d for d in dirs if d and os.path.isdir(d)]

    def trigger(self):
        with self.lock:
            if self.scanning:
                return False
            self.scanning = True
        threading.Thread(target=self._scan_worker, daemon=True).start()
        return True

    def _scan_worker(self):
        try:
            found = {}
            for root in self.scan_dirs():
                self._walk(root, found, depth=0)
            with self.lock:
                self.models = sorted(found.values(), key=lambda m: m["name"].lower())
                self.last_scan = time.time()
        finally:
            with self.lock:
                self.scanning = False

    def _walk(self, root, found, depth):
        if depth > 5:
            return
        try:
            for entry in os.scandir(root):
                try:
                    if entry.is_dir(follow_symlinks=False):
                        low = entry.name.lower()
                        if entry.name.startswith(".") or any(s in low for s in SKIP_DIR_PARTS):
                            continue
                        self._walk(entry.path, found, depth + 1)
                    elif entry.is_file(follow_symlinks=True):
                        if entry.name.lower().endswith(".gguf"):
                            found[entry.path] = {
                                "path": entry.path,
                                "name": entry.name,
                                "dir": os.path.dirname(entry.path),
                                "size_gb": round(entry.stat().st_size / (1024 ** 3), 2),
                            }
                except OSError:
                    continue
        except OSError:
            pass

    def get(self):
        with self.lock:
            return list(self.models), self.scanning, self.last_scan


manager = ServerManager()
cfg = Config()
scanner = ModelScanner(cfg)
toolmgr = ToolManager()
hardware = HardwareDetector()
manager.cfg = cfg
ls = cfg.data.get("last_settings", {}) or {}
try:
    if ls.get("port"):
        manager.port = int(ls["port"])
    if ls.get("host"):
        manager.host = ls["host"]
    if ls.get("model"):
        manager.model = ls["model"]
except (TypeError, ValueError):
    pass


def scan_dirs_for_config():
    dirs = [d for d in scanner.scan_dirs() if d]
    return dirs


def build_server_args(s):
    args = []
    args += ["-m", s["model"]]
    args += ["--host", s.get("host") or "127.0.0.1"]
    try:
        args += ["--port", str(int(s.get("port") or 8080))]
    except (TypeError, ValueError):
        return None, "El puerto debe ser un número"
    ctx = (s.get("ctx") or "").strip()
    if ctx:
        args += ["-c", ctx]
    ngl = (s.get("ngl") or "auto").strip()
    if ngl:
        args += ["-ngl", ngl]
    threads = (s.get("threads") or "").strip()
    if threads:
        args += ["-t", threads]
    slots = (s.get("slots") or "").strip()
    if slots:
        args += ["-np", slots]
    temp = (s.get("temp") or "").strip()
    if temp:
        args += ["--temp", temp]
    top_p = (s.get("top_p") or "").strip()
    if top_p:
        args += ["--top-p", top_p]
    top_k = (s.get("top_k") or "").strip()
    if top_k:
        args += ["--top-k", top_k]
    repeat = (s.get("repeat") or "").strip()
    if repeat:
        args += ["--repeat-penalty", repeat]
    seed = (s.get("seed") or "").strip()
    if seed and seed != "-1":
        args += ["-s", seed]
    flash = (s.get("flash") or "auto").strip()
    if flash and flash != "auto":
        args += ["--flash-attn", flash]
    api_key = (s.get("api_key") or "").strip()
    if api_key:
        args += ["--api-key", api_key]
    extra = (s.get("extra_args") or "").strip()
    if extra:
        try:
            args += shlex.split(extra, posix=False)
        except ValueError as e:
            return None, "Argumentos extra inválidos: %s" % e
    return args, None


# ---------- actualización de llama.cpp ----------
GITHUB_REPO = "ggml-org/llama.cpp"
update_state = {"phase": "idle", "detail": "", "error": None}
_update_thread = None


def llama_version(exe_path):
    """Devuelve el número de build (ej: 5780) o None."""
    try:
        out = subprocess.run([exe_path, "--version"], capture_output=True, text=True,
                             timeout=15, creationflags=CREATE_NO_WINDOW, cwd=BIN_DIR)
        txt = (out.stdout or "") + "\n" + (out.stderr or "")
        m = re.search(r"version[:=]?\s*(\d+)", txt, re.IGNORECASE)
        if m:
            return int(m.group(1))
        m = re.search(r"build\s+(\d+)", txt, re.IGNORECASE)
        if m:
            return int(m.group(1))
    except Exception:
        pass
    return None


def _github_get(path):
    req = urllib.request.Request(
        "https://api.github.com/repos/%s/%s" % (GITHUB_REPO, path),
        headers={"User-Agent": "llama-admin/1.0",
                 "Accept": "application/vnd.github+json"})
    with urllib.request.urlopen(req, timeout=20) as r:
        return json.loads(r.read().decode("utf-8", "replace"))


def pick_asset(release):
    """Elige el zip Windows según la GPU detectada (cuda/avx2/…)."""
    prefs = []
    if any(g.get("usable") for g in hardware.get()["gpus"]):
        prefs += ["cuda", "vulkan", "avx512"]
    prefs += ["avx2", "avx", "cpu"]

    def score(name):
        n = name.lower()
        for i, p in enumerate(prefs):
            if p in n:
                return i
        return len(prefs)

    assets = [a for a in release.get("assets", [])
              if a.get("name", "").startswith("llama-")
              and "-bin-win-" in a.get("name", "")
              and a.get("name", "").endswith("-x64.zip")]
    if not assets:
        return None
    assets.sort(key=lambda a: score(a["name"]))
    return assets[0]


def _do_update(asset, build, tag):
    global update_state
    tmp_zip = os.path.join(BASE_DIR, "llama-update.zip")
    tmp_dir = os.path.join(BASE_DIR, ".update-tmp")
    try:
        update_state["phase"] = "downloading"
        update_state["detail"] = "Descargando %s…" % asset["name"]
        req = urllib.request.Request(asset["browser_download_url"],
                                     headers={"User-Agent": "llama-admin/1.0"})
        with urllib.request.urlopen(req, timeout=900) as r, open(tmp_zip, "wb") as f:
            shutil.copyfileobj(r, f, 256 * 1024)

        update_state["phase"] = "extracting"
        update_state["detail"] = "Descomprimiendo…"
        if os.path.isdir(tmp_dir):
            shutil.rmtree(tmp_dir, ignore_errors=True)
        os.makedirs(tmp_dir)
        with zipfile.ZipFile(tmp_zip) as z:
            z.extractall(tmp_dir)
        exe_dir = None
        for root, _, files in os.walk(tmp_dir):
            if "llama-server.exe" in files:
                exe_dir = root
                break
        if not exe_dir:
            raise RuntimeError("El paquete no contiene llama-server.exe")

        update_state["phase"] = "backup"
        update_state["detail"] = "Guardando copia de seguridad…"
        backup = os.path.join(BASE_DIR, "bin.backup-%s" % time.strftime("%Y%m%d-%H%M%S"))
        if os.path.isdir(BIN_DIR):
            shutil.move(BIN_DIR, backup)
        else:
            backup = None

        update_state["phase"] = "installing"
        update_state["detail"] = "Instalando…"
        shutil.copytree(exe_dir, BIN_DIR)
        os.remove(tmp_zip)
        shutil.rmtree(tmp_dir, ignore_errors=True)
        update_state["phase"] = "done"
        update_state["detail"] = "ok"
        update_state["error"] = None
        update_state["backup"] = os.path.basename(backup) if backup else None
    except Exception as e:
        update_state["phase"] = "error"
        update_state["error"] = str(e)
        update_state["detail"] = str(e)
        for p in (tmp_zip, tmp_dir):
            try:
                if os.path.isfile(p):
                    os.remove(p)
                elif os.path.isdir(p):
                    shutil.rmtree(p, ignore_errors=True)
            except OSError:
                pass


def start_update():
    global _update_thread
    release = _github_get("releases/latest")
    asset = pick_asset(release)
    if not asset:
        return False, "No se encontró un paquete Windows en la última versión"
    build = re.sub(r"\D", "", release.get("tag_name", ""))
    update_state.update({"phase": "starting", "detail": "Comenzando…", "error": None})
    _update_thread = threading.Thread(target=_do_update,
                                      args=(asset, build, release.get("tag_name", "")),
                                      daemon=True)
    _update_thread.start()
    return True, "Descarga iniciada (build %s)" % build


def find_backups():
    try:
        dirs = [d for d in os.listdir(BASE_DIR) if d.startswith("bin.backup-")]
        return sorted(dirs, reverse=True)
    except OSError:
        return []


def revert_update():
    backups = find_backups()
    if not backups:
        return False, "No hay copia de seguridad"
    if not os.path.isdir(BIN_DIR):
        return False, "bin no existe"
    backup = os.path.join(BASE_DIR, backups[0])
    failed = os.path.join(BASE_DIR, "bin.failed-%s" % time.strftime("%Y%m%d-%H%M%S"))
    try:
        shutil.move(BIN_DIR, failed)
        shutil.move(backup, BIN_DIR)
        return True, "Restaurada la copia %s" % backups[0]
    except OSError as e:
        return False, "Error al revertir: %s" % e


class Handler(BaseHTTPRequestHandler):
    server_version = "LlamaAdmin/1.0"
    protocol_version = "HTTP/1.1"

    def log_message(self, fmt, *args):
        pass

    # ---------- helpers ----------
    def _send_json(self, obj, status=200):
        data = json.dumps(obj, ensure_ascii=False).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(data)))
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        try:
            self.wfile.write(data)
        except (BrokenPipeError, ConnectionResetError):
            pass

    def _read_json(self):
        try:
            n = int(self.headers.get("Content-Length", 0))
        except (TypeError, ValueError):
            n = 0
        if n <= 0:
            return {}
        try:
            return json.loads(self.rfile.read(n).decode("utf-8-sig"))
        except Exception:
            return {}

    def _serve_static(self, path):
        rel = urllib.parse.unquote(path)
        if rel in ("/", ""):
            rel = "/index.html"
        rel = rel.lstrip("/")
        fp = os.path.normpath(os.path.join(STATIC_DIR, rel))
        if not fp.startswith(os.path.normpath(STATIC_DIR) + os.sep) or not os.path.isfile(fp):
            self._send_json({"error": "no encontrado"}, 404)
            return
        mime = mimetypes.guess_type(fp)[0] or "application/octet-stream"
        try:
            with open(fp, "rb") as f:
                data = f.read()
        except OSError:
            self._send_json({"error": "no encontrado"}, 404)
            return
        self.send_response(200)
        self.send_header("Content-Type", mime)
        self.send_header("Content-Length", str(len(data)))
        self.send_header("Cache-Control", "no-cache")
        self.end_headers()
        try:
            self.wfile.write(data)
        except (BrokenPipeError, ConnectionResetError):
            pass

    # ---------- GET ----------
    def do_GET(self):
        parsed = urllib.parse.urlparse(self.path)
        route = parsed.path

        if route == "/api/status":
            return self._send_json(manager.status())
        if route == "/api/logs":
            qs = urllib.parse.parse_qs(parsed.query)
            since = 0
            try:
                since = int(qs.get("since", ["0"])[0])
            except ValueError:
                pass
            lines, next_idx = manager.get_logs_since(max(since, 0))
            st = manager.status()
            return self._send_json({"lines": lines, "next": next_idx, "running": st["running"]})
        if route == "/api/models":
            qs = urllib.parse.parse_qs(parsed.query)
            fresh = qs.get("fresh", ["0"])[0] == "1"
            if fresh and not scanner.scanning:
                scanner.trigger()
            models, scanning, last = scanner.get()
            return self._send_json({
                "models": models, "scanning": scanning, "last_scan": last,
                "dirs": scan_dirs_for_config(),
            })
        if route == "/api/config":
            return self._send_json(cfg.data)
        if route == "/api/tools":
            tools = []
            if os.path.isdir(BIN_DIR):
                for f in sorted(os.listdir(BIN_DIR)):
                    if f.lower().endswith(".exe"):
                        p = os.path.join(BIN_DIR, f)
                        try:
                            size = os.path.getsize(p)
                        except OSError:
                            size = 0
                        tools.append({"name": f, "size_mb": round(size / (1024 ** 2), 1)})
            return self._send_json({
                "tools": tools,
                "bin_dir": BIN_DIR,
                "server_ok": os.path.isfile(SERVER_EXE),
            })
        if route == "/api/tool/status":
            return self._send_json(toolmgr.status())
        if route == "/api/hardware":
            return self._send_json(hardware.get())
        if route == "/api/auto-config":
            qs = urllib.parse.parse_qs(parsed.query)
            model = (qs.get("model", [""])[0]).strip()
            lang = (qs.get("lang", [""])[0]).strip() or "es"
            if not model or not os.path.isfile(model):
                return self._send_json({"ok": False, "error": "Modelo inválido"}, 400)
            return self._send_json({"ok": True, **auto_config(model, lang)})
        if route == "/api/tool/logs":
            qs = urllib.parse.parse_qs(parsed.query)
            since = 0
            try:
                since = int(qs.get("since", ["0"])[0])
            except ValueError:
                pass
            lines, next_idx = toolmgr.get_logs_since(max(since, 0))
            return self._send_json({"lines": lines, "next": next_idx, "running": toolmgr.status()["running"]})
        if route == "/api/update/check":
            current = llama_version(SERVER_EXE)
            try:
                release = _github_get("releases/latest")
            except Exception as e:
                return self._send_json({"ok": False, "error": "No se pudo contactar a GitHub: %s" % e}, 502)
            tag = release.get("tag_name", "")
            latest = int(re.sub(r"\D", "", tag) or 0)
            asset = pick_asset(release)
            return self._send_json({
                "ok": True,
                "current": current,
                "latest": latest,
                "latest_tag": tag,
                "up_to_date": bool(current and latest and current >= latest),
                "asset": {"name": asset["name"],
                          "size_mb": round((asset.get("size") or 0) / (1024 ** 2), 1)}
                          if asset else None,
                "backups": find_backups(),
            })
        if route == "/api/update/status":
            return self._send_json(dict(update_state))
        return self._serve_static(route)

    # ---------- POST ----------
    def do_POST(self):
        parsed = urllib.parse.urlparse(self.path)
        route = parsed.path
        body = self._read_json()

        if route == "/api/start":
            s = dict(cfg.data.get("last_settings", {}))
            s.update({k: v for k, v in body.items() if v is not None})
            s["model"] = (body.get("model") or "").strip()
            if not s["model"]:
                return self._send_json({"ok": False, "error": "Selecciona un modelo"}, 400)
            explicacion = None
            if body.get("auto"):
                auto = auto_config(s["model"], (body.get("lang") or "es"))
                base_args, err = build_server_args({
                    "model": s["model"],
                    "host": s.get("host"), "port": s.get("port"),
                    "api_key": s.get("api_key"), "extra_args": s.get("extra_args"),
                    "ctx": "", "ngl": "", "threads": "", "slots": "",
                    "temp": "", "top_p": "", "top_k": "", "repeat": "", "seed": "", "flash": "",
                })
                if err:
                    return self._send_json({"ok": False, "error": err}, 400)
                cleaned, skip = [], False
                for a in base_args:
                    if a == "-ngl":
                        skip = True
                        continue
                    if skip:
                        skip = False
                        continue
                    cleaned.append(a)
                args = cleaned + auto["args"]
                explicacion = auto["explicacion"]
            else:
                args, err = build_server_args(s)
                if err:
                    return self._send_json({"ok": False, "error": err}, 400)
            cfg.data["last_settings"] = s
            cfg.save()
            ok, msg = manager.start(args, s["model"], s.get("host") or "127.0.0.1",
                                    int(s.get("port") or 8080))
            return self._send_json({"ok": ok, "message": msg, "explicacion": explicacion,
                                    "status": manager.status()},
                                   200 if ok else 400)

        if route == "/api/stop":
            ok, msg = manager.stop()
            return self._send_json({"ok": ok, "message": msg, "status": manager.status()})

        if route == "/api/chat":
            return self._proxy_chat(body)

        if route == "/api/tool/run":
            tool = body.get("tool")
            args = body.get("args")
            if tool not in TOOL_EXES or not isinstance(args, list):
                return self._send_json({"ok": False, "error": "herramienta o argumentos inválidos"}, 400)
            toolmgr.clear_logs()
            ok, msg = toolmgr.run(TOOL_EXES[tool], [str(a) for a in args], TOOL_LABELS[tool])
            return self._send_json({"ok": ok, "message": msg, "status": toolmgr.status()},
                                   200 if ok else 400)

        if route == "/api/tool/stop":
            ok, msg = toolmgr.stop()
            return self._send_json({"ok": ok, "message": msg, "status": toolmgr.status()})

        if route == "/api/tool/terminal":
            tool = body.get("tool")
            args = body.get("args")
            if tool not in TOOL_EXES or not isinstance(args, list):
                return self._send_json({"ok": False, "error": "herramienta o argumentos inválidos"}, 400)
            exe = os.path.join(BIN_DIR, TOOL_EXES[tool])
            if not os.path.isfile(exe):
                return self._send_json({"ok": False, "error": "No se encontró %s" % exe}, 400)
            try:
                subprocess.Popen([exe] + [str(a) for a in args],
                                 cwd=BIN_DIR, creationflags=CREATE_NEW_CONSOLE)
            except Exception as e:
                return self._send_json({"ok": False, "error": str(e)}, 400)
            return self._send_json({"ok": True, "message": "Ventana de terminal abierta"})

        if route == "/api/update/run":
            if update_state["phase"] in ("downloading", "extracting", "installing", "backup"):
                return self._send_json({"ok": False, "error": "Ya hay una actualización en curso"}, 409)
            if manager.status()["running"]:
                return self._send_json({"ok": False,
                                        "error": "Detén el servidor antes de actualizar"}, 409)
            if toolmgr.status()["running"]:
                return self._send_json({"ok": False,
                                        "error": "Espera a que termine la herramienta en curso"}, 409)
            try:
                ok, msg = start_update()
            except Exception as e:
                return self._send_json({"ok": False, "error": str(e)}, 500)
            return self._send_json({"ok": ok, "message": msg}, 200 if ok else 400)

        if route == "/api/update/revert":
            if update_state["phase"] in ("downloading", "extracting", "installing", "backup"):
                return self._send_json({"ok": False, "error": "Actualización en curso"}, 409)
            ok, msg = revert_update()
            return self._send_json({"ok": ok, "message": msg}, 200 if ok else 400)

        if route == "/api/tokenize":
            model = (body.get("model") or "").strip()
            text = body.get("text") or ""
            if not model or not os.path.isfile(model):
                return self._send_json({"ok": False, "error": "Modelo inválido"}, 400)
            if not text:
                return self._send_json({"ok": False, "error": "Escribe algún texto para tokenizar"}, 400)
            exe = os.path.join(BIN_DIR, "llama-tokenize.exe")
            try:
                proc = subprocess.Popen(
                    [exe, "-m", model, "--stdin", "--show-count"],
                    stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                    stderr=subprocess.STDOUT, text=True, encoding="utf-8",
                    errors="replace", creationflags=CREATE_NO_WINDOW, cwd=BIN_DIR,
                )
                out, _ = proc.communicate(text, timeout=120)
            except subprocess.TimeoutExpired:
                proc.kill()
                return self._send_json({"ok": False, "error": "Tiempo de espera agotado"}, 500)
            except Exception as e:
                return self._send_json({"ok": False, "error": str(e)}, 500)
            tokens, count = parse_tokenize_output(out)
            return self._send_json({"ok": True, "count": count, "tokens": tokens})

        if route == "/api/config":
            action = body.get("action")
            if action == "save_preset":
                name = (body.get("name") or "").strip()
                settings = body.get("settings")
                if not name or not isinstance(settings, dict):
                    return self._send_json({"ok": False, "error": "Nombre o configuración inválidos"}, 400)
                cfg.data["presets"][name] = settings
                cfg.save()
                return self._send_json({"ok": True, "presets": cfg.data["presets"]})
            if action == "delete_preset":
                name = body.get("name")
                cfg.data["presets"].pop(name, None)
                cfg.save()
                return self._send_json({"ok": True, "presets": cfg.data["presets"]})
            if action == "save_last":
                settings = body.get("settings")
                if isinstance(settings, dict):
                    cfg.data["last_settings"] = settings
                    cfg.save()
                return self._send_json({"ok": True})
            if action == "set_scan_dirs":
                dirs = body.get("dirs")
                if isinstance(dirs, list):
                    cfg.data["scan_dirs"] = [d.strip() for d in dirs if d.strip()]
                    cfg.save()
                    scanner.trigger()
                return self._send_json({"ok": True})
            return self._send_json({"ok": False, "error": "acción desconocida"}, 400)

        return self._send_json({"ok": False, "error": "ruta desconocida"}, 404)

    # ---------- proxy de chat ----------
    def _proxy_chat(self, body):
        st = manager.status()
        if not st["running"]:
            return self._send_json({"ok": False, "error": "El servidor no está en ejecución"}, 502)
        port = st["port"]
        try:
            payload = json.dumps(body).encode("utf-8")
            conn = HTTPConnection("127.0.0.1", port, timeout=60)
            conn.request("POST", "/v1/chat/completions", body=payload, headers={
                "Content-Type": "application/json",
                "Accept": "text/event-stream",
            })
            resp = conn.getresponse()
        except Exception as e:
            return self._send_json({"ok": False, "error": "No se pudo conectar con llama-server: %s" % e}, 502)
        ctype = resp.getheader("Content-Type", "text/event-stream")
        self.send_response(resp.status)
        self.send_header("Content-Type", ctype)
        self.send_header("Cache-Control", "no-cache")
        self.send_header("Connection", "close")
        self.end_headers()
        try:
            while True:
                chunk = resp.read(65536)
                if not chunk:
                    break
                self.wfile.write(chunk)
                self.wfile.flush()
        except (BrokenPipeError, ConnectionResetError, OSError):
            pass
        finally:
            conn.close()


def main():
    port = 8756
    browse = False
    for arg in sys.argv[1:]:
        if arg == "--browse":
            browse = True
        else:
            try:
                port = int(arg)
            except ValueError:
                pass
    server = ThreadingHTTPServer(("127.0.0.1", port), Handler)
    server.daemon_threads = True
    url = "http://127.0.0.1:%d" % port
    print("=" * 56)
    print("  Llama Admin — Administrador web de llama.cpp")
    print("  Abre la interfaz en: %s" % url)
    print("  Presiona Ctrl+C para detener el administrador")
    print("=" * 56)
    if browse:
        threading.Timer(0.8, lambda: webbrowser.open(url)).start()
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nDeteniendo administrador...")
        if manager.status()["running"]:
            manager.stop()
        server.shutdown()


if __name__ == "__main__":
    main()

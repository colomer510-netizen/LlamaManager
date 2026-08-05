import os
import platform
import ctypes
import threading
import time
import subprocess
import re
import struct

from backend.paths import get_app_dir

CREATE_NO_WINDOW = 0x08000000

def get_bin_dir():
    return os.path.join(get_app_dir(), "bin")

def get_server_exe():
    return os.path.join(get_bin_dir(), "llama-server.exe")

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

        try:
            out = subprocess.run([get_server_exe(), "--list-devices"],
                                 capture_output=True, text=True, timeout=20,
                                 creationflags=CREATE_NO_WINDOW, cwd=get_bin_dir())
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

hardware = HardwareDetector()

GGUF_VALUE_SIZES = {0: 1, 1: 1, 2: 2, 3: 2, 4: 4, 5: 4, 6: 4, 7: 1,
                    10: 8, 11: 8, 12: 8}

def read_gguf_meta(path):
    meta = {"arch": None, "n_layers": None, "ctx_train": None,
            "n_params": None, "ftype": None}
    try:
        with open(path, "rb") as f:
            if f.read(4) != b"GGUF":
                return meta
            f.read(4)          
            f.read(8)          
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
                    elif vtype == 9:  
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

def get_auto_config(model_path, lang="es"):
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

import os
import asyncio
import psutil
from backend.paths import get_data_dir
from backend.services.binary_resolver import get_binary_path

class ToolManager:
    def __init__(self):
        self.bench_process = None
        self.bench_log = os.path.join(get_data_dir(), "bench.log")
        self.quantize_process = None
        self.quantize_log = os.path.join(get_data_dir(), "quantize.log")
        self.convert_process = None
        self.convert_log = os.path.join(get_data_dir(), "convert.log")
        
    async def run_benchmark(self, model: str, threads: int, ngl: int, prompt_tokens: int, gen_tokens: int, binary_strategy: str = "auto"):
        if self.bench_process and self.bench_process.returncode is None:
            return False # Already running
            
        exe = get_binary_path("llama-bench", binary_strategy)
        if not os.path.exists(exe):
            raise Exception("llama-bench no encontrado en la carpeta bin/ ni en el PATH del sistema")
            
        # Empty previous log
        with open(self.bench_log, "w", encoding="utf-8") as f:
            f.write("Iniciando Benchmark...\n")
            
        args = [
            exe,
            "-m", model,
            "-t", str(threads),
            "-ngl", str(ngl),
            "-p", str(prompt_tokens),
            "-n", str(gen_tokens),
            "-r", "1" # Solo una repetición para que sea rápido (según plan)
        ]
        
        with open(self.bench_log, "a", encoding="utf-8") as f:
            f.write(f"Comando: {' '.join(args)}\n")

        self.bench_process = await asyncio.create_subprocess_exec(
            *args,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.STDOUT
        )
        
        # Start a background task to stream output to file
        asyncio.create_task(self._log_output(self.bench_process, self.bench_log))
        return True
        
    async def stop_benchmark(self):
        if self.bench_process and self.bench_process.returncode is None:
            try:
                parent = psutil.Process(self.bench_process.pid)
                for child in parent.children(recursive=True):
                    child.kill()
                parent.kill()
                with open(self.bench_log, "a", encoding="utf-8") as f:
                    f.write("\nBenchmark detenido por el usuario.\n")
                return True
            except psutil.NoSuchProcess:
                pass
        return False

    async def run_quantize(self, input_model: str, output_model: str, method: str, binary_strategy: str = "auto"):
        if self.quantize_process and self.quantize_process.returncode is None:
            return False # Already running
            
        exe = get_binary_path("llama-quantize", binary_strategy)
        if not os.path.exists(exe):
            raise Exception("llama-quantize no encontrado en la carpeta bin/ ni en el PATH del sistema")
            
        with open(self.quantize_log, "w", encoding="utf-8") as f:
            f.write("Iniciando Cuantización...\n")
            
        args = [
            exe,
            input_model,
            output_model,
            method
        ]
        
        with open(self.quantize_log, "a", encoding="utf-8") as f:
            f.write(f"Comando: {' '.join(args)}\n")

        self.quantize_process = await asyncio.create_subprocess_exec(
            *args,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.STDOUT
        )
        
        asyncio.create_task(self._log_output(self.quantize_process, self.quantize_log))
        return True
        
    async def stop_quantize(self):
        if self.quantize_process and self.quantize_process.returncode is None:
            try:
                parent = psutil.Process(self.quantize_process.pid)
                for child in parent.children(recursive=True):
                    child.kill()
                parent.kill()
                with open(self.quantize_log, "a", encoding="utf-8") as f:
                    f.write("\nCuantización detenida por el usuario.\n")
                return True
            except psutil.NoSuchProcess:
                pass
        return False
        
    async def run_convert(self, model_dir: str, outtype: str, output_path: str = ""):
        if self.convert_process and self.convert_process.returncode is None:
            return False # Already running
            
        from backend.services.converter import start_conversion
        
        with open(self.convert_log, "w", encoding="utf-8") as f:
            f.write("Iniciando Conversión...\nDependiendo de tu conexión, instalar las librerías necesarias puede tomar unos minutos la primera vez.\n")
            
        try:
            self.convert_process = await start_conversion(model_dir, outtype, output_path)
            asyncio.create_task(self._log_output(self.convert_process, self.convert_log))
            return True
        except Exception as e:
            with open(self.convert_log, "a", encoding="utf-8") as f:
                f.write(f"\nError al iniciar: {str(e)}\n")
            return False

    async def stop_convert(self):
        if self.convert_process and self.convert_process.returncode is None:
            try:
                parent = psutil.Process(self.convert_process.pid)
                for child in parent.children(recursive=True):
                    child.kill()
                parent.kill()
                with open(self.convert_log, "a", encoding="utf-8") as f:
                    f.write("\nConversión detenida por el usuario.\n")
                return True
            except psutil.NoSuchProcess:
                pass
        return False

    async def _log_output(self, process, log_file):
        with open(log_file, "a", encoding="utf-8") as f:
            while True:
                line = await process.stdout.readline()
                if not line:
                    break
                decoded = line.decode('utf-8', errors='ignore')
                f.write(decoded)
                f.flush()

tool_manager = ToolManager()

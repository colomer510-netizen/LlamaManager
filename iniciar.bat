@echo off
title Llama Admin Moderno - Administrador de llama.cpp
cd /d "%~dp0"

echo [1/2] Comprobando entorno virtual y dependencias...
if not exist .venv (
    echo Creando entorno virtual...
    python -m venv .venv
    .venv\Scripts\pip install -r requirements.txt
)

echo [2/2] Iniciando servidor web con FastAPI...
start http://127.0.0.1:8756
.venv\Scripts\uvicorn backend.main:app --host 127.0.0.1 --port 8756 --reload
pause

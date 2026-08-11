@echo off
title Llama Admin Pro
cd /d "%~dp0"

echo [1/2] Checking virtual environment...
if not exist .venv (
    echo Creating virtual environment...
    python -m venv .venv
    .venv\Scripts\pip install fastapi uvicorn pydantic
)

echo [2/2] Starting Llama Admin Pro on http://127.0.0.1:8756 ...
start http://127.0.0.1:8756
.venv\Scripts\uvicorn backend.main:app --host 127.0.0.1 --port 8756 --reload
pause

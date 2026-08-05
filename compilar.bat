@echo off
title Compilador Llama Admin
cd /d "%~dp0"

echo [1/3] Preparando entorno para compilar...
if not exist .venv (
    echo Creando entorno virtual temporal...
    python -m venv .venv
)

echo Instalando dependencias necesarias...
.venv\Scripts\pip install -r requirements.txt
.venv\Scripts\pip install pyinstaller

echo.
echo [2/3] Generando ejecutable portátil...
:: Limpiamos compilaciones anteriores
rmdir /s /q build dist 2>nul
del /q LlamaAdmin.spec 2>nul

:: Llamamos a PyInstaller
.venv\Scripts\pyinstaller --noconfirm ^
    --onedir ^
    --windowed ^
    --icon=NONE ^
    --name="LlamaAdmin" ^
    --add-data="static;static" ^
    --add-data="bin;bin" ^
    --hidden-import="uvicorn" ^
    --hidden-import="uvicorn.logging" ^
    --hidden-import="uvicorn.loops" ^
    --hidden-import="uvicorn.loops.auto" ^
    --hidden-import="uvicorn.protocols" ^
    --hidden-import="uvicorn.protocols.http" ^
    --hidden-import="uvicorn.protocols.http.auto" ^
    --hidden-import="uvicorn.protocols.websockets" ^
    --hidden-import="uvicorn.protocols.websockets.auto" ^
    --hidden-import="uvicorn.lifespan" ^
    --hidden-import="uvicorn.lifespan.on" ^
    --hidden-import="fastapi" ^
    backend\main.py

echo.
echo [3/3] Listo. El ejecutable se encuentra en la carpeta 'dist\LlamaAdmin\'
pause

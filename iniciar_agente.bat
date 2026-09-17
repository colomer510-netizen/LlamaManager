@echo off
echo ==========================================
echo     Iniciando Agente Inteligente...
echo ==========================================
cd /d "%~dp0"
if exist "bin\agent.exe" (
    bin\agent.exe
) else if exist "agent.exe" (
    agent.exe
) else (
    echo Error: No se encontro el ejecutable del agente en bin\agent.exe o agent.exe
    pause
)

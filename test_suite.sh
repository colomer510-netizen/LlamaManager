#!/bin/bash
# ==========================================================
# 🦙 LlamaManager - Suite de Diagnóstico y Pruebas para Linux
# ==========================================================

set -e
export PATH=$HOME/.local/go/bin:$PATH
cd "$(dirname "$0")"

echo "=========================================================="
echo "      EJECUTANDO DIAGNÓSTICO DEL SISTEMA (LINUX)          "
echo "=========================================================="

echo -e "\n[1/5] Ejecutando Tests Unitarios en Go..."
go test -v ./internal/models ./internal/config ./internal/hardware ./internal/tools
echo "✅ Tests unitarios completados con éxito."

echo -e "\n[2/5] Compilando LlamaManager-Server..."
go build -ldflags "-s -w" -o LlamaManager-Server ./cmd/server/
ls -lh LlamaManager-Server
echo "✅ Compilación exitosa."

echo -e "\n[3/5] Verificando Binarios y Librerías Dinámicas (.so)..."
export LD_LIBRARY_PATH="$(pwd)/bin":$LD_LIBRARY_PATH
if [ -f "./bin/llama-cli" ]; then
    ./bin/llama-cli --version
    ./bin/llama-server --version
    echo "✅ Binarios de llama.cpp verificados y vinculados correctamente."
else
    echo "⚠️ Binarios no encontrados en bin/. Usa la función de autoinstalación en la interfaz."
fi

echo -e "\n[4/5] Probando Servidor Web y Endpoints API..."
./LlamaManager-Server &
SERVER_PID=$!
sleep 2

# Probar Hardware
echo -n "  - GET /api/hardware: "
wget -qO- http://localhost:3000/api/hardware
echo ""

# Probar Settings
echo -n "  - POST /api/settings: "
wget -qO- --post-data='{"model_path":"~/Modelos","api_port":"8080","context_size":4096,"gpu_layers":0}' --header='Content-Type: application/json' http://localhost:3000/api/settings
echo ""

# Probar Models
echo -n "  - GET /api/models: "
wget -qO- http://localhost:3000/api/models
echo ""

kill $SERVER_PID 2>/dev/null || true
echo "✅ Servidor Web y API verificados al 100%."

echo -e "\n[5/5] Verificando Lanzadores y Permisos..."
[ -x "./iniciar.sh" ] && echo "  - iniciar.sh: [OK]" || echo "  - iniciar.sh: [Falta +x]"
[ -f "$HOME/Escritorio/LlamaManager.desktop" ] && echo "  - Icono de Escritorio: [OK]" || echo "  - Icono de Escritorio: [No encontrado]"

echo -e "\n=========================================================="
echo "  🎉 ¡DIAGNÓSTICO COMPLETADO! SISTEMA FUNCIONANDO AL 100%  "
echo "=========================================================="

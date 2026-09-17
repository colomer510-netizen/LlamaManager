#!/bin/bash
# Script para iniciar el Agente Inteligente independientemente de LlamaManager

# Obtener la ruta del directorio donde está el script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "=========================================="
echo "    Iniciando Agente Inteligente..."
echo "=========================================="

cd "$DIR"

# Verificar si el ejecutable del agente existe
if [ -f "./bin/agent" ]; then
    # Ejecutar el agente directamente
    ./bin/agent
else
    echo "Error: No se encontró el ejecutable del agente en ./bin/agent"
    echo "Asegúrate de haber compilado el agente primero."
    echo "Presiona Enter para salir..."
    read
fi

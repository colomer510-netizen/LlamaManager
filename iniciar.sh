#!/bin/bash
# Script de inicio para LlamaManager en Linux
cd "$(dirname "$0")"

echo "==================================================="
echo "  Mantén esta ventana abierta mientras uses la IA"
echo "  (Puedes minimizarla, pero NO la cierres)"
echo "==================================================="
echo ""

# Ejecutamos el servidor normalmente (sin background)
./LlamaManager-Server

#!/bin/bash
# Script de inicio silencioso para LlamaManager en Linux
cd "$(dirname "$0")"

# Ejecutar el servidor en segundo plano
# Todo el log y control se maneja ahora desde la interfaz web (localhost:3000)
nohup ./LlamaManager-Server > /dev/null 2>&1 &

# Salir inmediatamente para que la ventana de la terminal se cierre automáticamente
exit 0

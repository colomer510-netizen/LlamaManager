# 🦙 LlamaManager - Terminal & Super Agent Edition (V6.0)

**Gestor de Terminal Profesional, Autónomo y Portable para Modelos GGUF con Integración Oficial de llama.cpp y Super Agente Linux.**

Bienvenido a la carpeta especial **`cli-edition/`** de LlamaManager. Esta versión está diseñada para entornos de alto rendimiento, servidores, desarrolladores y usuarios avanzados que desean administrar modelos locales de Inteligencia Artificial directamente desde la consola, sin dependencias pesadas, sin necesidad de instalación tradicional y con soporte para un **Super Agente Autónomo de Diagnóstico Linux**.

---

## ✨ Características Principales

* 🤖 **Super Agente Copiloto Linux:**
  * **Telemetría viva:** Inspecciona automáticamente kernel (`uname`), memoria RAM (`free -h`), espacio en disco principal (`df -h /`), núcleos de procesador (`nproc`) y tiempo encendido (`uptime`).
  * **Sugerencia proactiva de comandos:** Si la tarea lo requiere, propone el comando en Bash exacto.
  * **Escudo de Seguridad:** Filtra y bloquea comandos potencialmente destructivos (`rm -rf`, `mkfs`, `dd`, etc.).
  * **Bucle de Realimentación:** Ejecuta el comando tras confirmación del usuario y reinyecta la salida en el contexto de la IA para un análisis final.
  * **Streaming SSE en Tiempo Real:** Respuestas instantáneas palabra por palabra (incluso con modelos de razonamiento como DeepSeek-R1).
* 🛡️ **Auto-Fallback GPU a CPU (Aislamiento de Drivers):**
  * Inyección automática de variables vacías (`GGML_VK_VISIBLE_DEVICES=""`, `CUDA_VISIBLE_DEVICES=""`) al configurar 0 capas de GPU para proteger el sistema ante crasheos de drivers Vulkan/Nvidia.
* ⚡ **Test de Estrés y Benchmark Comparativo:**
  * Diagnóstico de subsistencia de CPU (`-ngl 0`) y GPU (`-ngl 99`) para determinar estabilidad y velocidad relativa.
* 👻 **Arranque Silencioso con Gestión de PID Desacoplado (`Setpgid`):**
  * Ejecución de servidores API en segundo plano sin terminal invasiva, inmune a señales de cierre.
  * Detección dinámica al iniciar y botón de parada de emergencia para liberar memoria RAM al instante.
* 🌐 **Compatibilidad Total OpenAI / Bionic / Web:**
  * Servidor HTTP local estándar en puerto configurable (por defecto `:8080`) con soporte de red local (`0.0.0.0`) y autenticación por API Key opcional.
* 💬 **Chat Interactivo Nativo (`llama cli`):**
  * Sesión conversacional directa en la consola con renderizado a color y soporte para modelos modernos.
* 📥 **Descarga Directa desde HuggingFace:**
  * Búsqueda y descarga de archivos `.gguf` con el motor oficial integrado.
* 🖥️ **Lanzadores de Un Solo Clic:**
  * Scripts `.sh` y accesos directos `.desktop` para escritorio y menú de aplicaciones.

---

## 🏗️ Arquitectura del Sistema CLI

```text
LlamaManager CLI Edition
├── 1. Core Engine (Binario nativo ~/.local/bin/llama)
├── 2. Hardware Scanner (/proc/meminfo, CPU, VRAM)
├── 3. Persistent Settings (llama_settings.json)
├── 4. Background Daemon (PID Manager /tmp/gestor_llama_server.pid con Setpgid)
├── 5. Super Agent Engine (Telemetría en vivo + SSE Streaming + Bash Execution)
└── 6. Launchers (iniciar.sh + gestor-llama.desktop)
```

---

## 🚀 Compilación y Uso

```bash
# 1. Clonar el repositorio
git clone https://github.com/colomer510-netizen/LlamaManager.git
cd LlamaManager/cli-edition

# 2. Compilar el ejecutable portable en Go
go build -o gestor-llama main.go
chmod +x gestor-llama

# 3. Iniciar el gestor
./gestor-llama
```

---

## 📋 Menú de Navegación

| Opción | Módulo | Descripción |
|---|---|---|
| `[1]` | **Servidor API** | Despliega el servidor OpenAI-compatible en segundo plano o interactivo. |
| `[2]` | **Chat CLI Estándar** | Conversación directa con el modelo en terminal (`llama cli`). |
| `[3]` | **Super Agente Linux** | Copiloto autónomo que analiza tu hardware, propone soluciones y ejecuta comandos supervisados. |
| `[4]` | **Test de Estrés / Benchmark** | Evaluación comparativa de rendimiento entre CPU y GPU. |
| `[5]` | **Descargador HuggingFace** | Descarga cualquier modelo `.gguf` por repositorio y nombre de archivo. |
| `[6]` | **Ajustes y Preferencias** | Configuración permanente de puertos, contexto, GPU y servidores RPC. |
| `[7]` | **Actualizador Oficial** | Actualiza el motor de `llama.cpp` a la última versión disponible. |
| `[8]` | **Crear Accesos Directos** | Genera lanzadores de doble clic `.sh` y `.desktop`. |
| `[9]` | **Consola en Vivo** | Visualiza en tiempo real los logs del servidor silencioso (`tail -f`). |
| `[10]`| **Apagar Servidor** | Termina el proceso fantasma y libera la RAM del sistema. |

---

*Desarrollado con pasión para la comunidad de código abierto por [Enoc Colomer](https://github.com/colomer510-netizen).*

# Llama Admin — Administrador de llama.cpp

Interfaz web local (un solo archivo de backend) para gestionar [llama.cpp](https://github.com/ggml-org/llama.cpp)
en Windows: inicia y detiene `llama-server`, chatea con tus modelos GGUF, ejecuta benchmarks,
convierte modelos y mantiene los binarios actualizados — **sin instalar nada más que Python**.

## Características

| Pestaña | Qué hace |
|---|---|
| ⚙️ Servidor | Iniciar/detener `llama-server`, logs en vivo, presets, escaneo de modelos GGUF, **⚡ inicio automático** según recursos |
| 💬 Chat | Chat streaming con el modelo en ejecución, temp/máx. tokens ajustables, abortar generación |
| ⚡ Benchmark | Ejecuta `llama-bench` (tok/s de prompt y generación) con botón **⚡ Auto** que rellena los parámetros según tu hardware |
| 🔧 Utilidades | `llama-cli` (generación única o terminal interactiva), `llama-tokenize`, `llama-quantize` y **actualización de llama.cpp desde GitHub** |

### Extras

- **⚡ Inicio automático**: detecta CPU, RAM y VRAM (vía `nvidia-smi` / `--list-devices`) y calcula
  los argumentos óptimos (`-t`, `-c`, `-ngl`, `-np`, flash-attn) leyendo la metadata del GGUF.
  Explica cada decisión antes de iniciar.
- **Temas**: oscuro, claro o automático (sigue al sistema).
- **Idiomas**: español / inglés (toda la interfaz y las explicaciones automáticas).
- **Actualización integrada**: compara tu versión con la última release oficial de GitHub, descarga
  el paquete adecuado (CUDA/Vulkan si hay GPU, si no CPU), hace copia de seguridad y permite revertir.
- **Sin dependencias**: solo librería estándar de Python.

## Requisitos

- **Windows** (64 bits)
- **Python 3.9+** (probado con 3.14)
- **Binarios de llama.cpp** en la carpeta `bin\` (ver [Instalación](#instalación))

## Instalación

1. **Descarga los binarios de llama.cpp** para Windows desde
   [GitHub releases](https://github.com/ggml-org/llama.cpp/releases) (elige el paquete según tu CPU/GPU,
   por ejemplo `llama-*-bin-win-cpu-x64.zip` o `-cuda-*`).
   Extrae el contenido en la carpeta `bin\` del proyecto, de modo que exista `bin\llama-server.exe`.
2. **Coloca tus modelos GGUF** en cualquier carpeta (por defecto busca en el escritorio, descargas,
   documentos y raíces de disco; puedes configurar las rutas desde la interfaz).

> El botón **🔍 Comprobar actualización** de la pestaña Utilidades hace este paso 1 automáticamente
> desde la propia interfaz.

## Uso

**Doble clic en `iniciar.bat`** (o manualmente):

```
python app.py [puerto] [--browse]
```

- Por defecto sirve en `http://127.0.0.1:8756`; `--browse` abre el navegador automáticamente.
- El administrador **detiene `llama-server` al cerrarse** con Ctrl+C.

### Flujo rápido

1. Abre la interfaz → pestaña **Servidor**.
2. Elige el modelo (o pulsa 🔍 Buscar).
3. Pulsa **⚡ Inicio automático** → revisa las decisiones → **Iniciar con estos valores**.
4. Cambia a **Chat** y conversa. Para medir velocidad, ve a **Benchmark**.

## Estructura del proyecto

```
administrador de llama.cpp/
├── app.py              # Backend completo: servidor HTTP + orquestación de procesos
├── iniciar.bat         # Lanzador (busca python/py y abre el navegador)
├── static/
│   └── index.html      # Interfaz web (SPA, JS vanilla, sin frameworks)
├── bin/                # Binarios de llama.cpp (llama-server, llama-cli, llama-bench…)
├── config.json         # Configuración persistente (se crea al primer uso)
└── plan-inicio-automatico.md  # Plan original de la función de auto-configuración
```

## Arquitectura

Aplicación web local monolítica de dos capas, **solo stdlib de Python**:

```
Navegador ──fetch / SSE──> app.py (127.0.0.1:8756) ──subprocess / HTTP──> llama-*.exe
```

- **Backend** (`app.py`, ~1.400 líneas): un `ThreadingHTTPServer` con enrutado REST por ruta.
  Componentes singletons:
  - `Config` — persistencia en `config.json` (ajustes, presets, directorios de búsqueda).
  - `ServerManager` — ciclo de vida de `llama-server`: lanza el proceso, captura stdout en un
    buffer circular y verifica salud con polling a `/health` (estados `off/loading/ok/error`).
    Detecta servidores huérfanos en el puerto tras reiniciar el administrador.
  - `ToolManager` — ejecuta `llama-bench`/`llama-cli`/`llama-quantize` con logs compartidos
    e índices incrementales.
  - `ModelScanner` — escanea discos buscando `.gguf` (con lista de carpetas a omitir).
  - `HardwareDetector` — CPU/RAM (API de Windows vía `ctypes`) y GPU (caché de 10 s).
  - `read_gguf_meta` / `auto_config` — parsea la cabecera GGUF y calcula argumentos óptimos.
  - Actualizador — consulta la GitHub API, elige paquete por GPU, descarga, respalda e instala
    en un hilo con estado consultable.
- **Frontend** (`static/index.html`): SPA de una página, JavaScript vanilla. Comunicación por
  `fetch` + polling (estado 3 s, herramientas 2 s, logs 1,5 s, progreso de actualización 1,5 s)
  y **SSE** para el chat. Internacionalización con diccionarios ES/EN y temas vía variables CSS.

### Decisiones de diseño

- Todo en un archivo backend para que la instalación sea copiar y ejecutar.
- Polling en lugar de WebSockets (simplicidad con stdlib).
- `config.json` como única persistencia (sin base de datos).
- Verificación de procesos por puerto (`netstat`) para no perder el estado al reiniciar el admin.
- La UI nunca ejecuta comandos directamente: todo pasa por la API (la interfaz solo ve JSON).

## API

| Endpoint | Método | Descripción |
|---|---|---|
| `/api/status` | GET | Estado del servidor (running, health, modelo, puerto, uptime, model_info) |
| `/api/logs?since=N` | GET | Logs incrementales del servidor |
| `/api/models?fresh=1` | GET | Modelos GGUF encontrados (con escaneo opcional) |
| `/api/config` | GET | Configuración completa (settings, presets, scan_dirs) |
| `/api/tools` | GET | Binarios `.exe` disponibles en `bin\` |
| `/api/hardware` | GET | CPU, RAM y GPUs detectadas (con `usable`/VRAM) |
| `/api/auto-config?model=…&lang=es` | GET | Argumentos óptimos + explicación en el idioma elegido |
| `/api/update/check` | GET | Versión local vs última release de GitHub |
| `/api/update/status` | GET | Progreso de una actualización en curso |
| `/api/tool/status` · `/api/tool/logs?since=N` | GET | Estado y logs de herramientas |
| `/api/start` | POST | Inicia el servidor; con `auto: true` usa auto-config; acepta `lang` |
| `/api/stop` | POST | Detiene el servidor |
| `/api/chat` | POST | Proxy streaming (SSE) hacia `/completion` del servidor |
| `/api/tool/run` · `/api/tool/stop` | POST | Ejecuta/detiene `bench`/`cli`/`quantize` con argumentos |
| `/api/tool/terminal` | POST | Abre terminal interactiva de `llama-cli` en ventana nueva |
| `/api/tokenize` | POST | Cuenta tokens con `llama-tokenize` |
| `/api/config` | POST | Acciones: `save_preset`, `delete_preset`, guardar directorios, guardar ajustes |
| `/api/update/run` | POST | Descarga e instala la última versión (rechaza si hay procesos en uso) |
| `/api/update/revert` | POST | Restaura la copia de seguridad anterior |

## Actualización de llama.cpp

1. Pestaña **Utilidades** → **🔍 Comprobar actualización**.
2. Si hay una versión nueva, pulsa **⬇ Descargar e instalar** (se bloquea si `llama-server` está
   en ejecución o una herramienta está activa).
3. Al terminar, los binarios anteriores quedan respaldados en `bin.backup-AAAAMMDD-HHMMSS`
   y puedes volver a ellos con **↩ Revertir**.

## Solución de problemas

| Problema | Solución |
|---|---|
| "No se encontró llama-server.exe" | Descarga los binarios oficiales y extráelos en `bin\` (o usa Comprobar actualización) |
| No aparecen modelos | Pulsa 🔍 Buscar o añade tus carpetas en "Directorios de búsqueda" |
| Inicio automático da `-ngl 0` | Es correcto si no hay GPU utilizable (las integradas se ignoran) |
| "Detén el servidor antes de actualizar" | Para el servidor en la pestaña Servidor antes de actualizar |
| El administrador no abre el navegador | Ejecuta `python app.py --browse` manualmente |

## Agradecimientos y Créditos

Este proyecto existe gracias al increíble trabajo de la comunidad de código abierto, y en especial al motor que lo impulsa:

* **[llama.cpp](https://github.com/ggml-org/llama.cpp)**: El asombroso motor de inferencia en C/C++ creado por Georgi Gerganov y la organización GGML, que permite ejecutar modelos de inteligencia artificial localmente con un rendimiento excepcional en casi cualquier hardware. Todos los binarios que orquesta este administrador pertenecen a su proyecto.

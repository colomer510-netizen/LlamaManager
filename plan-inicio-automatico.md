# Plan: Iniciar todo con 1 clic (configuración automática por recursos)

Fecha prevista: siguiente sesión.

## Objetivo

Añadir un botón **"⚡ Inicio automático"** (y equivalentes por herramienta) que
detecte los recursos de la PC (CPU, RAM, VRAM de GPU) y genere los argumentos
óptimos de `llama-server`, `llama-bench`, `llama-quantize`, etc., **sin que el
usuario configure nada**. El usuario solo elige el modelo y pulsa iniciar.

Idea central por herramienta:

| Herramienta | Qué se autoajusta |
|---|---|
| llama-server | hilos (-t), capas GPU (-ngl), contexto (-c), flash-attn (-fa), slots (-np) |
| llama-bench | hilos, -ngl, batch, n_prompt/n_gen según RAM |
| llama-quantize | hilos según CPU; aviso si origen ya está cuantizado |
| llama-cli | hilos y -ngl igual que el servidor |

## 1. Detección de recursos

Nuevo módulo en `app.py`: `hardware.py` (o clase `HardwareDetector`).

- **CPU**
  - `os.cpu_count()` → hilos totales.
  - Hilos recomendados: `min(cpu_count, 16)` si hay GPU, si no `cpu_count` para generation.
- **RAM disponible**
  - Windows: `Get-CimInstance Win32_OperatingSystem` → `TotalVisibleMemorySize`.
  - API propia: `GET /api/hardware` devuelve todo en JSON.
- **GPU (VRAM)**
  - `nvidia-smi --query-gpu=name,memory.total,memory.free --format=csv,noheader`
  - Si no existe `nvidia-smi` → mirar `wmic path win32_VideoController get name` (solo nombre, sin VRAM fiable).
  - Alternativa: `bin\llama-server.exe --list-devices` (corre rápido, listo para usar).
- **Modelo**
  - Tamaño del archivo GGUF (`os.path.getsize`).
  - Parámetros/layers: parsear metadata si es posible (o estimar por tamaño).

## 2. Heurísticas sugeridas (se pueden afinar mañana)

| Recurso | Regla |
|---|---|
| `-t` (hilos) | `min(cpu_count, 16)` en generación |
| `-ngl` (GPU) | Si `VRAM_usable >= tamaño_modelo * 1.25` → `all`; si no → `floor((VRAM_usable - KV_reserva) / (tamaño_modelo / n_layers))` con mínimo 0 |
| `-c` (contexto) | RAM ≥ 32 GB → 8192 · RAM ≥ 16 GB → 4096 · RAM ≥ 8 GB → 2048 · resto → 1024 |
| `-fa` (flash attn) | `on` si GPU detectada, si no `auto` |
| `-np` (slots) | 1 si RAM < 16 GB, si no `auto` |

KV_reserva aproximada: `~0.9 bytes por token por capa * ctx * n_layers` (ajustar con la práctica).

## 3. Cambios en `app.py`

- `GET /api/hardware` → `{ cpu: {cores, threads, nombre}, ram_gb, gpu: [{nombre, vram_total, vram_libre}] }`
  - cachear ~10 s para no relanzar `nvidia-smi` a cada rato.
- `POST /api/auto-config {model}` → `{ args: [...], explicacion: [lineas que digan "por qué"] }`
  - devuelve los argumentos recomendados + explicación legible para mostrar al usuario.
- En `/api/start`: aceptar `auto: true` → llama a `auto-config` y usa esos args.
- Misma lógica para bench/cli/quantize (parámetros mínimos en la petición).

## 4. Cambios en `static/index.html`

- Botón **"⚡ Inicio automático"** junto a "▶ Iniciar servidor".
- Al pulsarlo: pide modelo → `POST /api/auto-config` → muestra las decisiones
  (pills legibles: `hilos: 16 por CPU 8 núcleos`, `GPU: 1 capas por VRAM 6 GB`) →
  pregunta "¿Aplicar?" → inicia.
- Checkbox recordatorio: "Ajustar también temperatura/contexto" (opcional).
- En Benchmark: botón "⚡ Auto" que rellena hilos/ngl/batch con lo detectado.

## 5. Tareas ordenadas

- [x] `GET /api/hardware` con detección CPU/RAM/GPU + caché
- [x] Función cálculo de `-ngl` (leer metadata del GGUF para `n_layers`)
- [x] `POST /api/auto-config` para servidor
- [x] Botón "⚡ Inicio automático" en UI (Servidor)
- [x] Auto máx. contexto según RAM (tabla)
- [x] Aplicar la misma lógica a llama-cli y llama-bench
- [x] Mostrar "por qué" de cada ajuste al usuario
- [x] Pruebas en esta PC (CPU solament) y con GPU (si la hay)
- [x] Verificar con `nvidia-smi` ausente → fallback CPU

## 6. Pruebas a realizar mañana

1. [x] Sin GPU: `-ngl` queda en `0`, hilos al máximo útil, contexto según RAM → verificado (8 hilos, ctx 4096, ngl 0, slots 1)
2. [ ] Con GPU y modelo pequeño: `-ngl all`, flash attn `on` (pendiente: probar en máquina con GPU NVIDIA)
3. [ ] Modelo grande + VRAM poca: `-ngl` parcial calculado (pendiente)
4. [x] `--list-devices` funciona y `nvidia-smi` no existe → no rompe (esta PC: "Available devices: (none)")
5. [x] Inicio automático tarda < 2 s en calcular (solo lee cabecera GGUF, sin cargar modelo)

## Notas
- No requiere librerías extra (solo stdlib; para métricas de GPU se usa `nvidia-smi` que viene con los drivers).
- Si el usuario ya configuró valores manuales, respetarlos (el auto solo actúa si pide "auto").
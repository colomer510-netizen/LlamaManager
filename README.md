# 🦙 LlamaManager V4

**Gestor Inteligente, Ligero y Persistente para Modelos GGUF locales.**

LlamaManager es una interfaz y administrador de entorno diseñado para facilitar la ejecución, configuración y uso de modelos de inteligencia artificial locales (archivos `.gguf`) propulsados por el motor [llama.cpp](https://github.com/ggerganov/llama.cpp).

---

## ✨ Características Principales (V4)

* 🖥️ **GUI Nativa (Desktop)**: Interfaz de escritorio moderna y rápida construida con **Wails**. No dependes de un navegador web externo.
* 💬 **Chat Integrado (Nativo)**: Conversa con la IA mediante una interfaz de "globitos" estilo ChatGPT, todo dentro de la misma aplicación.
* 👻 **Ejecución Silenciosa**: El motor `llama-server.exe` corre 100% en segundo plano de manera invisible, sin consolas ni comandos molestos estorbando tu pantalla.
* 🌐 **Modo Servidor Ligero**: Incluye una versión alternativa de compilación (`cmd/server`) ideal para correr en servidores headless (Linux) y acceder remotamente vía navegador web.
* 📥 **Autoinstalador Inteligente**: Capaz de buscar, descargar y extraer automáticamente los últimos binarios oficiales de `llama.cpp` directamente desde GitHub.
* ⚙️ **Optimización Automática**: Escanea tu hardware (CPU, núcleos lógicos, RAM) para sugerir la mejor configuración de hilos.
* 🚀 **Soporte GPU**: Configura fácilmente el uso de tu tarjeta de video (NGL - Número de capas GPU) para acelerar drásticamente las respuestas.

---

## 🏗️ Arquitectura del Proyecto

El proyecto sigue el estándar profesional de desarrollo en Go (*Standard Go Project Layout*), separando completamente la interfaz de la lógica de negocio para permitir compilaciones multi-plataforma.

```text
📁 LlamaManager/
│
├── 📁 cmd/                         ← Aplicaciones compilables (Puntos de entrada)
│   ├── 📁 desktop/                 ← 🖥️ App Nativa Windows (Wails GUI)
│   │   ├── app.go                  ← Lógica puente (Go ↔ JS)
│   │   ├── main.go                 ← Inicialización de la ventana nativa
│   │   └── 📁 frontend/            ← Interfaz visual (HTML/CSS/JS)
│   │
│   └── 📁 server/                  ← 🌐 Servidor Web ligero (Multiplataforma)
│       └── main.go                 
│
├── 📁 internal/                    ← 🧠 NÚCLEO LÓGICO COMPARTIDO
│   ├── 📁 config/                  ← Gestión de settings persistentes (JSON)
│   ├── 📁 hardware/                ← Detección y análisis de CPU/RAM
│   ├── 📁 models/                  ← Escaneo local de archivos .gguf
│   └── 📁 web/                     ← Manejadores HTTP para la versión de Servidor
│
├── 📁 public/                      ← HTML/JS de la versión web tradicional
├── go.mod                          ← Dependencias
└── README.md
```

---

## 🛠️ Requisitos Previos (Para Desarrolladores)

Si deseas compilar el código fuente por ti mismo, necesitas:

1. **Go 1.21** o superior instalado en tu sistema.
2. El framework **Wails v2** instalado globalmente:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

---

## 🔨 Instrucciones de Compilación

Dependiendo de tus necesidades, LlamaManager puede compilarse de dos formas distintas:

### 1. Compilar Versión Desktop Nativa (Windows)
Esta es la versión principal con ventana propia y chat nativo.
```powershell
cd cmd/desktop
wails build -clean
```
> El ejecutable final se generará en: `cmd/desktop/build/bin/LlamaManager.exe`

### 2. Compilar Versión Servidor Web (Linux / Mac / Windows)
Ideal para entornos remotos o servidores sin interfaz gráfica nativa.
```powershell
# Compilar para el sistema actual
go build -ldflags "-s -w" -o LlamaManager-Server.exe ./cmd/server/

# Compilación cruzada (Desde Windows para Linux)
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -ldflags "-s -w" -o LlamaManager-Server-Linux ./cmd/server/
```

---

## 🚀 Uso de la Aplicación

1. **Ubicación del ejecutable:** Coloca el archivo `LlamaManager.exe` en tu carpeta deseada. La aplicación buscará (o descargará automáticamente) los archivos de `llama.cpp` en una subcarpeta llamada `bin/` junto a él.
2. **Ajustes:** En la pestaña de *Configuraciones*, define la ruta base donde guardas tus modelos `.gguf` (ej: `D:\OLLAMA AI\GGUF`).
3. **Lanzar:** Ve al *Lanzador*, selecciona tu modelo y haz clic en **"Iniciar y Abrir Chat"**. 
4. El sistema iniciará en las sombras y podrás conversar fluidamente con la IA.

---

## 🤝 Contribuciones
¡Las contribuciones, issues y pull requests son bienvenidos! Si encuentras un bug o tienes una idea para mejorar LlamaManager, siéntete libre de abrir un issue.

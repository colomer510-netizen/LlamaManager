# 🦙 LlamaManager V4.1

**Gestor Inteligente, Ligero y Persistente para Modelos GGUF locales.**

LlamaManager es una interfaz y administrador de entorno diseñado para facilitar la ejecución, configuración y uso de modelos de inteligencia artificial locales (archivos `.gguf`) propulsados por el motor [llama.cpp](https://github.com/ggerganov/llama.cpp).

---

## 📦 Descargas Rápidas (Releases)

En la carpeta `releases/` de este repositorio encontrarás las versiones listas para usar:
* `LlamaManager-Windows-V4.zip` - Versión lista para Windows (Incluye scripts de arranque).
* `LlamaManager-Linux-Ubuntu.tar.gz` - Versión lista para Ubuntu/Linux (Incluye acceso directo .desktop).

---

## ✨ Características Principales (V4.1)

* 👻 **Arranque 100% Silencioso**: El servidor de la IA arranca en las sombras. Nunca más verás una terminal negra estorbando en tu pantalla. Todo corre de fondo.
* 💚 **Consola Web Integrada (Novedad V4.1)**: Todos los registros (logs) del modelo y del sistema ahora se envían en tiempo real directamente a una interfaz de consola verde hacker dentro de la página web mediante tecnología Server-Sent Events (SSE).
* 💓 **Auto-Cierre Inteligente (Heartbeat)**: Si cierras el navegador, LlamaManager detectará que te fuiste (ausencia de latido) y tras 2 minutos matará automáticamente el proceso de la IA para liberar tu memoria RAM. ¡Cero procesos zombie!
* 🛡️ **Auto-Fallback GPU a CPU**: Si el optimizador de hardware establece `0` capas para la gráfica (GPU), LlamaManager inyecta variables de entorno protectoras (`GGML_VK_VISIBLE_DEVICES=""`, `CUDA_VISIBLE_DEVICES=""`) para desactivar Vulkan y prevenir que drivers gráficos defectuosos estrellen el sistema. 
* 📥 **Autoinstalador Inteligente**: Capaz de buscar, descargar y extraer automáticamente los últimos binarios oficiales de `llama.cpp` directamente desde GitHub.

---

## 🏗️ Arquitectura del Proyecto

El proyecto sigue el estándar profesional de desarrollo en Go (*Standard Go Project Layout*), separando completamente la interfaz de la lógica de negocio para permitir compilaciones multi-plataforma.

```text
📁 LlamaManager/
│
├── 📁 cmd/                         ← Aplicaciones compilables (Puntos de entrada)
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
├── 📁 releases/                    ← Binarios compilados listos para descargar
├── go.mod                          ← Dependencias
└── README.md
```

---

## 🔨 Instrucciones de Compilación (Si deseas compilarlo tú mismo)

1. **Go 1.21** o superior instalado en tu sistema.
2. Clona el repositorio y ejecuta:

### Linux
```bash
go build -o LlamaManager-Server cmd/server/main.go
```

### Windows (Cross-compilation desde Linux)
```bash
GOOS=windows GOARCH=amd64 go build -o LlamaManager-Server.exe cmd/server/main.go
```

---

## 🚀 Uso de la Aplicación

1. **Ubicación del ejecutable:** Coloca el archivo `LlamaManager` (o `LlamaManager-Server.exe`) en tu carpeta deseada junto con la carpeta `public/`.
2. **Ajustes:** En la web, ve a Configuración, define la ruta base donde guardas tus modelos `.gguf`.
3. **Lanzar:** Selecciona tu modelo y haz clic en **"Iniciar Servidor Local"**. 
4. El sistema iniciará en las sombras. Haz clic en "Abrir Chat" para conversar con la IA usando el puerto 8080.
5. Al terminar, **cierra el navegador** y la aplicación liberará los recursos por sí sola.

---

## 🤝 Contribuciones
¡Las contribuciones, issues y pull requests son bienvenidos! Si encuentras un bug o tienes una idea para mejorar LlamaManager, siéntete libre de abrir un issue.

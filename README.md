# LlamaManager 🦙

![Versión](https://img.shields.io/badge/Versi%C3%B3n-2.0.0-blue)
![Plataforma](https://img.shields.io/badge/Plataforma-Windows%20%7C%20Linux%20%7C%20macOS-success)
![Lenguaje](https://img.shields.io/badge/Lenguaje-Go%20%7C%20HTML%20%7C%20JS-00ADD8)

**LlamaManager** es un administrador avanzado y ligero para orquestar herramientas, chats y modelos locales de `llama.cpp`. 
Anteriormente escrito en PowerShell estricto para Windows, ahora ha evolucionado a una **arquitectura en Golang con una Interfaz Web nativa (Dashboard)**, haciéndolo 100% multiplataforma, infinitamente más escalable y resistente a fallos de hardware.

---

## ✨ Características Principales (V2.0)

* 🌐 **Interfaz Gráfica Web (GUI):** Olvídate de la consola negra. Al abrir el gestor, se levanta un servidor local ultra ligero que lanza una hermosa interfaz en tu navegador (consumiendo casi 0 RAM extra).
* 🛡️ **Vulkan Crash Fix (Instalador Automático):** ¿Tu GPU o drivers colapsan con el error `vkQueueSubmit: Invalid queue`? El gestor incluye un instalador automático que descarga e instala los binarios de `llama.cpp` en su versión de CPU Pura desde GitHub, esquivando totalmente a tu tarjeta de video defectuosa.
* 🧠 **Autoconfiguración Inteligente:** Detecta automáticamente tu cantidad de Memoria RAM y Núcleos de CPU para inyectar los mejores parámetros (`-c` y `-t`) antes de arrancar tu modelo.
* ⚙️ **Benchmark Diagnóstico Oficial:** Integración oculta con `llama-bench` para torturar tu CPU y GPU, midiendo con exactitud los Tokens por Segundo y determinando si tu gráfica es apta para correr IAs.
* 📂 **Buscador Dinámico de GGUF:** Escanea tu carpeta actual o cualquier ruta personalizada en tu disco duro para encontrar y lanzar tus modelos `.gguf` con un clic.

---

## 🚀 Instalación y Uso

### Prerrequisitos
Solo necesitas tener [Go instalado](https://go.dev/dl/) si deseas compilarlo desde el código fuente.

### Opción 1: Ejecutar desde el código (Desarrolladores)
1. Clona el repositorio:
   ```bash
   git clone https://github.com/colomer510-netizen/LlamaManager.git
   cd LlamaManager
   ```
2. Instala las dependencias y corre el programa:
   ```bash
   go mod tidy
   go run ./cmd/llama-manager
   ```

### Opción 2: Compilar el ejecutable
Si quieres empaquetarlo para compartirlo como un `.exe` (Windows) o un binario en Linux:
```bash
# Para Windows
go build -o llama-manager.exe ./cmd/llama-manager

# Para Linux/macOS
go build -o llama-manager ./cmd/llama-manager
```
Solo haz doble clic en el ejecutable resultante y tu navegador web se abrirá automáticamente.

---

## 🛠️ Estructura del Proyecto

El código ha sido refactorizado usando Clean Architecture en Go:

```text
LlamaManager/
├── cmd/
│   └── llama-manager/
│       └── main.go           # Punto de entrada de la aplicación
├── internal/
│   ├── hardware/             # Detección de CPU, RAM, y motor de Configuración
│   ├── models/               # Escáner de extensiones .gguf
│   ├── process/              # Lanzador de subprocesos (Server/CLI)
│   ├── tester/               # Sistema de testeo y benchmark
│   └── web/                  # Servidor HTTP y enrutador API REST
├── public/
│   └── index.html            # Interfaz Web (Vanilla HTML/CSS/JS)
└── go.mod                    # Dependencias de Go
```

---

## 💡 ¿Por qué Golang en lugar de PowerShell?
El proyecto original (`Universal-BAT-Script`) cumplió su propósito, pero PowerShell está limitado principalmente a ecosistemas Windows y su manejo de subprocesos asíncronos y servidores web es engorroso. Al migrar a **Go**, ganamos:
1. **Multiplataforma Real:** El mismo código se compila para Linux, Mac y Windows.
2. **Servidor HTTP Nativo:** Podemos servir una API y una página web sin dependencias externas (Ni Node.js, ni Python).
3. **Manejo Seguro de Errores:** Evitamos los crasheos silenciosos de la terminal CMD.

---

## 🤝 Contribuciones
¡Las contribuciones, issues y pull requests son bienvenidos! Si encuentras un bug o tienes una idea para mejorar LlamaManager, siéntete libre de abrir un issue.

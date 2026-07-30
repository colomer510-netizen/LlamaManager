# 🦙 LlamaManager para Windows

![Versión](https://img.shields.io/badge/Versión-4.0.0-blue.svg)
![Plataforma](https://img.shields.io/badge/Plataforma-Windows-lightgrey.svg)
![Licencia](https://img.shields.io/badge/Licencia-MIT-green.svg)

**LlamaManager** es una interfaz de línea de comandos (CLI) avanzada, interactiva y colorida, construida en PowerShell y empaquetada como `.exe` nativo para orquestar todas las herramientas del proyecto **[llama.cpp](https://github.com/ggerganov/llama.cpp)**. 

Elimina por completo la necesidad de escribir y memorizar comandos complejos, permitiéndote gestionar modelos, correr servidores, analizar documentos, cuantizar archivos y chatear con Inteligencia Artificial directamente desde un menú guiado.

---

## ✨ Novedades y Características Principales

- 🖱️ **Navegación Visual Interactiva:** Menús dinámicos controlados con las flechas del teclado.
- 🚀 **NUEVO - Ejecutable Nativo (.exe):** Ahora todo el sistema está empaquetado en un archivo `LlamaManager.exe` para mayor comodidad y para evadir los molestos bloqueos de scripts de PowerShell de Windows.
- 💬 **Modo Principiante (Chat Rápido):** ¿No quieres configurar hilos ni memoria? Usa el Modo Simple, elige un modelo y ¡listo!
- 🗃️ **NUEVO - Gestor de Modelos Avanzado:** Explora todos los modelos en tu disco duro, revisa su metadata, cópialos o elimínalos para liberar espacio directamente desde la interfaz.
- 💾 **Perfiles y Presets:** Guarda tus configuraciones favoritas (modelo, contexto, prompts) para ejecutarlas con un clic.
- 🧠 **Librería de System Prompts:** Elige el rol de tu IA con un clic (Asistente Útil, Programador Experto, Traductor, etc.) antes de empezar a chatear.
- ⚙️ **Auto-Offload Inteligente de GPU:** El sistema detecta automáticamente tu tarjeta gráfica (NVIDIA, AMD o Intel) y te recomienda usar `999` capas si detecta VRAM suficiente para procesar el modelo entero en tu GPU, acelerando todo al máximo.
- 📦 **NUEVO - Cuantizador por Lotes:** Puedes reducir el peso (cuantizar) modelos individuales, o apuntar a una carpeta entera y dejar que el Manager comprima todos tus archivos uno por uno de forma automática.
- 🌐 **Instalador Mágico y Auto-Updater:** Si te faltan los archivos de `llama.cpp`, el sistema se conecta a GitHub, descarga la última versión oficial para Windows y se autoconfigura en segundos. También cuenta con un actualizador manual con un solo clic y descargas directas de HuggingFace con barra de progreso.
- 🛡️ **Seguridad Grado Empresarial (v4.0.0):** Protección robusta contra inyección de comandos mediante el manejo nativo de arrays en PowerShell, asegurando que los nombres de archivos maliciosos no puedan ejecutar comandos arbitrarios en tu sistema.

---

## 🛠️ Instalación y Estructura

Para que LlamaManager funcione, simplemente necesitas descargar este repositorio.

### 1. Clonar o Descargar el repositorio
```bash
git clone https://github.com/tu-usuario/LlamaManager.git
cd LlamaManager
```

### 2. Estructura de carpetas requerida
```text
LlamaManager/
 ├── bin/                    <-- El Auto-Updater instalará llama.cpp aquí
 ├── models/                 <-- Descarga tus modelos .gguf de HuggingFace aquí
 ├── profiles/               <-- (Generado automáticamente por el script)
 ├── src/                    <-- Código fuente
 │    └── LlamaManager.ps1   <-- Código fuente original en PowerShell
 ├── assets/                 <-- Recursos gráficos e íconos
 ├── LlamaManager.exe        <-- ✨ Lanzador Principal (¡Doble clic aquí!)
 └── README.md
```

### 3. Ejecución Inicial (Magia 🪄)
Simplemente haz doble clic en **`LlamaManager.exe`**. 

Si es la primera vez que lo abres y no tienes el motor de `llama.cpp` instalado, el programa detectará que la carpeta `bin/` está vacía y te preguntará si deseas descargar automáticamente la última versión desde GitHub. ¡Dile que sí y el sistema hará el resto!

---

## 📸 Uso Rápido

Al abrir `LlamaManager.exe`, verás un menú principal con las siguientes opciones estrella:
1. **Ejecutar una herramienta (Avanzado)**: Configura opciones manuales (Contexto, Hilos, Semillas).
2. **Ejecutar (Modo Simple)**: Solo elige tu archivo `.gguf` y empieza a chatear.
3. **Buscar y Gestionar Modelos**: Un administrador interno para borrar modelos pesados y ver su metadata en un clic.
4. **Perfiles**: Repite ejecuciones anteriores al instante.
5. **Descargar de HuggingFace**: Pega una URL directa y mira la barra de progreso nativa descargar tus modelos directamente a tu carpeta `/models`.

---

## 🙏 Créditos y Atribución

Este proyecto es simplemente un orquestador. Todo el mérito de la ejecución, velocidad e inferencia de los modelos pertenece al increíble equipo y colaboradores de **`llama.cpp`**.

- **Georgi Gerganov y la comunidad de llama.cpp**: Por construir el motor de inferencia en C/C++ más eficiente del mundo. Visita el proyecto original en: [https://github.com/ggerganov/llama.cpp](https://github.com/ggerganov/llama.cpp)
- **Creado y mantenido por [colomer510-netizen](https://github.com/colomer510-netizen)**. Asistido y mejorado masivamente con Antigravity AI.

---

## 🤝 Cómo Contribuir

¡Las contribuciones son bienvenidas! Si eres desarrollador y quieres mejorar este proyecto (agregar nuevos wizards, mejorar la UI web, o arreglar bugs):

1. Haz un **Fork** de este repositorio.
2. Crea una rama para tu característica (`git checkout -b feature/NuevaCaracteristica`).
3. Haz **Commit** de tus cambios (`git commit -m 'Añadir nueva característica'`).
4. Sube los cambios a tu rama (`git push origin feature/NuevaCaracteristica`).
5. Abre un **Pull Request**.

Si tienes ideas o encuentras algún error, siéntete libre de abrir un **Issue**.

---
*Hecho para entusiastas de la IA Local. ¡Disfruta ejecutando modelos sin límites!*

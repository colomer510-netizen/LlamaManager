# 🦙 LlamaManager para Windows

![Versión](https://img.shields.io/badge/Versión-3.0.0-blue.svg)
![Plataforma](https://img.shields.io/badge/Plataforma-Windows%20PowerShell-lightgrey.svg)
![Licencia](https://img.shields.io/badge/Licencia-MIT-green.svg)

**LlamaManager** es una interfaz de línea de comandos (CLI) avanzada, interactiva y colorida, construida en PowerShell para orquestar todas las herramientas del proyecto **[llama.cpp](https://github.com/ggerganov/llama.cpp)**. 

Elimina por completo la necesidad de escribir y memorizar comandos complejos, permitiéndote gestionar modelos, correr servidores, analizar imágenes y chatear con Inteligencia Artificial directamente desde un menú guiado.

---

## ✨ Características Principales

- 🖱️ **Navegación Visual:** Menús interactivos controlados con las flechas del teclado.
- 💬 **Chat y Web UI:** Chatea en la terminal o lanza una interfaz web idéntica a ChatGPT conectada a tu servidor local.
- 💾 **Perfiles y Presets:** Guarda tus configuraciones favoritas (modelo, contexto, prompts) para ejecutarlas con un clic.
- 🕒 **Historial Automático:** El sistema recuerda tus últimos 50 comandos exitosos. Reutilízalos o lanza el modo "Chat Rápido".
- 👁️ **Soporte Multimodal y Audio:** Wizards especializados para modelos de visión (LLaVA) y texto-a-voz (TTS).
- ⚙️ **Detección Automática de Hardware:** Identifica si tienes GPU dedicada y ajusta las capas recomendadas (`-ngl`) para máxima velocidad.
- 🌐 **Auto-Updater:** Se conecta a GitHub, descarga la última versión oficial de `llama.cpp` y actualiza tus binarios de forma transparente.
- 📄 **Analizador de Documentos:** Procesa y resume archivos `.txt` enteros pasando el documento directamente al contexto del modelo.

---

## 🛠️ Instalación y Estructura

Para que LlamaManager funcione, necesitas descargar este repositorio y colocar los ejecutables oficiales de llama.cpp en la carpeta correcta.

### 1. Clonar el repositorio
Clona o descarga este repositorio en cualquier lugar de tu disco duro.
```bash
git clone https://github.com/tu-usuario/LlamaManager.git
cd LlamaManager
```

### 2. Estructura de carpetas requerida
Asegúrate de que la carpeta raíz contenga los archivos tal cual se muestra:
```text
LlamaManager/
 ├── bin/                    <-- ¡Coloca los .exe y .dll de llama.cpp aquí!
 ├── models/                 <-- (Opcional) Coloca tus archivos .gguf aquí
 ├── profiles/               <-- (Generado automáticamente por el script)
 ├── LlamaManager.ps1        <-- Código fuente principal
 ├── LlamaManager.bat        <-- Lanzador (Click para abrir)
 └── README.md
```

### 3. Descarga los binarios (Si no usas el Auto-Updater)
Puedes descargar la última release de [llama.cpp en GitHub](https://github.com/ggerganov/llama.cpp/releases) y extraer los `.exe` (como `llama-cli.exe`, `llama-server.exe`) y los `.dll` dentro de la carpeta `bin/`. 

*Nota: Alternativamente, si pones solo un ejecutable en `bin/`, puedes usar la opción de "Actualizar llama.cpp" del menú principal para que baje el resto automáticamente.*

### 4. Ejecución
Haz doble clic en **`LlamaManager.bat`**. ¡Eso es todo!

---

## 📸 Uso Rápido

Al abrir `LlamaManager.bat`, verás un menú similar a este:
1. **Ejecutar una herramienta**: Escoge entre Generación, Servidor, Cuantización, etc. Un wizard paso a paso te pedirá:
   - Archivo `.gguf` (usa el buscador integrado).
   - Longitud de contexto.
   - Hilos de CPU y GPU.
   - System Prompts (Elige entre "Asistente", "Programador", "Traductor", etc).
2. **Descargar modelo de HuggingFace**: Pega una URL directa a un archivo `.gguf` y se descargará directo a tu carpeta `models/`.
3. **Leer Metadata GGUF**: Averigua los parámetros, arquitectura y tokens base de cualquier archivo descargado sin cargarlo entero en memoria.

---

## 🙏 Créditos y Atribución

Este proyecto es simplemente un orquestador. Todo el mérito de la ejecución, velocidad e inferencia de los modelos pertenece al increíble equipo y colaboradores de **`llama.cpp`**.

- **Georgi Gerganov y la comunidad de llama.cpp**: Por construir el motor de inferencia en C/C++ más eficiente del mundo. Visita el proyecto original en: [https://github.com/ggerganov/llama.cpp](https://github.com/ggerganov/llama.cpp)
- **Creado y mantenido por [colomer510-netizen](https://github.com/colomer510-netizen)**. Asistido por Antigravity AI.

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

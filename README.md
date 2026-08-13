[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

# LlamaManager

LlamaManager es una herramienta para gestionar, desplegar y probar modelos LLaMA y variantes afines (p. ej. Llama 2). Proporciona utilidades para descargar pesos, configurar entornos, ejecutar inferencias locales y orquestar despliegues de pruebas.

Este README está escrito en español. Si prefieres la versión en inglés, se puede añadir en una futura actualización.

## Características

- Descarga y gestión de pesos de modelos LLaMA.
- Scripts de instalación y configuración de entornos (conda/venv/Docker).
- Ejecución de inferencias locales con ejemplos y notebooks.
- Soporte para ajustar parámetros de inferencia (temperatura, top_k, top_p, etc.).
- Plantillas para despliegue en servidores locales y contenedores.

## Requisitos

- Python 3.8+ (se recomienda 3.10+)
- Git
- (Opcional) CUDA 11.x / drivers compatibles para aceleración con GPU
- (Opcional) Docker, si desea usar contenedores

Asegúrate de tener suficiente espacio en disco para los pesos del modelo (varios GB según la variante).

## Instalación rápida

1. Clona el repositorio:

   git clone https://github.com/colomer510-netizen/LlamaManager.git
   cd LlamaManager

2. Crea un entorno virtual y activa:

   python -m venv .venv
   source .venv/bin/activate  # macOS / Linux
   .\.venv\Scripts\activate   # Windows (PowerShell)

3. Instala dependencias:

   pip install -r requirements.txt

4. Configura variables opcionales en `config.example.yaml` y crea `config.yaml` con tus valores.

## Uso

- Listar modelos disponibles (local/config):

  python -m llamamanager list-models

- Descargar un modelo:

  python -m llamamanager download --model llama-2-7b

- Ejecutar una inferencia de ejemplo:

  python -m llamamanager infer --model llama-2-7b --prompt "Hola, ¿cómo estás?"

- Levantar servicio local (ejemplo con FastAPI):

  python -m llamamanager serve --model llama-2-7b --host 0.0.0.0 --port 8000

(Estos comandos son ejemplos; consulta la sección `Referencia de comandos` o ejecutables en `./scripts` para la lista completa.)

## Configuración

Copia `config.example.yaml` a `config.yaml` y ajusta las opciones:

- model_storage_path: Ruta donde se guardarán los pesos descargados.
- default_model: Modelo por defecto para inferencias.
- gpu: true/false para habilitar uso de GPU cuando esté disponible.
- api_key: (opcional) claves para integraciones externas.

## Desarrollo

- Ejecuta tests:

  pytest

- Formatea el código con:

  black .

- Linting:

  flake8

Contribuciones son bienvenidas: abre un issue describiendo tu propuesta y crea un pull request cuando estés listo.

## Estructura del repositorio (resumen)

- ./llamamanager/        - código principal
- ./scripts/             - scripts de conveniencia (descarga, conversiones, deploy)
- ./examples/            - prompts y notebooks de ejemplo
- requirements.txt       - dependencias de Python
- config.example.yaml    - ejemplo de configuración

Si tu copia difiere, ajusta esta sección para reflejar la estructura real.

## Licencia

This project is licensed under the GNU General Public License v3.0 — see the LICENSE file for details.
Copyright (C) 2026 colomer510-netizen

## Contacto

Mantenedor: colomer510-netizen

Para preguntas, abre un issue en el repositorio.

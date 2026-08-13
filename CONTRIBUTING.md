# Contribuir — Guía rápida

Gracias por querer contribuir a LlamaManager. Aquí están los pasos y normas esenciales para que tu contribución sea rápida de revisar y aceptar.

1) Antes de empezar
- Lee la licencia (LICENSE) — este proyecto está bajo GNU GPL v3.
- Abre un issue si vas a trabajar en un cambio no trivial (funcionalidad nueva, refactor grande) para discutir la implementación.

2) Entorno de desarrollo
- Clona el repositorio y crea un entorno virtual:

  git clone https://github.com/colomer510-netizen/LlamaManager.git
  cd LlamaManager
  python -m venv .venv
  source .venv/bin/activate   # macOS / Linux
  .\.venv\Scripts\activate  # Windows (PowerShell)

- Instala dependencias:

  pip install -r requirements.txt

3) Flujo de trabajo (fork -> branch -> PR)
- Haz fork del repo y trabaja en una rama descriptiva:

  git checkout -b feat/nombre-descriptivo

- Mantén los commits pequeños y enfocados. Usa un mensaje de commit corto (máx ~50 chars) y en caso necesario un cuerpo explicativo.

4) Estilo de código y calidad
- Formateo: usa black

  pip install black
  black .

- Linting: usa flake8

  pip install flake8
  flake8 .

- Tests: añade pruebas para cambios funcionales y ejecuta pytest antes de enviar PR

  pip install pytest
  pytest

- Asegúrate de que los tests pasan y que el código está formateado y sin errores de lint.

5) Pull Request (PR)
- Abre el PR desde tu fork a la rama main del upstream.
- Título claro: tipo(short): descripción breve — por ejemplo: feat(api): add model-download endpoint
- En la descripción incluye:
  - Qué problema se resuelve
  - Cambios principales
  - Cómo probarlo localmente (comandos)
  - Si aplica, referencia a issues: "Fixes #NN"

- PR checklist mínima:
  - [ ] Los tests pasan (pytest)
  - [ ] Código formateado con black
  - [ ] No hay errores de lint (flake8)
  - [ ] Añadiste o actualizaste tests si corresponde

6) Revisión y merge
- Los mantenedores revisarán el PR y pueden pedir cambios. Responde a los comentarios y actualiza tu rama.
- El merge lo hará el equipo de mantenimiento cuando el PR cumpla los criterios y pase CI.

7) Código y dependencias
- Mantén las dependencias razonablemente actualizadas y evita añadir paquetes innecesarios.
- Asegúrate de que cualquier dependencia nueva tiene licencia compatible con GPLv3.

8) Contacto y comportamiento
- Sé respetuoso y claro en la comunicación. Si hay desacuerdo técnico, prioriza argumentos técnicos y soluciones concretas.
- Para preguntas rápidas, abre un issue y etiqueta al mantenedor (@colomer510-netizen).

Gracias por tu contribución — tu ayuda mejora el proyecto. Si quieres, puedo añadir una plantilla de PR e ISSUE en .github/ cuando lo solicites.
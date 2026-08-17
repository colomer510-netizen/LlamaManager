# Gestor Universal de Llama.cpp (BAT a EXE)

Este mini-repositorio contiene un script interactivo en `.bat` diseñado para usuarios de Windows. Permite descargar, gestionar y ejecutar modelos de LLaMA utilizando las herramientas nativas de `llama.cpp` de una forma sumamente sencilla e interactiva.

## 🚀 Características
- **Portabilidad total**: Diseñado con rutas dinámicas (`%~dp0`) para ser compilado en un archivo `.exe` y llevado en un USB a cualquier PC.
- **Validación inteligente**: Verifica automáticamente si `llama.cpp` está instalado en la computadora destino y ofrece instalarlo si no lo encuentra.
- **Detección automática**: Escanea la carpeta en busca de modelos `.gguf` y permite seleccionarlos con un número.
- **Soporte Drag & Drop**: Puedes arrastrar un modelo `.gguf` directamente sobre el archivo (o su `.exe`) para iniciar un chat instantáneo.
- **Herramientas integradas**:
  - `llama-cli` (Chat interactivo en terminal)
  - `llama-server` (Servidor Web)
  - `llama-quantize` (Cuantizador para reducir tamaño de memoria)
  - `llama-bench` (Pruebas de rendimiento de CPU/GPU)
  - `llama-imatrix` (Matrices de importancia para cuantización precisa)
  - `llama-gguf-split` (Dividir y unir modelos gigantes)

## 🛠️ Cómo compilar a .EXE

Para que este script funcione como un programa independiente y se distribuya fácilmente, se recomienda convertirlo a un archivo ejecutable (`.exe`). 

### Herramientas Recomendadas
1. **Advanced BAT to EXE Converter** (Recomendado)
   Es muy estable y permite añadir íconos personalizados fácilmente.
   - 📥 [Sitio web oficial para descargar](https://www.battoexeconverter.com/)

2. **Bat To Exe Converter** (de Fatih Kodak)
   Un clásico, muy personalizable y ampliamente utilizado por la comunidad.

3. **IExpress (Nativo de Windows)**
   Windows tiene un empaquetador secreto incluido en el sistema. Puedes probarlo presionando `Win + R`, escribiendo `iexpress` y siguiendo el asistente.

### Pasos Generales para la Compilación:
1. Abre tu convertidor de preferencia.
2. Selecciona el archivo `gestor_llama.bat`.
3. Selecciona la opción de **Modo Consola** (Importante: *no uses el modo "Invisible" o "Ghost", ya que los usuarios necesitan ver el menú en la consola*).
4. (Opcional) Asigna un ícono en formato `.ico`.
5. Presiona **Compilar**. El `.exe` generado funcionará en cualquier computadora Windows.

## 📜 Código Fuente
El código completo del gestor se encuentra en el archivo [`gestor_llama.bat`](gestor_llama.bat) en esta misma carpeta. Puedes editarlo con cualquier editor de texto plano (como Bloc de Notas o VS Code) para añadir nuevas funciones en el futuro.

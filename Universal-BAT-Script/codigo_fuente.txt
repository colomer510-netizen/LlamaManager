@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
title Gestor Universal de Llama.cpp
:: Configurar color a Cyan sobre Negro para un look mas moderno
color 0B

:: Asegurar que el directorio de trabajo es el mismo donde esta el .bat/.exe
cd /d "%~dp0"

:: Si arrastras un archivo .gguf directamente sobre el archivo .bat/.exe
if not "%~1"=="" (
    set "model=%~1"
    goto cli_directo
)

set "opcion_actual=1"
set "total_opciones=8"

:menu_loop
cls
echo =================================================================
echo                 GESTOR UNIVERSAL DE LLAMA.CPP
echo =================================================================
echo.
echo Usa las FLECHAS (Arriba/Abajo) para moverte y ENTER para elegir.
echo.

if "!opcion_actual!"=="1" ( echo ^> [X] 1. Instalar / Actualizar llama.cpp ^< ) else ( echo   [ ] 1. Instalar / Actualizar llama.cpp )
echo.
if "!opcion_actual!"=="2" ( echo ^> [X] 2. Iniciar Chat en Terminal -llama cli- ^< ) else ( echo   [ ] 2. Iniciar Chat en Terminal -llama cli- )
if "!opcion_actual!"=="3" ( echo ^> [X] 3. Iniciar Servidor Web -llama serve- ^< ) else ( echo   [ ] 3. Iniciar Servidor Web -llama serve- )
echo.
if "!opcion_actual!"=="4" ( echo ^> [X] 4. Benchmark de Rendimiento -llama bench- ^< ) else ( echo   [ ] 4. Benchmark de Rendimiento -llama bench- )
if "!opcion_actual!"=="5" ( echo ^> [X] 5. Cuantizar un modelo -llama quantize- ^< ) else ( echo   [ ] 5. Cuantizar un modelo -llama quantize- )
if "!opcion_actual!"=="6" ( echo ^> [X] 6. Generar Matriz de Importancia -llama-imatrix- ^< ) else ( echo   [ ] 6. Generar Matriz de Importancia -llama-imatrix- )
if "!opcion_actual!"=="7" ( echo ^> [X] 7. Dividir/Unir GGUF -llama-gguf-split- ^< ) else ( echo   [ ] 7. Dividir/Unir GGUF -llama-gguf-split- )
echo.
if "!opcion_actual!"=="8" ( echo ^> [X] 8. Salir ^< ) else ( echo   [ ] 8. Salir )
echo =================================================================

:: Capturar tecla con PowerShell sin pausar visiblemente y limpiando el buffer de teclas
for /f %%a in ('powershell -NoProfile -ExecutionPolicy Bypass -Command "$Host.UI.RawUI.FlushInputBuffer(); while(1){if([Console]::KeyAvailable){$k=[Console]::ReadKey($true).Key; if($k -eq 'UpArrow'){$k='Up';break} elseif($k -eq 'DownArrow'){$k='Down';break} elseif($k -eq 'Enter'){$k='Enter';break}}}; $k"') do set "key=%%a"

if "!key!"=="Up" (
    set /a opcion_actual-=1
    if !opcion_actual! lss 1 set opcion_actual=!total_opciones!
    goto menu_loop
)
if "!key!"=="Down" (
    set /a opcion_actual+=1
    if !opcion_actual! gtr !total_opciones! set opcion_actual=1
    goto menu_loop
)
if "!key!"=="Enter" (
    if "!opcion_actual!"=="1" goto instalar
    if "!opcion_actual!"=="2" goto cli
    if "!opcion_actual!"=="3" goto servidor
    if "!opcion_actual!"=="4" goto bench
    if "!opcion_actual!"=="5" goto cuantizar
    if "!opcion_actual!"=="6" goto imatrix
    if "!opcion_actual!"=="7" goto split
    if "!opcion_actual!"=="8" goto salir
)
goto menu_loop

:verificar_instalacion
:: Funcion para verificar si los binarios estan disponibles
where llama >nul 2>nul
if !errorlevel! neq 0 (
    echo [ADVERTENCIA] No se encontro 'llama' en el PATH.
    echo Si ya lo instalaste, ignora este mensaje. Intentando continuar...
    echo.
)
exit /b

:seleccionar_modelo
echo.
echo Buscando modelos .gguf en esta carpeta: %cd%
set count=0
for %%f in (*.gguf) do (
    set /a count+=1
    set "modelo[!count!]=%%f"
    echo !count!. %%f
)
echo.
if !count!==0 (
    echo No se encontraron archivos .gguf automaticamente.
    echo Por favor, arrastra tu modelo aqui o escribe la ruta completa:
    set /p model="Ruta al modelo: "
    set model=!model:"=!
    exit /b
)

set /a manual_opt=!count!+1
echo !manual_opt!. Escribir ruta o arrastrar archivo manualmente

set /p mod_opt="Elige un modelo (1-!manual_opt!): "
if "!mod_opt!"=="!manual_opt!" (
    set /p model="Arrastra el archivo o escribe la ruta: "
    set model=!model:"=!
) else (
    set "model=!modelo[%mod_opt%]!"
)
exit /b

:autodetectar_sistema
echo.
echo [INFO] Analizando caracteristicas del sistema...

:: 1. Detectar RAM aproximada
set "ram_gb=8"
for /f "tokens=2 delims==" %%A in ('wmic computersystem get TotalPhysicalMemory /value 2^>nul') do (
    set "bytes=%%A"
    :: Quitar los ultimos 9 digitos para aproximar gigabytes (ya que batch no soporta enteros > 32 bits)
    set "bytes_str=!bytes:~0,-9!"
    if not "!bytes_str!"=="" set /a ram_gb=!bytes_str!
)
if !ram_gb! geq 16 (
    set "ctx_flag=-c 4096"
    echo [OK] RAM detectada: !ram_gb! GB. Asignando contexto extendido (4096).
) else (
    set "ctx_flag=-c 2048"
    echo [OK] RAM detectada: !ram_gb! GB. Asignando contexto seguro (2048).
)

:: 2. Detectar Hilos CPU
set /a threads=%NUMBER_OF_PROCESSORS%
if !threads! gtr 2 (
    set /a opt_threads=!threads!-1
) else (
    set opt_threads=!threads!
)
set "thread_flag=-t !opt_threads!"
echo [OK] Procesador de !threads! hilos. Reservando 1 para el sistema, asignando !opt_threads! a IA.

:: 3. Prueba de Fuego de GPU
echo [INFO] Realizando prueba de estres grafico (GPU)...
llama cli -m "%model%" -n 1 -ngl 99 >nul 2>nul
if !errorlevel! neq 0 (
    echo [ADVERTENCIA] Fallo grafico detectado (ej. Drivers de Vulkan incompatibles).
    echo [INFO] Cambiando a Modo Seguro por CPU.
    set "gpu_flag=-ngl 0"
) else (
    echo [OK] Prueba grafica superada. Usando maxima aceleracion de GPU.
    set "gpu_flag=-ngl 99"
)
echo.
exit /b

:instalar
cls
echo =======================================================
echo Instalando/Actualizando llama.cpp...
echo =======================================================
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://llama.app/install.ps1 | iex"
echo.
echo Proceso finalizado.
pause
goto menu_loop

:cli
cls
echo =======================================================
echo Iniciar Chat en Terminal
echo =======================================================
call :verificar_instalacion
call :seleccionar_modelo

:cli_directo
if not exist "%model%" (
    echo El archivo no existe: "%model%"
    pause
    goto menu_loop
)

call :autodetectar_sistema
echo Iniciando chat interactivo optimizado...
llama cli -m "%model%" !ctx_flag! !thread_flag! !gpu_flag! -n -1 --color -i
pause
goto menu_loop

:servidor
cls
echo =======================================================
echo Iniciar Servidor Web Local
echo =======================================================
call :verificar_instalacion
call :seleccionar_modelo

if not exist "%model%" (
    echo El archivo no existe: "%model%"
    pause
    goto menu_loop
)

call :autodetectar_sistema
echo Iniciando servidor web en http://127.0.0.1:8080 ...
llama serve -m "%model%" --port 8080 !ctx_flag! !thread_flag! !gpu_flag!
pause
goto menu_loop

:bench
cls
echo =======================================================
echo Benchmark de Rendimiento
echo =======================================================
call :verificar_instalacion
call :seleccionar_modelo

if not exist "%model%" (
    echo El archivo no existe.
    pause
    goto menu_loop
)
call :autodetectar_sistema
echo Ejecutando prueba de velocidad optimizada...
llama bench -m "%model%" !thread_flag! !gpu_flag!
pause
goto menu_loop

:cuantizar
cls
echo =======================================================
echo Cuantizar un Modelo
echo =======================================================
call :verificar_instalacion
echo Arrastra el archivo de modelo de entrada (.gguf/.bin) aqui:
set /p input="Ruta de entrada: "
set input=!input:"=!

if not exist "!input!" (
    echo El archivo de entrada no existe: !input!
    pause
    goto menu_loop
)

echo.
echo Escribe el nombre de salida para el nuevo archivo cuantizado:
set /p output="Ruta de salida (ej: modelo_Q4.gguf): "
set output=!output:"=!

echo.
echo Metodos comunes: Q4_K_M (recomendado), Q5_K_M, Q8_0
set /p method="Metodo de cuantizacion (ej. Q4_K_M): "

echo.
echo Iniciando cuantizacion...
llama quantize "!input!" "!output!" !method!
pause
goto menu_loop

:imatrix
cls
echo =======================================================
echo Generar Matriz de Importancia
echo =======================================================
call :verificar_instalacion
call :seleccionar_modelo

if not exist "%model%" (
    echo El archivo no existe.
    pause
    goto menu_loop
)

echo.
echo Arrastra el archivo de datos de calibracion (.txt) aqui:
set /p data="Ruta del texto (ej. wiki.txt): "
set data=!data:"=!

echo.
echo Iniciando calculo de imatrix (esto puede tomar tiempo)...
llama-imatrix -m "%model%" -f "!data!" -o "imatrix.dat"
echo.
echo Archivo guardado como imatrix.dat
pause
goto menu_loop

:split
cls
echo =======================================================
echo Dividir / Unir GGUF
echo =======================================================
call :verificar_instalacion
echo 1. Dividir archivo grande a pequenos (split)
echo 2. Unir archivos pequenos a uno grande (merge)
set /p split_opt="Opcion (1-2): "

if "%split_opt%"=="1" (
    echo.
    echo Arrastra el archivo GGUF original:
    set /p input="Entrada: "
    set input=!input:"=!
    echo Escribe el prefijo de salida -ej. modelo_dividido-:
    set /p output="Salida: "
    set output=!output:"=!
    llama-gguf-split --split "!input!" "!output!"
) else if "%split_opt%"=="2" (
    echo.
    echo Arrastra el PRIMER archivo dividido -ej. archivo-00001-of-00005.gguf-:
    set /p input="Primera parte: "
    set input=!input:"=!
    echo Escribe el nombre final del archivo -ej. modelo_unido.gguf-:
    set /p output="Salida: "
    set output=!output:"=!
    llama-gguf-split --merge "!input!" "!output!"
) else (
    echo Opcion invalida.
)
pause
goto menu_loop

:salir
exit

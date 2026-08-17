@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
title Gestor Universal de Llama.cpp

:: Asegurar que el directorio de trabajo es el mismo donde esta el .bat/.exe
cd /d "%~dp0"

:: Si arrastras un archivo .gguf directamente sobre el archivo .bat/.exe
if not "%~1"=="" (
    set "model=%~1"
    goto cli_directo
)

:menu
cls
echo =================================================================
echo                 GESTOR UNIVERSAL DE LLAMA.CPP
echo =================================================================
echo.
echo --- INSTALACION ---
echo 1. Instalar / Actualizar llama.cpp (requiere internet)
echo.
echo --- USO PRINCIPAL ---
echo 2. Iniciar Chat en Terminal (llama-cli)
echo 3. Iniciar Servidor Web (llama-server)
echo.
echo --- HERRAMIENTAS AVANZADAS ---
echo 4. Benchmark de Rendimiento (llama-bench)
echo 5. Cuantizar un modelo (llama-quantize)
echo 6. Generar Matriz de Importancia (llama-imatrix)
echo 7. Dividir/Unir GGUF (llama-gguf-split)
echo.
echo 8. Salir
echo =================================================================
set /p opcion="Elige una opcion (1-8): "

if "%opcion%"=="1" goto instalar
if "%opcion%"=="2" goto cli
if "%opcion%"=="3" goto servidor
if "%opcion%"=="4" goto bench
if "%opcion%"=="5" goto cuantizar
if "%opcion%"=="6" goto imatrix
if "%opcion%"=="7" goto split
if "%opcion%"=="8" goto salir

echo Opcion no valida.
pause
goto menu

:verificar_instalacion
:: Funcion para verificar si los binarios estan disponibles
where llama-cli.exe >nul 2>nul
if !errorlevel! neq 0 (
    echo [ADVERTENCIA] No se encontro 'llama-cli' en el sistema.
    echo Probablemente necesites ejecutar la opcion 1 (Instalar) primero.
    pause
    goto menu
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

:instalar
cls
echo =======================================================
echo Instalando/Actualizando llama.cpp...
echo =======================================================
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://llama.app/install.ps1 | iex"
echo.
echo Proceso finalizado.
pause
goto menu

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
    goto menu
)

echo Iniciando chat interactivo...
llama-cli.exe -m "%model%" -c 2048 -n -1 --color -i
pause
goto menu

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
    goto menu
)

echo Iniciando servidor web en http://127.0.0.1:8080 ...
llama-server.exe -m "%model%" --port 8080 -c 2048
pause
goto menu

:bench
cls
echo =======================================================
echo Benchmark de Rendimiento (llama-bench)
echo =======================================================
call :verificar_instalacion
call :seleccionar_modelo

if not exist "%model%" (
    echo El archivo no existe.
    pause
    goto menu
)
echo Ejecutando prueba de velocidad...
llama-bench.exe -m "%model%"
pause
goto menu

:cuantizar
cls
echo =======================================================
echo Cuantizar un Modelo (Reducir tamano/memoria)
echo =======================================================
call :verificar_instalacion
echo Arrastra el archivo de modelo de entrada (.gguf/.bin) aqui:
set /p input="Ruta de entrada: "
set input=!input:"=!

if not exist "!input!" (
    echo El archivo de entrada no existe: !input!
    pause
    goto menu
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
llama-quantize.exe "!input!" "!output!" !method!
pause
goto menu

:imatrix
cls
echo =======================================================
echo Generar Matriz de Importancia (llama-imatrix)
echo =======================================================
call :verificar_instalacion
call :seleccionar_modelo

if not exist "%model%" (
    echo El archivo no existe.
    pause
    goto menu
)

echo.
echo Arrastra el archivo de datos de calibracion (.txt) aqui:
set /p data="Ruta del texto (ej. wiki.txt): "
set data=!data:"=!

echo.
echo Iniciando calculo de imatrix (esto puede tomar tiempo)...
llama-imatrix.exe -m "%model%" -f "!data!" -o "imatrix.dat"
echo.
echo Archivo guardado como imatrix.dat
pause
goto menu

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
    echo Escribe el prefijo de salida (ej. modelo_dividido):
    set /p output="Salida: "
    set output=!output:"=!
    llama-gguf-split.exe --split "!input!" "!output!"
) else if "%split_opt%"=="2" (
    echo.
    echo Arrastra el PRIMER archivo dividido (ej. archivo-00001-of-00005.gguf):
    set /p input="Primera parte: "
    set input=!input:"=!
    echo Escribe el nombre final del archivo (ej. modelo_unido.gguf):
    set /p output="Salida: "
    set output=!output:"=!
    llama-gguf-split.exe --merge "!input!" "!output!"
) else (
    echo Opcion invalida.
)
pause
goto menu

:salir
exit

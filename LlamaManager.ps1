<#
.SYNOPSIS
    LlamaManager - Administrador interactivo para herramientas de llama.cpp
.DESCRIPTION
    Script de PowerShell avanzado que proporciona una interfaz de linea de comandos
    interactiva con navegacion por flechas, wizards guiados para cada herramienta,
    explorador de modelos .gguf, sistema de perfiles/presets, y ejecucion directa.
.NOTES
    Autor: Generado por Antigravity AI
    Fecha: 2026-06-03
    Requiere: PowerShell 5.1+ | llama.cpp binarios en subcarpeta bin/
#>

# ================================================================
#                    CONFIGURACION GLOBAL
# ================================================================

# Ruta base: la carpeta donde vive este script (junto a bin/)
$Script:BasePath      = $PSScriptRoot
$Script:BinPath       = Join-Path $Script:BasePath "bin"
$Script:ProfilePath   = Join-Path $Script:BasePath "profiles"
$Script:Version       = "3.0.0"

$Script:ModelsPath    = Join-Path $Script:BasePath "models"
$Script:HistoryFile   = Join-Path $Script:ProfilePath "history.json"
$Script:ServerPidFile = Join-Path $Script:ProfilePath "server.pid"

# Detectar GPU
$Script:HasDedicatedGpu = $false
try {
    $gpu = Get-CimInstance Win32_VideoController -ErrorAction SilentlyContinue | Where-Object { $_.Name -match "NVIDIA|AMD|Radeon" }
    if ($gpu) { $Script:HasDedicatedGpu = $true }
} catch {}

# Biblioteca de System Prompts
$Script:PromptLibrary = @(
    @{ Name = "Asistente Util"; Desc = "Responde de forma concisa y servicial"; Prompt = "Eres un asistente util, inteligente y conciso. Responde siempre en el idioma en que te hablen." }
    @{ Name = "Programador Experto"; Desc = "Optimizado para codigo y desarrollo"; Prompt = "Eres un ingeniero de software senior. Al escribir codigo, proporciona solo la solucion optima con breves comentarios. Piensa paso a paso." }
    @{ Name = "Traductor Universal"; Desc = "Solo traduce texto"; Prompt = "Eres un traductor profesional. Tu unica tarea es traducir el texto que recibes al idioma solicitado, sin anadir conversacion adicional." }
    @{ Name = "Redactor / Escritor"; Desc = "Creativo y expansivo"; Prompt = "Eres un escritor creativo experto. Utiliza un lenguaje rico y elabora historias o textos bien estructurados." }
    @{ Name = "Analista de Datos"; Desc = "Logico y estructurado"; Prompt = "Eres un analista de datos experto. Piensa de forma logica, extrae conclusiones basadas en evidencia y formatea tu respuesta en tablas o listas estructuradas." }
)


# Rutas de busqueda predeterminadas para modelos .gguf
$Script:ModelSearchPaths = @(
    $Script:BasePath
    (Join-Path $Script:BasePath "models")
    (Join-Path $env:USERPROFILE "Downloads")
    (Join-Path $env:USERPROFILE "models")
    (Join-Path $env:USERPROFILE "Documents\models")
)

# Cantidad de hilos por defecto = procesadores logicos
$Script:DefaultThreads = [Environment]::ProcessorCount

# Tabla de categorizacion de ejecutables
$Script:ToolCategories = @{
    "llama-cli.exe"              = @{ Category = "Generacion"; Desc = "Chat / generacion de texto interactiva" }
    "llama.exe"                  = @{ Category = "Generacion"; Desc = "Lanzador principal de llama.cpp" }
    "llama-server.exe"           = @{ Category = "Servidor";   Desc = "Servidor API compatible con OpenAI" }
    "llama-quantize.exe"         = @{ Category = "Cuantizacion"; Desc = "Cuantizar modelos GGUF" }
    "llama-bench.exe"            = @{ Category = "Benchmark";  Desc = "Benchmark de rendimiento" }
    "llama-batched-bench.exe"    = @{ Category = "Benchmark";  Desc = "Benchmark por lotes" }
    "llama-perplexity.exe"       = @{ Category = "Evaluacion"; Desc = "Evaluacion de perplejidad" }
    "llama-completion.exe"       = @{ Category = "Generacion"; Desc = "Completacion de texto (single-shot)" }
    "llama-tokenize.exe"         = @{ Category = "Utilidad";   Desc = "Inspeccion de tokenizador" }
    "llama-imatrix.exe"          = @{ Category = "Cuantizacion"; Desc = "Generar matriz de importancia" }
    "llama-gguf-split.exe"       = @{ Category = "Utilidad";   Desc = "Dividir/unir shards GGUF" }
    "llama-tts.exe"              = @{ Category = "Audio";      Desc = "Texto a voz (TTS)" }
    "llama-mtmd-cli.exe"         = @{ Category = "Multimodal"; Desc = "CLI multimodal (imagen+texto)" }
    "llama-gemma3-cli.exe"       = @{ Category = "Multimodal"; Desc = "CLI para Gemma 3 (multimodal)" }
    "llama-llava-cli.exe"        = @{ Category = "Multimodal"; Desc = "CLI para LLaVA (multimodal)" }
    "llama-minicpmv-cli.exe"     = @{ Category = "Multimodal"; Desc = "CLI para MiniCPM-V (multimodal)" }
    "llama-qwen2vl-cli.exe"      = @{ Category = "Multimodal"; Desc = "CLI para Qwen2-VL (multimodal)" }
    "llama-mtmd-debug.exe"       = @{ Category = "Debug";      Desc = "Debug multimodal" }
    "llama-fit-params.exe"       = @{ Category = "Utilidad";   Desc = "Ajuste automatico de parametros" }
    "llama-template-analysis.exe"= @{ Category = "Utilidad";   Desc = "Analisis de templates de chat" }
    "llama-results.exe"          = @{ Category = "Utilidad";   Desc = "Visor de resultados" }
    "rpc-server.exe"             = @{ Category = "Servidor";   Desc = "Servidor RPC backend" }
}

# Tipos de cuantizacion disponibles con descripcion
$Script:QuantTypes = @(
    @{ Name = "Q4_K_M";   Desc = "Recomendado - buen balance calidad/tamano - 4.5 bpw" }
    @{ Name = "Q4_K_S";   Desc = "Pequeno - calidad aceptable, menor tamano - 4.3 bpw" }
    @{ Name = "Q5_K_M";   Desc = "Alta calidad - mayor tamano - 5.5 bpw" }
    @{ Name = "Q5_K_S";   Desc = "Alta calidad - ligeramente menor - 5.3 bpw" }
    @{ Name = "Q6_K";     Desc = "Muy alta calidad - 6.5 bpw" }
    @{ Name = "Q8_0";     Desc = "Casi sin perdida, grande - 8.0 bpw" }
    @{ Name = "Q3_K_M";   Desc = "Agresivo - modelos grandes en poca RAM - 3.9 bpw" }
    @{ Name = "Q3_K_S";   Desc = "Muy agresivo - 3.5 bpw" }
    @{ Name = "Q3_K_L";   Desc = "Agresivo mejorado - 3.9 bpw" }
    @{ Name = "Q2_K";     Desc = "Extremo - solo para pruebas - 2.6 bpw" }
    @{ Name = "Q4_0";     Desc = "Legacy - compatible con todo - 4.0 bpw" }
    @{ Name = "Q4_1";     Desc = "Legacy mejorado - 4.5 bpw" }
    @{ Name = "Q5_0";     Desc = "Legacy alta calidad - 5.0 bpw" }
    @{ Name = "Q5_1";     Desc = "Legacy alta calidad+ - 5.5 bpw" }
    @{ Name = "IQ2_XXS";  Desc = "iQuant extremo, muy pequeno - 2.1 bpw" }
    @{ Name = "IQ2_XS";   Desc = "iQuant, pequeno - 2.3 bpw" }
    @{ Name = "IQ2_S";    Desc = "iQuant, estandar - 2.5 bpw" }
    @{ Name = "IQ2_M";    Desc = "iQuant, medio - 2.7 bpw" }
    @{ Name = "IQ3_XXS";  Desc = "iQuant 3bit, muy pequeno - 3.1 bpw" }
    @{ Name = "IQ3_XS";   Desc = "iQuant 3bit, pequeno - 3.3 bpw" }
    @{ Name = "IQ3_S";    Desc = "iQuant 3bit, estandar - 3.4 bpw" }
    @{ Name = "IQ3_M";    Desc = "iQuant 3bit, medio - 3.6 bpw" }
    @{ Name = "IQ4_NL";   Desc = "iQuant 4bit, non-linear - 4.3 bpw" }
    @{ Name = "IQ4_XS";   Desc = "iQuant 4bit, extra small - 4.1 bpw" }
    @{ Name = "F16";      Desc = "Float16 - sin perdida, tamano completo" }
    @{ Name = "BF16";     Desc = "BFloat16 - sin perdida, optimizado" }
    @{ Name = "F32";      Desc = "Float32 - maxima precision, muy grande" }
    @{ Name = "COPY";     Desc = "Copiar sin recuantizar" }
)


# ================================================================
#                     FUNCIONES DE UI
# ================================================================

function Write-Color {
    <#
    .SYNOPSIS
        Escribe texto con color al host.
    #>
    param(
        [string]$Text,
        [ConsoleColor]$Color = "White",
        [switch]$NoNewline
    )
    Write-Host $Text -ForegroundColor $Color -NoNewline:$NoNewline
}

function Write-ColorLine {
    <#
    .SYNOPSIS
        Escribe multiples segmentos coloreados en una sola linea.
    #>
    param([array]$Segments)
    foreach ($seg in $Segments) {
        Write-Host $seg.Text -ForegroundColor $seg.Color -NoNewline
    }
    Write-Host ""
}

function Show-Banner {
    <#
    .SYNOPSIS
        Muestra el banner ASCII estilizado con informacion de version.
    #>
    Clear-Host
    $v = $Script:Version
    Write-Host ""
    Write-Host "    +===============================================================+" -ForegroundColor Cyan
    Write-Host "    |                                                               |" -ForegroundColor Cyan
    Write-Host "    |      ##      ##          ###    ##   ## ###                    |" -ForegroundColor Cyan
    Write-Host "    |      ##      ##         ## ##   ### ### ## ##                  |" -ForegroundColor Cyan
    Write-Host "    |      ##      ##        ##   ##  ## # ## #####                 |" -ForegroundColor Cyan
    Write-Host "    |      ##      ##        #######  ##   ## ## ##                 |" -ForegroundColor Cyan
    Write-Host "    |      ######  ######    ##   ##  ##   ## ## ##                 |" -ForegroundColor Cyan
    Write-Host "    |                                                               |" -ForegroundColor Cyan
    Write-Host "    |           M A N A G E R   v${v}                          |" -ForegroundColor Cyan
    Write-Host "    |                                                               |" -ForegroundColor Cyan
    Write-Host "    +===============================================================+" -ForegroundColor Cyan
    Write-Host ""
    Write-ColorLine @(
        @{ Text = "    Ruta bin: "; Color = "DarkGray" },
        @{ Text = $Script:BinPath; Color = "Yellow" }
    )
    $psVer = $PSVersionTable.PSVersion.ToString()
    Write-ColorLine @(
        @{ Text = "    Hilos CPU: "; Color = "DarkGray" },
        @{ Text = "$Script:DefaultThreads"; Color = "Green" },
        @{ Text = " | PowerShell: "; Color = "DarkGray" },
        @{ Text = $psVer; Color = "Green" }
    )
    Write-Host ""
}

function Show-Separator {
    <#
    .SYNOPSIS
        Dibuja una linea separadora visual.
    #>
    param([string]$Title = "", [ConsoleColor]$Color = "DarkCyan")
    if ($Title) {
        $line = "---"
        $padLen = 56 - $Title.Length
        if ($padLen -lt 1) { $padLen = 1 }
        $pad = "-" * $padLen
        Write-Host "  $line $Title $pad" -ForegroundColor $Color
    } else {
        Write-Host ("  " + ("-" * 62)) -ForegroundColor $Color
    }
}

function Show-Menu {
    <#
    .SYNOPSIS
        Menu interactivo navegable con flechas del teclado.
        Devuelve el indice seleccionado o -1 si se presiona Escape.
    #>
    param(
        [string]$Title,
        [string[]]$Options,
        [string[]]$Descriptions = @(),
        [ConsoleColor[]]$Colors = @(),
        [int]$DefaultIndex = 0,
        [switch]$ShowBack
    )

    if ($Options.Count -eq 0) {
        Write-Color "  [!] No hay opciones disponibles." "Yellow"
        return -1
    }

    $selectedIndex = [Math]::Min($DefaultIndex, $Options.Count - 1)
    $startLine     = [Console]::CursorTop

    # total de lineas del menu, se calcula en la primera pasada
    $totalMenuLines = 0

    while ($true) {
        # Mover cursor al inicio del menu para re-dibujar en el mismo lugar
        if ($totalMenuLines -gt 0) {
            [Console]::SetCursorPosition(0, $startLine)
        }

        $linesDrawn = 0

        # Titulo
        Write-Host ""
        Write-Color "  $Title" "Cyan"
        Show-Separator
        $linesDrawn += 3

        # Instrucciones
        $escText = ""
        if ($ShowBack) { $escText = "  [Esc] Volver" }
        Write-ColorLine @(
            @{ Text = "  ["; Color = "DarkGray" },
            @{ Text = "Up/Down"; Color = "Yellow" },
            @{ Text = "] Navegar  ["; Color = "DarkGray" },
            @{ Text = "Enter"; Color = "Green" },
            @{ Text = "] Seleccionar"; Color = "DarkGray" },
            @{ Text = $escText; Color = "Red" }
        )
        Write-Host ""
        $linesDrawn += 2

        # Opciones
        for ($i = 0; $i -lt $Options.Count; $i++) {
            $numLabel  = "[{0}]" -f ($i + 1)

            if ($i -eq $selectedIndex) {
                $optColor = "White"
                if ($Colors.Count -gt $i -and $null -ne $Colors[$i]) { $optColor = $Colors[$i] }
                # Opcion seleccionada: fondo resaltado
                Write-Host " >> " -NoNewline -ForegroundColor "Green"
                Write-Host "$numLabel " -NoNewline -ForegroundColor "DarkYellow"
                Write-Host $Options[$i] -NoNewline -ForegroundColor "Black" -BackgroundColor "Green"

                if ($Descriptions.Count -gt $i -and $Descriptions[$i]) {
                    $desc = $Descriptions[$i]
                    Write-Host " -- $desc" -ForegroundColor "DarkGreen"
                } else {
                    Write-Host ""
                }
            } else {
                $optColor = "White"
                if ($Colors.Count -gt $i -and $null -ne $Colors[$i]) { $optColor = $Colors[$i] }
                Write-Host "    " -NoNewline -ForegroundColor "DarkGray"
                Write-Host "$numLabel " -NoNewline -ForegroundColor "DarkGray"
                Write-Host $Options[$i] -NoNewline -ForegroundColor $optColor

                if ($Descriptions.Count -gt $i -and $Descriptions[$i]) {
                    $desc = $Descriptions[$i]
                    Write-Host " -- $desc" -ForegroundColor "DarkGray"
                } else {
                    Write-Host ""
                }
            }
            $linesDrawn++
        }

        Write-Host ""
        $linesDrawn++

        # Guardar total de lineas para el proximo redibujado
        if ($totalMenuLines -eq 0) {
            $totalMenuLines = $linesDrawn
            $startLine      = [Console]::CursorTop - $totalMenuLines
        }

        # Leer tecla (sin eco)
        $key = [Console]::ReadKey($true)

        switch ($key.Key) {
            "UpArrow" {
                if ($selectedIndex -le 0) { $selectedIndex = $Options.Count - 1 } else { $selectedIndex-- }
            }
            "DownArrow" {
                if ($selectedIndex -ge ($Options.Count - 1)) { $selectedIndex = 0 } else { $selectedIndex++ }
            }
            "Enter" {
                return $selectedIndex
            }
            "Escape" {
                if ($ShowBack) { return -1 }
            }
            default {
                # Seleccion numerica directa (1-9)
                $charStr = $key.KeyChar.ToString()
                $numVal = -1
                $parsed = 0
                if ([int]::TryParse($charStr, [ref]$parsed) -and $parsed -ge 1 -and $parsed -le 9) {
                    $numVal = $parsed - 1
                }
                if ($numVal -ge 0 -and $numVal -lt $Options.Count) {
                    return $numVal
                }
            }
        }
    }
}

function Read-ValidInput {
    <#
    .SYNOPSIS
        Lee input del usuario con validacion, valor por defecto, y mensajes descriptivos.
    #>
    param(
        [string]$Prompt,
        [string]$Default = "",
        [ValidateSet("string","int","float","path","port")]
        [string]$ValidationType = "string",
        [switch]$Required,
        [int]$Min = [int]::MinValue,
        [int]$Max = [int]::MaxValue
    )

    while ($true) {
        $displayDefault = ""
        if ($Default) { $displayDefault = " [$Default]" }
        Write-Host ""
        Write-Host "  $Prompt" -NoNewline -ForegroundColor "Yellow"
        Write-Host "$displayDefault" -NoNewline -ForegroundColor "DarkGray"
        Write-Host ": " -NoNewline -ForegroundColor "Yellow"

        $userInput = Read-Host
        $value = $Default
        if (-not [string]::IsNullOrWhiteSpace($userInput)) { $value = $userInput.Trim() }

        # Vacio y requerido
        if ([string]::IsNullOrWhiteSpace($value)) {
            if ($Required -and -not $Default) {
                Write-Color "  [X] Este campo es obligatorio." "Red"
                continue
            }
            return $value
        }

        # Validacion por tipo
        switch ($ValidationType) {
            "int" {
                $parsed = 0
                if (-not [int]::TryParse($value, [ref]$parsed)) {
                    Write-Color "  [X] Ingresa un numero entero valido." "Red"
                    continue
                }
                if ($parsed -lt $Min -or $parsed -gt $Max) {
                    Write-Color "  [X] El valor debe estar entre $Min y $Max." "Red"
                    continue
                }
                return $parsed
            }
            "float" {
                $parsed = 0.0
                if (-not [double]::TryParse($value, [ref]$parsed)) {
                    Write-Color "  [X] Ingresa un numero decimal valido (ej: 0.7)." "Red"
                    continue
                }
                return $parsed
            }
            "port" {
                $parsed = 0
                if (-not [int]::TryParse($value, [ref]$parsed) -or $parsed -lt 1 -or $parsed -gt 65535) {
                    Write-Color "  [X] Ingresa un puerto valido (1-65535)." "Red"
                    continue
                }
                return $parsed
            }
            "path" {
                if (-not (Test-Path $value)) {
                    Write-Color "  [X] La ruta no existe: $value" "Red"
                    Write-Color "  Deseas intentar de nuevo? (S/n)" "Yellow"
                    $retry = Read-Host
                    if ($retry -eq 'n' -or $retry -eq 'N') { return $null }
                    continue
                }
                return $value
            }
            default {
                return $value
            }
        }
    }
}

function Show-Confirm {
    <#
    .SYNOPSIS
        Muestra una confirmacion Si/No y devuelve $true o $false.
    #>
    param(
        [string]$Message,
        [bool]$Default = $true
    )
    $hint = "(s/N)"
    if ($Default) { $hint = "(S/n)" }
    Write-Host ""
    Write-Host "  $Message " -NoNewline -ForegroundColor "Yellow"
    Write-Host "$hint " -NoNewline -ForegroundColor "DarkGray"
    $response = Read-Host

    if ([string]::IsNullOrWhiteSpace($response)) { return $Default }
    return ($response -match '^[SsYy]')
}


# ================================================================
#                   FUNCIONES DE DESCUBRIMIENTO
# ================================================================

function Find-LlamaExecutables {
    <#
    .SYNOPSIS
        Escanea la carpeta bin/ y devuelve una lista de ejecutables con metadatos.
    #>

    if (-not (Test-Path $Script:BinPath)) {
        Write-Color "  [X] ERROR: No se encontro la carpeta bin en: $Script:BinPath" "Red"
        Write-Color "  Asegurate de que el script esta en la raiz de llama.cpp" "DarkGray"
        return @()
    }

    $executables = Get-ChildItem -Path $Script:BinPath -Filter "*.exe" -File |
        Sort-Object Name |
        ForEach-Object {
            $info = $Script:ToolCategories[$_.Name]
            $cat  = "Otro"
            $desc = "Herramienta de llama.cpp"
            if ($info) { $cat = $info.Category; $desc = $info.Desc }
            [PSCustomObject]@{
                Name        = $_.Name
                BaseName    = $_.BaseName
                Path        = $_.FullName
                SizeMB      = [Math]::Round($_.Length / 1MB, 2)
                Category    = $cat
                Description = $desc
            }
        }

    return $executables
}

function Find-GgufModels {
    <#
    .SYNOPSIS
        Busca archivos .gguf en las rutas de busqueda configuradas.
    #>
    param(
        [string[]]$SearchPaths = $Script:ModelSearchPaths,
        [switch]$Recursive
    )

    $models = @()
    $searchedPaths = @()

    foreach ($searchPath in $SearchPaths) {
        if (-not (Test-Path $searchPath)) { continue }
        $searchedPaths += $searchPath

        $depth = 2
        if ($Recursive) { $depth = 5 }

        try {
            $found = Get-ChildItem -Path $searchPath -Filter "*.gguf" -File -Recurse -Depth $depth -ErrorAction SilentlyContinue
            foreach ($file in $found) {
                # Evitar duplicados por ruta completa
                if ($models.Path -notcontains $file.FullName) {
                    $models += [PSCustomObject]@{
                        Name    = $file.Name
                        Path    = $file.FullName
                        Dir     = $file.DirectoryName
                        SizeGB  = [Math]::Round($file.Length / 1GB, 2)
                        SizeMB  = [Math]::Round($file.Length / 1MB, 0)
                    }
                }
            }
        } catch {
            # Silenciar errores de acceso a directorios protegidos
        }
    }

    if ($models.Count -eq 0 -and $searchedPaths.Count -gt 0) {
        Write-Host ""
        Write-Color "  [!] No se encontraron archivos .gguf en las rutas configuradas:" "Yellow"
        foreach ($p in $searchedPaths) {
            Write-Color "    - $p" "DarkGray"
        }
    }

    return $models | Sort-Object Name
}

function Browse-ForModel {
    <#
    .SYNOPSIS
        Explorador interactivo de modelos .gguf con opcion de busqueda manual.
    #>

    Write-Host ""
    Write-Color "  [~] Buscando modelos .gguf..." "Cyan"

    $models = Find-GgufModels

    if ($models.Count -eq 0) {
        Write-Host ""
        Write-Color "  No se encontraron modelos automaticamente." "Yellow"
        Write-Host ""

        $choice = Show-Menu -Title "Que deseas hacer?" -Options @(
            "Ingresar ruta manualmente"
            "Buscar en otra carpeta"
            "Buscar recursivamente (puede tardar)"
            "Cancelar"
        ) -ShowBack

        switch ($choice) {
            0 {
                $path = Read-ValidInput -Prompt "Ruta al archivo .gguf" -ValidationType "path" -Required
                if ($path -and $path.EndsWith(".gguf")) { return $path }
                Write-Color "  [X] El archivo debe ser .gguf" "Red"
                return $null
            }
            1 {
                $dir = Read-ValidInput -Prompt "Ruta de la carpeta a buscar" -ValidationType "path" -Required
                if ($dir) {
                    $models = Find-GgufModels -SearchPaths @($dir)
                    if ($models.Count -eq 0) {
                        Write-Color "  [X] No se encontraron modelos en esa carpeta." "Red"
                        return $null
                    }
                    # Continuar al selector de abajo
                } else { return $null }
            }
            2 {
                $models = Find-GgufModels -Recursive
                if ($models.Count -eq 0) {
                    Write-Color "  [X] No se encontraron modelos en ninguna ubicacion." "Red"
                    return $null
                }
            }
            default { return $null }
        }
    }

    # Mostrar lista de modelos encontrados
    $optionNames = @()
    $optionDescs = @()
    foreach ($m in $models) {
        $sizeLabel = "$($m.SizeMB) MB"
        if ($m.SizeGB -ge 1) { $sizeLabel = "$($m.SizeGB) GB" }
        $optionNames += $m.Name
        $optionDescs += "$sizeLabel -- $($m.Dir)"
    }
    $optionNames += "Ingresar ruta manualmente"
    $optionDescs += "Escribir la ruta completa al archivo"

    $sel = Show-Menu -Title "Selecciona un modelo .gguf" -Options $optionNames -Descriptions $optionDescs -ShowBack

    if ($sel -eq -1) { return $null }
    if ($sel -eq $models.Count) {
        # Ruta manual
        $path = Read-ValidInput -Prompt "Ruta completa al .gguf" -ValidationType "path" -Required
        return $path
    }

    $chosen = $models[$sel]
    $chosenSize = "$($chosen.SizeMB) MB"
    if ($chosen.SizeGB -ge 1) { $chosenSize = "$($chosen.SizeGB) GB" }
    Write-ColorLine @(
        @{ Text = "  [OK] Modelo: "; Color = "Green" },
        @{ Text = $chosen.Name; Color = "White" },
        @{ Text = " ($chosenSize)"; Color = "DarkGray" }
    )
    return $chosen.Path
}



# ================================================================
#                   NUEVAS FUNCIONES V2
# ================================================================

function Prompt-SystemPromptLibrary {
    $choice = Show-Menu -Title "Seleccionar System Prompt" -Options @("Escribir Manualmente", "Omitir (Sin System Prompt)", "------------------") + ($Script:PromptLibrary | ForEach-Object { $_.Name }) -Descriptions @("Escribir tu propio texto", "No anadir argumento", "") + ($Script:PromptLibrary | ForEach-Object { $_.Desc })
    
    if ($choice -eq 0) { return Read-ValidInput -Prompt "Escribe tu System Prompt" }
    if ($choice -eq 1 -or $choice -eq 2) { return $null }
    
    $idx = $choice - 3
    $selected = $Script:PromptLibrary[$idx].Prompt
    Write-Color "  [OK] Prompt cargado: $($Script:PromptLibrary[$idx].Name)" "Green"
    return $selected
}

function Add-History {
    param([hashtable]$CommandInfo)
    Initialize-ProfileDir
    
    $history = @()
    if (Test-Path $Script:HistoryFile) {
        try { $history = Get-Content $Script:HistoryFile -Raw | ConvertFrom-Json } catch {}
    }
    
    $newItem = @{
        Timestamp = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
        ToolName  = $CommandInfo.ToolName
        Executable = $CommandInfo.Executable
        Arguments = $CommandInfo.Arguments
    }
    
    $newHistory = @($newItem) + $history
    if ($newHistory.Count -gt 50) { $newHistory = $newHistory[0..49] }
    
    try { $newHistory | ConvertTo-Json -Depth 5 | Set-Content -Path $Script:HistoryFile -Encoding UTF8 } catch { Write-Color "  [!] No se pudo guardar el historial." "Yellow" }
}

function Show-History {
    if (-not (Test-Path $Script:HistoryFile)) {
        Write-Color "  [!] El historial esta vacio." "Yellow"
        Start-Sleep -Seconds 1
        return
    }
    
    try { $history = Get-Content $Script:HistoryFile -Raw | ConvertFrom-Json } catch { return }
    if (-not $history -or $history.Count -eq 0) { return }
    
    $opts = @()
    $descs = @()
    foreach ($h in $history) {
        $opts += "[$($h.Timestamp)] $($h.ToolName)"
        $cmd = $h.Arguments -join " "
        if ($cmd.Length -gt 60) { $cmd = $cmd.Substring(0, 57) + "..." }
        $descs += $cmd
    }
    
    $sel = Show-Menu -Title "Historial de Ejecuciones" -Options $opts -Descriptions $descs -ShowBack
    if ($sel -ge 0) {
        $chosen = $history[$sel]
        $cmdInfo = @{ Executable = $chosen.Executable; Arguments = @($chosen.Arguments); ToolName = $chosen.ToolName; Config=@{} }
        Show-PostWizardMenu -CommandInfo $cmdInfo
    }
}

function Invoke-QuickChat {
    if (-not (Test-Path $Script:HistoryFile)) { Write-Color "  [!] No hay historial previo para Chat Rapido." "Yellow"; Start-Sleep -Seconds 2; return }
    try { $history = Get-Content $Script:HistoryFile -Raw | ConvertFrom-Json } catch { return }
    if (-not $history) { return }
    
    $lastCli = $history | Where-Object { $_.ToolName -eq "llama-cli" } | Select-Object -First 1
    if (-not $lastCli) { Write-Color "  [!] No se encontro un modelo de chat en el historial." "Yellow"; Start-Sleep -Seconds 2; return }
    
    Invoke-LlamaCommand -ExePath $lastCli.Executable -Arguments $lastCli.Arguments -ToolName "llama-cli"
}


function Start-WebUI {
    Write-Host ""
    Show-Separator "Web UI Local" "Cyan"
    
    $htmlContent = @"
<!DOCTYPE html>
<html>
<head>
    <title>LlamaManager Web Chat</title>
    <meta charset="utf-8">
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #343541; color: #ececf1; margin: 0; display: flex; height: 100vh; flex-direction: column; }
        #header { background: #202123; padding: 15px; text-align: center; border-bottom: 1px solid #4d4d4f; font-weight: bold; }
        #chat { flex: 1; overflow-y: auto; padding: 20px; display: flex; flex-direction: column; gap: 15px; }
        .msg { max-width: 80%; padding: 15px; border-radius: 8px; line-height: 1.5; }
        .msg.user { background: #343541; align-self: flex-end; border: 1px solid #565869; }
        .msg.bot { background: #444654; align-self: flex-start; }
        #input-area { background: #343541; padding: 20px; border-top: 1px solid #565869; display: flex; justify-content: center; }
        #input-box { width: 60%; max-width: 800px; display: flex; gap: 10px; }
        input[type="text"] { flex: 1; padding: 12px; border-radius: 6px; border: 1px solid #565869; background: #40414f; color: white; outline: none; }
        button { padding: 12px 20px; background: #19c37d; color: white; border: none; border-radius: 6px; cursor: pointer; font-weight: bold; }
        button:hover { background: #1a8859; }
    </style>
</head>
<body>
    <div id="header">Llama.cpp Local Server (127.0.0.1:8080)</div>
    <div id="chat"></div>
    <div id="input-area">
        <div id="input-box">
            <input type="text" id="prompt" placeholder="Escribe un mensaje..." onkeypress="if(event.key==='Enter') send()">
            <button onclick="send()">Enviar</button>
        </div>
    </div>
    <script>
        const chat = document.getElementById('chat');
        const promptInput = document.getElementById('prompt');
        let history = [];

        function appendMsg(text, sender) {
            const div = document.createElement('div');
            div.className = 'msg ' + sender;
            div.innerText = text;
            chat.appendChild(div);
            chat.scrollTop = chat.scrollHeight;
        }

        async function send() {
            const text = promptInput.value.trim();
            if (!text) return;
            appendMsg(text, 'user');
            promptInput.value = '';
            history.push({ role: "user", content: text });
            
            const typingDiv = document.createElement('div');
            typingDiv.className = 'msg bot';
            typingDiv.innerText = 'Escribiendo...';
            chat.appendChild(typingDiv);
            chat.scrollTop = chat.scrollHeight;

            try {
                const response = await fetch('http://127.0.0.1:8080/v1/chat/completions', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ messages: history, stream: false })
                });
                
                chat.removeChild(typingDiv);
                
                if (response.ok) {
                    const data = await response.json();
                    const reply = data.choices[0].message.content;
                    appendMsg(reply, 'bot');
                    history.push({ role: "assistant", content: reply });
                } else {
                    appendMsg("Error al conectar con el servidor. Verifica que llama-server este ejecutandose.", 'bot');
                }
            } catch (err) {
                chat.removeChild(typingDiv);
                appendMsg("Error de conexion: " + err, 'bot');
            }
        }
    </script>
</body>
</html>
"@
    $tempHtml = Join-Path $env:TEMP "llama_webui.html"
    $htmlContent | Set-Content -Path $tempHtml -Encoding UTF8
    
    Write-Color "  [OK] Abriendo interfaz web en el navegador predeterminado..." "Green"
    Start-Process $tempHtml
    Start-Sleep -Seconds 1
}

function Start-ServerBackground {
    param([string]$ExePath, [string[]]$Arguments)
    Initialize-ProfileDir
    
    $argString = $Arguments -join " "
    Write-Color "  [~] Iniciando servidor en segundo plano..." "Cyan"
    
    $proc = Start-Process -FilePath $ExePath -ArgumentList $argString -WindowStyle Hidden -PassThru
    
    if ($proc) {
        $proc.Id | Out-File -FilePath $Script:ServerPidFile -Encoding UTF8
        Write-Color "  [OK] Servidor iniciado con PID: $($proc.Id)" "Green"
        Write-Color "  Puedes detenerlo desde la opcion 'Monitor de Servidor'." "DarkGray"
    } else {
        Write-Color "  [X] Fallo al iniciar el proceso." "Red"
    }
    Start-Sleep -Seconds 2
}

function Show-ServerMonitor {
    Write-Host ""
    Show-Separator "Monitor de Servidor" "Cyan"
    
    if (-not (Test-Path $Script:ServerPidFile)) {
        Write-Color "  [!] No hay ningun servidor registrado ejecutandose en segundo plano." "Yellow"
        Start-Sleep -Seconds 2
        return
    }
    
    $pidStr = Get-Content $Script:ServerPidFile | Out-String
    $pidNum = 0
    if ([int]::TryParse($pidStr.Trim(), [ref]$pidNum)) {
        $proc = Get-Process -Id $pidNum -ErrorAction SilentlyContinue
        if ($proc -and $proc.ProcessName -match "llama") {
            Write-ColorLine @(
                @{ Text = "  [+] Estado: "; Color = "White" },
                @{ Text = "EN EJECUCION"; Color = "Green" },
                @{ Text = " (PID: $pidNum)"; Color = "DarkGray" }
            )
            Write-Host ""
            
            Write-Host ""
            $action = Show-Menu -Title "Servidor en Ejecucion" -Options @("[A] Abrir Chat en Navegador (Web UI)", "[B] Detener Servidor", "[X] Volver") -ShowBack
            if ($action -eq 0) {
                Start-WebUI
            } elseif ($action -eq 1) {
                Stop-Process -Id $pidNum -Force
                Remove-Item $Script:ServerPidFile -Force
                Write-Color "  [OK] Servidor detenido." "Green"
            }

        } else {
            Write-Color "  [X] El proceso PID $pidNum ya no se encuentra en ejecucion." "Red"
            Remove-Item $Script:ServerPidFile -Force
        }
    } else {
        Remove-Item $Script:ServerPidFile -Force
    }
    Start-Sleep -Seconds 1
}


function Invoke-AutoUpdater {
    Write-Host ""
    Show-Separator "Actualizador de llama.cpp" "Cyan"
    Write-Color "  Consultando GitHub buscando la ultima version de Windows..." "DarkGray"
    
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/ggerganov/llama.cpp/releases/latest" -UseBasicParsing
        $winAsset = $release.assets | Where-Object { $_.name -match "win" -and $_.name -match "bin" -and $_.name -match "\.zip" } | Select-Object -First 1
        
        if (-not $winAsset) {
            Write-Color "  [X] No se encontro un archivo .zip para Windows en la ultima release." "Red"
            Start-Sleep -Seconds 2
            return
        }
        
        Write-ColorLine @(
            @{ Text = "  Version actual remota: "; Color = "DarkGray" },
            @{ Text = $release.tag_name; Color = "Green" }
        )
        
        $confirm = Show-Confirm -Message "Descargar e instalar $( $winAsset.name )?" -Default $true
        if (-not $confirm) { return }
        
        $tempZip = Join-Path $env:TEMP "llama_update.zip"
        $tempExtract = Join-Path $env:TEMP "llama_update_extract"
        
        Write-Color "  Descargando (puede tardar un momento)..." "Yellow"
        Invoke-WebRequest -Uri $winAsset.browser_download_url -OutFile $tempZip -UseBasicParsing
        
        Write-Color "  Extrayendo archivos..." "Yellow"
        if (Test-Path $tempExtract) { Remove-Item $tempExtract -Recurse -Force }
        Expand-Archive -Path $tempZip -DestinationPath $tempExtract -Force
        
        # Encontrar los .exe extraidos
        $exes = Get-ChildItem -Path $tempExtract -Filter "*.exe" -File -Recurse
        if ($exes.Count -eq 0) {
            Write-Color "  [X] No se encontraron ejecutables en el archivo descargado." "Red"
        } else {
            # Cerrar procesos que podrian estar usando los exes
            Get-Process -Name "llama*" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
            
            foreach ($exe in $exes) {
                $dest = Join-Path $Script:BinPath $exe.Name
                Copy-Item -Path $exe.FullName -Destination $dest -Force
            }
            
            # Copiar DLLs tambien
            $dlls = Get-ChildItem -Path $tempExtract -Filter "*.dll" -File -Recurse
            foreach ($dll in $dlls) {
                $dest = Join-Path $Script:BinPath $dll.Name
                Copy-Item -Path $dll.FullName -Destination $dest -Force
            }
            
            Write-Color "  [OK] Actualizacion completada con exito! Se copiaron $($exes.Count) ejecutables y $($dlls.Count) DLLs." "Green"
        }
        
        # Limpieza
        Remove-Item $tempZip -Force -ErrorAction SilentlyContinue
        Remove-Item $tempExtract -Recurse -Force -ErrorAction SilentlyContinue
        
    } catch {
        Write-Color "  [X] Error durante la actualizacion: $($_.Exception.Message)" "Red"
    }
    
    Write-Host "  Presiona cualquier tecla..." -ForegroundColor "DarkGray"
    [Console]::ReadKey($true) | Out-Null
}


function Invoke-DocumentAnalyzer {
    Write-Host ""
    Show-Separator "Analizador de Documentos" "Cyan"
    
    Write-Color "  Se leera un archivo de texto y se le pedira al modelo que lo analice." "DarkGray"
    
    $docPath = Read-ValidInput -Prompt "Ruta al archivo .txt a analizar" -ValidationType "path" -Required
    if (-not $docPath.EndsWith(".txt")) {
        $confirm = Show-Confirm -Message "El archivo no parece ser .txt, continuar?" -Default $false
        if (-not $confirm) { return }
    }
    
    $promptMsg = Read-ValidInput -Prompt "Prompt (Instruccion para el modelo)" -Default "Resume los puntos principales de este documento." -Required
    
    $modelPath = Browse-ForModel
    if (-not $modelPath) { return }
    
    $ctx = Read-ValidInput -Prompt "Longitud de contexto (-c) (Asegurate que quepa el documento)" -Default "8192" -ValidationType "int"
    
    $cliExe = Join-Path $Script:BinPath "llama-cli.exe"
    
    $allArgs = @("-m", "`"$modelPath`"", "-c", "$ctx", "--file", "`"$docPath`"", "-p", "`"$promptMsg`"", "-n", "-1")
    
    Write-Color "  Preparando para analizar..." "Yellow"
    Invoke-LlamaCommand -ExePath $cliExe -Arguments $allArgs
}

function Invoke-HfDownloader {
    Write-Host ""
    Show-Separator "Descargador de HuggingFace" "Cyan"
    Write-Color "  Introduce la URL directa de descarga del archivo .gguf" "DarkGray"
    Write-Color "  Ejemplo: https://huggingface.co/TheBloke/Mistral-7B-Instruct-v0.2-GGUF/resolve/main/mistral-7b-instruct-v0.2.Q4_K_M.gguf" "DarkGray"
    
    $url = Read-ValidInput -Prompt "URL" -Required
    if (-not $url.Contains("huggingface.co") -or -not $url.EndsWith(".gguf")) {
        $proceed = Show-Confirm -Message "La URL no parece un enlace directo GGUF de HuggingFace. Continuar?" -Default $false
        if (-not $proceed) { return }
    }
    
    if (-not (Test-Path $Script:ModelsPath)) { New-Item -Path $Script:ModelsPath -ItemType Directory | Out-Null }
    
    $filename = $url.Split('/')[-1]
    if ($filename.Contains("?")) { $filename = $filename.Split('?')[0] }
    
    $outPath = Join-Path $Script:ModelsPath $filename
    
    Write-Host ""
    Write-Color "  Descargando: $filename" "Yellow"
    Write-Color "  Destino: $outPath" "DarkGray"
    Write-Color "  Por favor espera, esto puede tardar dependiendo de tu conexion..." "Cyan"
    
    try {
        Invoke-WebRequest -Uri $url -OutFile $outPath -UseBasicParsing
        Write-Color "  [OK] Descarga completada con exito!" "Green"
    } catch {
        Write-Color "  [X] Error durante la descarga: $($_.Exception.Message)" "Red"
    }
    
    Write-Host "  Presiona cualquier tecla..." -ForegroundColor "DarkGray"
    [Console]::ReadKey($true) | Out-Null
}

function Show-GgufMetadata {
    param([string]$ModelPath)
    
    $cliExe = Join-Path $Script:BinPath "llama-cli.exe"
    if (-not (Test-Path $cliExe)) {
        Write-Color "  [X] No se encontro llama-cli.exe para leer metadata." "Red"
        return
    }

    Write-Host ""
    Show-Separator "Leyendo Metadata del Modelo" "Cyan"
    Write-Color "  Analizando archivo: $(Split-Path $ModelPath -Leaf)" "DarkGray"
    
    try {
        $output = & $cliExe -m $ModelPath -n 1 -c 128 2>&1
        $arch = ($output | Select-String "llm_load_print_meta: model type").Line
        $vocab = ($output | Select-String "llm_load_print_meta: vocab type").Line
        $params = ($output | Select-String "llm_load_print_meta: model params").Line
        $size = ($output | Select-String "llm_load_print_meta: model size").Line
        
        Write-Host ""
        if ($arch)   { Write-Color "  [+] Arquitectura : $($arch.Split('=')[-1].Trim())" "Green" }
        if ($vocab)  { Write-Color "  [+] Vocabulario  : $($vocab.Split('=')[-1].Trim())" "Green" }
        if ($params) { Write-Color "  [+] Parametros   : $($params.Split('=')[-1].Trim())" "Green" }
        if ($size)   { Write-Color "  [+] Tamano       : $($size.Split('=')[-1].Trim())" "Green" }
        
        $ctx = ($output | Select-String "n_ctx_train").Line
        if ($ctx) { Write-Color "  [+] Contexto base: $($ctx.Split('=')[-1].Trim()) tokens" "Green" }
        
    } catch {
        Write-Color "  [X] Error al leer metadata." "Red"
    }
    
    Write-Host ""
    Write-Host "  Presiona cualquier tecla..." -ForegroundColor "DarkGray"
    [Console]::ReadKey($true) | Out-Null
}

# ================================================================
#                 CONSTRUCTORES DE PARAMETROS
# ================================================================

function Build-CommonParams {
    <#
    .SYNOPSIS
        Dialogo interactivo para parametros comunes (modelo, contexto, hilos, GPU).
    #>
    param([switch]$SkipModel)

    $params = @{}

    # 1. Modelo
    if (-not $SkipModel) {
        $modelPath = Browse-ForModel
        if (-not $modelPath) {
            Write-Color "  [X] Se requiere un modelo para continuar." "Red"
            return $null
        }
        $params["model"] = $modelPath
    }

    # 2. Contexto
    $ctx = Read-ValidInput -Prompt "Longitud de contexto (-c)" -Default "4096" -ValidationType "int" -Min 128 -Max 1048576
    if ($ctx) { $params["context"] = $ctx }

    # 3. Hilos
    $threads = Read-ValidInput -Prompt "Hilos de CPU (-t)" -Default "$Script:DefaultThreads" -ValidationType "int" -Min 1 -Max 256
    if ($threads) { $params["threads"] = $threads }

    # 4. GPU Layers
    $defGpu = if ($Script:HasDedicatedGpu) { "99" } else { "0" }
    if ($Script:HasDedicatedGpu) { Write-Color "  [+] GPU Detectada: Se recomienda 99 para descargar todo a VRAM." "Green" }
    $ngl = Read-ValidInput -Prompt "Capas GPU (-ngl, 0=solo CPU)" -Default $defGpu -ValidationType "int" -Min 0 -Max 999
    if ($ngl -and $ngl -gt 0) { $params["gpu_layers"] = $ngl }

    return $params
}

function Format-ArgsArray {
    <#
    .SYNOPSIS
        Convierte un hashtable de parametros a un array de argumentos de linea de comandos.
    #>
    param([hashtable]$Params)

    $argsArr = @()

    if ($Params.ContainsKey("model"))      { $argsArr += "-m";   $argsArr += ('"' + $Params['model'] + '"') }
    if ($Params.ContainsKey("context"))    { $argsArr += "-c";   $argsArr += "$($Params['context'])" }
    if ($Params.ContainsKey("threads"))    { $argsArr += "-t";   $argsArr += "$($Params['threads'])" }
    if ($Params.ContainsKey("gpu_layers")) { $argsArr += "-ngl"; $argsArr += "$($Params['gpu_layers'])" }

    return $argsArr
}


# ================================================================
#                    WIZARDS POR HERRAMIENTA
# ================================================================

function Invoke-CliWizard {
    <#
    .SYNOPSIS
        Wizard guiado para llama-cli.exe - el modo chat/generacion interactivo.
    #>
    param([string]$ExePath)

    Write-Host ""
    Show-Separator "WIZARD: llama-cli" "Magenta"
    Write-Color "  Configuracion paso a paso para generacion de texto" "DarkGray"

    # Parametros comunes
    $common = Build-CommonParams
    if (-not $common) { return $null }

    $allArgs = [System.Collections.ArrayList]@()
    $allArgs.AddRange((Format-ArgsArray $common))

    # --- Parametros especificos de CLI ---

    # Modo interactivo
    $interactive = Show-Confirm -Message "Activar modo conversacion interactiva? (-cnv)" -Default $true
    if ($interactive) {
        $allArgs.Add("-cnv") | Out-Null
    }

    # Temperatura
    $temp = Read-ValidInput -Prompt "Temperatura (--temp, creatividad)" -Default "0.7" -ValidationType "float"
    if ($temp) { $allArgs.Add("--temp") | Out-Null; $allArgs.Add("$temp") | Out-Null }

    # Top-P
    $topP = Read-ValidInput -Prompt "Top-P (--top-p, nucleus sampling)" -Default "0.9" -ValidationType "float"
    if ($topP) { $allArgs.Add("--top-p") | Out-Null; $allArgs.Add("$topP") | Out-Null }

    # Top-K
    $topK = Read-ValidInput -Prompt "Top-K (--top-k)" -Default "40" -ValidationType "int" -Min 0 -Max 1000
    if ($topK) { $allArgs.Add("--top-k") | Out-Null; $allArgs.Add("$topK") | Out-Null }

    # Repeat penalty
    $repeatPenalty = Read-ValidInput -Prompt "Penalizacion de repeticion (--repeat-penalty)" -Default "1.1" -ValidationType "float"
    if ($repeatPenalty) { $allArgs.Add("--repeat-penalty") | Out-Null; $allArgs.Add("$repeatPenalty") | Out-Null }

    # System prompt
    $sysPrompt = Prompt-SystemPromptLibrary
    if ($sysPrompt) {
        $allArgs.Add("--system-prompt") | Out-Null
        $allArgs.Add(('"' + $sysPrompt + '"')) | Out-Null
    }

    # Seed
    $wantSeed = Show-Confirm -Message "Fijar semilla para reproducibilidad?" -Default $false
    if ($wantSeed) {
        $seed = Read-ValidInput -Prompt "Semilla (-s)" -Default "42" -ValidationType "int" -Min 0
        if ($seed) { $allArgs.Add("-s") | Out-Null; $allArgs.Add("$seed") | Out-Null }
    }

    # Tokens maximos
    $maxTokens = Read-ValidInput -Prompt "Tokens maximos a generar (-n, -1=infinito)" -Default "-1" -ValidationType "int"
    if ($maxTokens) { $allArgs.Add("-n") | Out-Null; $allArgs.Add("$maxTokens") | Out-Null }

    # Avanzado
    $advanced = Show-Confirm -Message "Configurar opciones avanzadas?" -Default $false
    if ($advanced) {
        # Min-P
        $minP = Read-ValidInput -Prompt "Min-P (--min-p)" -Default "0.05" -ValidationType "float"
        if ($minP) { $allArgs.Add("--min-p") | Out-Null; $allArgs.Add("$minP") | Out-Null }

        # Batch size
        $batch = Read-ValidInput -Prompt "Batch size (-b)" -Default "2048" -ValidationType "int" -Min 1
        if ($batch) { $allArgs.Add("-b") | Out-Null; $allArgs.Add("$batch") | Out-Null }

        # Flash attention
        $flash = Show-Confirm -Message "Activar Flash Attention? (-fa)" -Default $false
        if ($flash) { $allArgs.Add("-fa") | Out-Null }

        # Argumentos adicionales en texto libre
        $extra = Read-ValidInput -Prompt "Argumentos adicionales (texto libre, Enter para omitir)"
        if ($extra) {
            $extra.Split(" ") | ForEach-Object { $allArgs.Add($_) | Out-Null }
        }
    }

    return @{
        Executable = $ExePath
        Arguments  = $allArgs.ToArray()
        ToolName   = "llama-cli"
        Config     = $common
    }
}

function Invoke-ServerWizard {
    <#
    .SYNOPSIS
        Wizard guiado para llama-server.exe - servidor API compatible con OpenAI.
    #>
    param([string]$ExePath)

    Write-Host ""
    Show-Separator "WIZARD: llama-server" "Magenta"
    Write-Color "  Configuracion del servidor API (compatible con OpenAI)" "DarkGray"

    # Parametros comunes
    $common = Build-CommonParams
    if (-not $common) { return $null }

    $allArgs = [System.Collections.ArrayList]@()
    $allArgs.AddRange((Format-ArgsArray $common))

    # --- Parametros del servidor ---

    # Host
    $listenHost = Read-ValidInput -Prompt "Direccion de escucha (--host)" -Default "127.0.0.1"
    if ($listenHost) { $allArgs.Add("--host") | Out-Null; $allArgs.Add("$listenHost") | Out-Null }

    # Puerto
    $port = Read-ValidInput -Prompt "Puerto (--port)" -Default "8080" -ValidationType "port"
    if ($port) { $allArgs.Add("--port") | Out-Null; $allArgs.Add("$port") | Out-Null }

    # Slots paralelos
    $slots = Read-ValidInput -Prompt "Slots paralelos (-np)" -Default "1" -ValidationType "int" -Min 1 -Max 64
    if ($slots) { $allArgs.Add("-np") | Out-Null; $allArgs.Add("$slots") | Out-Null }

    # API Key
    $apiKey = Read-ValidInput -Prompt "API Key (--api-key, Enter para omitir)"
    if ($apiKey) { $allArgs.Add("--api-key") | Out-Null; $allArgs.Add("$apiKey") | Out-Null }

    # Embeddings
    $embedding = Show-Confirm -Message "Activar modo de embeddings? (--embedding)" -Default $false
    if ($embedding) { $allArgs.Add("--embedding") | Out-Null }

    # System prompt file
    $sysPrompt = Prompt-SystemPromptLibrary
    if ($sysPrompt) {
        Write-Color "  [!] Nota: El servidor recibe system prompts principalmente via API request, pero lo configuraremos." "Yellow"
        $allArgs.Add("--system-prompt") | Out-Null
        $allArgs.Add(('"' + $sysPrompt + '"')) | Out-Null
    }
    $sysFile = ""
    if ($sysFile -and (Test-Path $sysFile)) {
        $allArgs.Add("--system-prompt-file") | Out-Null
        $allArgs.Add(('"' + $sysFile + '"')) | Out-Null
    } elseif ($sysFile) {
        Write-Color "  [!] Archivo no encontrado, se omitira." "Yellow"
    }

    # Flash attention
    $flash = Show-Confirm -Message "Activar Flash Attention? (-fa)" -Default $false
    if ($flash) { $allArgs.Add("-fa") | Out-Null }

    # Chat template
    $template = Read-ValidInput -Prompt "Chat template (--chat-template, Enter para auto-detectar)"
    if ($template) { $allArgs.Add("--chat-template") | Out-Null; $allArgs.Add("$template") | Out-Null }

    # Extra
    $extra = Read-ValidInput -Prompt "Argumentos adicionales (texto libre, Enter para omitir)"
    if ($extra) {
        $extra.Split(" ") | ForEach-Object { $allArgs.Add($_) | Out-Null }
    }

    return @{
        Executable = $ExePath
        Arguments  = $allArgs.ToArray()
        ToolName   = "llama-server"
        Config     = $common
    }
}

function Invoke-QuantizeWizard {
    <#
    .SYNOPSIS
        Wizard guiado para llama-quantize.exe - cuantizacion de modelos.
    #>
    param([string]$ExePath)

    Write-Host ""
    Show-Separator "WIZARD: llama-quantize" "Magenta"
    Write-Color "  Cuantizacion de modelos GGUF a formatos mas pequenos" "DarkGray"

    # 1. Modelo de entrada
    Write-Host ""
    Write-Color "  -- Paso 1: Modelo de entrada --" "Cyan"
    $inputModel = Browse-ForModel
    if (-not $inputModel) { return $null }

    # 2. Tipo de cuantizacion
    Write-Host ""
    Write-Color "  -- Paso 2: Tipo de cuantizacion --" "Cyan"
    $quantNames = $Script:QuantTypes | ForEach-Object { $_.Name }
    $quantDescs = $Script:QuantTypes | ForEach-Object { $_.Desc }

    $quantIdx = Show-Menu -Title "Selecciona el tipo de cuantizacion" -Options $quantNames -Descriptions $quantDescs -ShowBack
    if ($quantIdx -eq -1) { return $null }
    $quantType = $Script:QuantTypes[$quantIdx].Name

    Write-ColorLine @(
        @{ Text = "  [OK] Tipo: "; Color = "Green" },
        @{ Text = $quantType; Color = "White" }
    )

    # 3. Ruta de salida
    Write-Host ""
    Write-Color "  -- Paso 3: Archivo de salida --" "Cyan"
    $inputFileName = [System.IO.Path]::GetFileNameWithoutExtension($inputModel)
    $inputDir      = [System.IO.Path]::GetDirectoryName($inputModel)
    $defaultOutput = Join-Path $inputDir "${inputFileName}-${quantType}.gguf"

    $outputModel = Read-ValidInput -Prompt "Ruta de salida" -Default $defaultOutput
    if (-not $outputModel) { $outputModel = $defaultOutput }

    # 4. Opciones adicionales
    Write-Host ""
    Write-Color "  -- Paso 4: Opciones adicionales --" "Cyan"

    $allArgs = [System.Collections.ArrayList]@()

    # Importance matrix (opcional)
    $useImatrix = Show-Confirm -Message "Usar matriz de importancia? (mejora calidad en quants bajos)" -Default $false
    if ($useImatrix) {
        $imatrixPath = Read-ValidInput -Prompt "Ruta al archivo de imatrix" -ValidationType "path" -Required
        if ($imatrixPath) {
            $allArgs.Add("--imatrix") | Out-Null
            $allArgs.Add(('"' + $imatrixPath + '"')) | Out-Null
        }
    }

    # Hilos
    $threads = Read-ValidInput -Prompt "Hilos de CPU (-t)" -Default "$Script:DefaultThreads" -ValidationType "int" -Min 1 -Max 256

    # Construir argumentos: llama-quantize [opciones] modelo_entrada modelo_salida tipo
    $allArgs.Add(('"' + $inputModel + '"')) | Out-Null
    $allArgs.Add(('"' + $outputModel + '"')) | Out-Null
    $allArgs.Add($quantType) | Out-Null
    $allArgs.Add("-t") | Out-Null
    $allArgs.Add("$threads") | Out-Null

    return @{
        Executable = $ExePath
        Arguments  = $allArgs.ToArray()
        ToolName   = "llama-quantize"
        Config     = @{
            input  = $inputModel
            output = $outputModel
            type   = $quantType
        }
    }
}

function Invoke-BenchWizard {
    <#
    .SYNOPSIS
        Wizard guiado para llama-bench.exe - benchmarking de rendimiento.
    #>
    param([string]$ExePath)

    Write-Host ""
    Show-Separator "WIZARD: llama-bench" "Magenta"
    Write-Color "  Benchmark de rendimiento del modelo" "DarkGray"

    # Modelo
    $modelPath = Browse-ForModel
    if (-not $modelPath) { return $null }

    $allArgs = [System.Collections.ArrayList]@()
    $allArgs.Add("-m") | Out-Null
    $allArgs.Add(('"' + $modelPath + '"')) | Out-Null

    # Hilos
    $threads = Read-ValidInput -Prompt "Hilos de CPU (-t)" -Default "$Script:DefaultThreads" -ValidationType "int" -Min 1 -Max 256
    if ($threads) { $allArgs.Add("-t") | Out-Null; $allArgs.Add("$threads") | Out-Null }

    # GPU Layers
    $defGpu = if ($Script:HasDedicatedGpu) { "99" } else { "0" }
    if ($Script:HasDedicatedGpu) { Write-Color "  [+] GPU Detectada: Se recomienda 99 para descargar todo a VRAM." "Green" }
    $ngl = Read-ValidInput -Prompt "Capas GPU (-ngl, 0=solo CPU)" -Default $defGpu -ValidationType "int" -Min 0 -Max 999
    if ($ngl -and $ngl -gt 0) { $allArgs.Add("-ngl") | Out-Null; $allArgs.Add("$ngl") | Out-Null }

    # Tamanos de contexto para probar
    $ctxSizes = Read-ValidInput -Prompt "Tamanos de contexto a probar (separados por coma)" -Default "512,1024,2048"
    if ($ctxSizes) {
        $allArgs.Add("-c") | Out-Null
        $allArgs.Add("$ctxSizes") | Out-Null
    }

    # Batch sizes
    $batchSizes = Read-ValidInput -Prompt "Batch sizes a probar (separados por coma)" -Default "512"
    if ($batchSizes) {
        $allArgs.Add("-b") | Out-Null
        $allArgs.Add("$batchSizes") | Out-Null
    }

    # Repeticiones
    $reps = Read-ValidInput -Prompt "Repeticiones por prueba (-r)" -Default "3" -ValidationType "int" -Min 1 -Max 100
    if ($reps) { $allArgs.Add("-r") | Out-Null; $allArgs.Add("$reps") | Out-Null }

    # Formato de salida
    $formatIdx = Show-Menu -Title "Formato de salida" -Options @("Markdown (md)", "CSV", "JSON", "SQL") -ShowBack
    if ($formatIdx -ge 0) {
        $formats = @("md", "csv", "json", "sql")
        $allArgs.Add("-o") | Out-Null
        $allArgs.Add($formats[$formatIdx]) | Out-Null
    }

    return @{
        Executable = $ExePath
        Arguments  = $allArgs.ToArray()
        ToolName   = "llama-bench"
        Config     = @{ model = $modelPath }
    }
}


function Invoke-MultimodalWizard {
    param([string]$ExePath)

    Write-Host ""
    Show-Separator "WIZARD: Multimodal (Vision)" "Magenta"
    Write-Color "  Configuracion para analisis de imagenes" "DarkGray"

    $common = Build-CommonParams
    if (-not $common) { return $null }

    $allArgs = [System.Collections.ArrayList]@()
    $allArgs.AddRange((Format-ArgsArray $common))

    $imagePath = Read-ValidInput -Prompt "Ruta a la imagen (.jpg, .png)" -ValidationType "path" -Required
    $allArgs.Add("--image") | Out-Null
    $allArgs.Add(('"' + $imagePath + '"')) | Out-Null

    $promptStr = Read-ValidInput -Prompt "Pregunta sobre la imagen" -Default "Describe esta imagen en detalle." -Required
    $allArgs.Add("-p") | Out-Null
    $allArgs.Add(('"' + $promptStr + '"')) | Out-Null
    
    # Proyector multimodal (mmproj) si lo requiere llava
    $needsMmproj = Show-Confirm -Message "Este modelo requiere un archivo mmproj (.gguf multimodal projector)?" -Default $false
    if ($needsMmproj) {
        Write-Color "  Busca el archivo del proyector:" "Yellow"
        $mmproj = Browse-ForModel
        if ($mmproj) {
            $allArgs.Add("--mmproj") | Out-Null
            $allArgs.Add(('"' + $mmproj + '"')) | Out-Null
        }
    }

    return @{
        Executable = $ExePath
        Arguments  = $allArgs.ToArray()
        ToolName   = "llama-multimodal"
        Config     = $common
    }
}

function Invoke-TtsWizard {
    param([string]$ExePath)

    Write-Host ""
    Show-Separator "WIZARD: Texto a Voz (TTS)" "Magenta"
    Write-Color "  Generador de audio usando modelos TTS" "DarkGray"

    $common = Build-CommonParams
    if (-not $common) { return $null }

    $allArgs = [System.Collections.ArrayList]@()
    $allArgs.AddRange((Format-ArgsArray $common))

    $text = Read-ValidInput -Prompt "Texto a narrar" -Required
    $allArgs.Add("-p") | Out-Null
    $allArgs.Add(('"' + $text + '"')) | Out-Null

    $outFile = Read-ValidInput -Prompt "Ruta de salida del audio (.wav)" -Default "salida.wav"
    $allArgs.Add("-o") | Out-Null
    $allArgs.Add(('"' + $outFile + '"')) | Out-Null

    return @{
        Executable = $ExePath
        Arguments  = $allArgs.ToArray()
        ToolName   = "llama-tts"
        Config     = $common
    }
}

function Invoke-GenericWizard {
    <#
    .SYNOPSIS
        Wizard generico para cualquier herramienta no cubierta por un wizard especifico.
        Ofrece seleccion de modelo + argumentos libres.
    #>
    param(
        [string]$ExePath,
        [string]$ToolName
    )

    Write-Host ""
    Show-Separator "WIZARD: $ToolName" "Magenta"
    Write-Color "  Configuracion generica para $ToolName" "DarkGray"

    $allArgs = [System.Collections.ArrayList]@()

    # Necesita modelo?
    $needsModel = Show-Confirm -Message "Esta herramienta necesita un modelo .gguf?" -Default $true
    if ($needsModel) {
        $modelPath = Browse-ForModel
        if ($modelPath) {
            $allArgs.Add("-m") | Out-Null
            $allArgs.Add(('"' + $modelPath + '"')) | Out-Null
        }
    }

    # Mostrar ayuda del ejecutable
    $showHelp = Show-Confirm -Message "Mostrar la ayuda del ejecutable (--help) primero?" -Default $false
    if ($showHelp) {
        Write-Host ""
        Show-Separator "AYUDA: $ToolName" "DarkYellow"
        try {
            & $ExePath --help 2>&1 | ForEach-Object { Write-Host "  $_" -ForegroundColor "DarkGray" }
        } catch {
            Write-Color "  [!] No se pudo obtener la ayuda." "Yellow"
        }
        Show-Separator
        Write-Host ""
    }

    # Argumentos libres
    Write-Host ""
    Write-Color "  Ingresa los argumentos separados por espacio." "DarkGray"
    Write-Color "  Ejemplo: -c 4096 -t 8 --verbose" "DarkGray"
    $extra = Read-ValidInput -Prompt "Argumentos"
    if ($extra) {
        # Parseo basico de argumentos respetando comillas
        $tokens = [System.Collections.ArrayList]@()
        $current = ""
        $inQuote = $false
        foreach ($char in $extra.ToCharArray()) {
            if ($char -eq '"') {
                $inQuote = -not $inQuote
                $current += $char
            } elseif ($char -eq ' ' -and -not $inQuote) {
                if ($current.Length -gt 0) {
                    $tokens.Add($current) | Out-Null
                    $current = ""
                }
            } else {
                $current += $char
            }
        }
        if ($current.Length -gt 0) { $tokens.Add($current) | Out-Null }
        $allArgs.AddRange($tokens.ToArray())
    }

    return @{
        Executable = $ExePath
        Arguments  = $allArgs.ToArray()
        ToolName   = $ToolName
        Config     = @{}
    }
}


# ================================================================
#                     SISTEMA DE PERFILES
# ================================================================

function Initialize-ProfileDir {
    <#
    .SYNOPSIS
        Crea la carpeta de perfiles si no existe.
    #>
    if (-not (Test-Path $Script:ProfilePath)) {
        New-Item -Path $Script:ProfilePath -ItemType Directory -Force | Out-Null
        Write-Color "  [+] Carpeta de perfiles creada: $Script:ProfilePath" "DarkGray"
    }
}

function Save-Profile {
    <#
    .SYNOPSIS
        Guarda una configuracion de comando como perfil JSON reutilizable.
    #>
    param([hashtable]$CommandInfo)

    Initialize-ProfileDir

    $profileName = Read-ValidInput -Prompt "Nombre del perfil (ej: chat-mistral-7b)" -Required
    if (-not $profileName) { return }

    # Sanitizar nombre de archivo
    $safeFileName = ($profileName -replace '[^\w\-\.]', '_') + ".json"

    $description = Read-ValidInput -Prompt "Descripcion (opcional)"

    $profile = @{
        Name         = $profileName
        Description  = $description
        ToolName     = $CommandInfo.ToolName
        Executable   = $CommandInfo.Executable
        Arguments    = $CommandInfo.Arguments
        Config       = $CommandInfo.Config
        CreatedAt    = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
        LastUsed     = $null
        Version      = $Script:Version
    }

    $filePath = Join-Path $Script:ProfilePath $safeFileName
    try { $profile | ConvertTo-Json -Depth 5 | Set-Content -Path $filePath -Encoding UTF8 } catch { Write-Color "  [X] Error al guardar el perfil." "Red"; return }

    Write-Host ""
    Write-ColorLine @(
        @{ Text = "  [OK] Perfil guardado: "; Color = "Green" },
        @{ Text = $profileName; Color = "White" },
        @{ Text = " -> $filePath"; Color = "DarkGray" }
    )
}

function Get-AllProfiles {
    <#
    .SYNOPSIS
        Lee todos los perfiles guardados y los devuelve como array.
    #>
    Initialize-ProfileDir

    $profiles = @()
    $files = Get-ChildItem -Path $Script:ProfilePath -Filter "*.json" -File -ErrorAction SilentlyContinue

    foreach ($file in $files) {
        try {
            $data = Get-Content -Path $file.FullName -Raw | ConvertFrom-Json
            $profiles += [PSCustomObject]@{
                Name        = $data.Name
                Description = $data.Description
                ToolName    = $data.ToolName
                Executable  = $data.Executable
                Arguments   = @($data.Arguments)
                Config      = $data.Config
                CreatedAt   = $data.CreatedAt
                LastUsed    = $data.LastUsed
                FilePath    = $file.FullName
            }
        } catch {
            $fname = $file.Name
            Write-Color "  [!] Error al leer perfil: $fname" "Yellow"
        }
    }

    return $profiles | Sort-Object Name
}

function Update-ProfileLastUsed {
    <#
    .SYNOPSIS
        Actualiza la fecha de ultimo uso de un perfil.
    #>
    param([string]$FilePath)

    try {
        $data = Get-Content -Path $FilePath -Raw | ConvertFrom-Json
        $data.LastUsed = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
        $data | ConvertTo-Json -Depth 5 | Set-Content -Path $FilePath -Encoding UTF8
    } catch {
        # No es critico si falla
    }
}

function Show-ProfileManager {
    <#
    .SYNOPSIS
        Interfaz completa para gestionar perfiles: ejecutar, ver, eliminar.
    #>

    while ($true) {
        $profiles = Get-AllProfiles

        if ($profiles.Count -eq 0) {
            Write-Host ""
            Write-Color "  [!] No hay perfiles guardados." "Yellow"
            Write-Color "  Ejecuta una herramienta y elige 'Guardar como perfil' al final." "DarkGray"
            Write-Host ""
            Write-Host "  Presiona cualquier tecla para volver..." -ForegroundColor "DarkGray"
            [Console]::ReadKey($true) | Out-Null
            return
        }

        # Construir opciones del menu
        $optNames = @()
        $optDescs = @()
        foreach ($p in $profiles) {
            $optNames += "> $($p.Name)"
            $lastUsedText = "Nunca usado"
            if ($p.LastUsed) { $lastUsedText = "Ultimo uso: $($p.LastUsed)" }
            $optDescs += "$($p.ToolName) | $lastUsedText | $($p.Description)"
        }
        $optNames += "[DEL] Eliminar un perfil"
        $optDescs += "Borrar perfiles guardados"

        $sel = Show-Menu -Title "Perfiles Guardados" -Options $optNames -Descriptions $optDescs -ShowBack

        if ($sel -eq -1) { return }

        if ($sel -eq $profiles.Count) {
            # Eliminar perfil
            $delNames = $profiles | ForEach-Object { $_.Name }
            $delIdx = Show-Menu -Title "Que perfil deseas eliminar?" -Options $delNames -ShowBack
            if ($delIdx -ge 0) {
                $target = $profiles[$delIdx]
                $confirm = Show-Confirm -Message "Estas seguro de eliminar '$($target.Name)'?" -Default $false
                if ($confirm) {
                    Remove-Item -Path $target.FilePath -Force
                    Write-Color "  [OK] Perfil eliminado: $($target.Name)" "Green"
                    Start-Sleep -Milliseconds 800
                }
            }
            continue
        }

        # Ejecutar perfil seleccionado
        $selected = $profiles[$sel]

        # Verificar que el ejecutable existe
        if (-not (Test-Path $selected.Executable)) {
            Write-Color "  [X] Ejecutable no encontrado: $($selected.Executable)" "Red"
            Write-Color "  El perfil podria estar desactualizado." "Yellow"
            continue
        }

        Write-Host ""
        Show-Separator "Perfil: $($selected.Name)" "Green"
        Show-CommandPreview -ExePath $selected.Executable -Arguments $selected.Arguments

        $runChoice = Show-Menu -Title "Que deseas hacer?" -Options @(
            ">> Ejecutar ahora"
            "[COPY] Copiar comando al portapapeles"
            "<< Volver"
        )

        switch ($runChoice) {
            0 {
                Update-ProfileLastUsed -FilePath $selected.FilePath
                Invoke-LlamaCommand -ExePath $selected.Executable -Arguments $selected.Arguments
            }
            1 {
                $cmdText = '"' + $selected.Executable + '" ' + ($selected.Arguments -join ' ')
                Set-Clipboard -Value $cmdText
                Write-Color "  [OK] Comando copiado al portapapeles." "Green"
                Start-Sleep -Milliseconds 800
            }
        }
    }
}


# ================================================================
#                     CAPA DE EJECUCION
# ================================================================

function Show-CommandPreview {
    <#
    .SYNOPSIS
        Muestra una vista previa estilizada del comando que se va a ejecutar.
    #>
    param(
        [string]$ExePath,
        [string[]]$Arguments
    )

    Write-Host ""
    Show-Separator "Vista Previa del Comando" "DarkYellow"
    Write-Host ""

    # Nombre del ejecutable
    $exeName = [System.IO.Path]::GetFileName($ExePath)
    Write-Host "  " -NoNewline
    Write-Host $exeName -NoNewline -ForegroundColor "Green"

    # Argumentos agrupados
    $i = 0
    while ($i -lt $Arguments.Count) {
        $arg = $Arguments[$i]
        if ($arg.StartsWith("-")) {
            # Es un flag/opcion
            Write-Host " " -NoNewline
            Write-Host $arg -NoNewline -ForegroundColor "Cyan"
            # Si el siguiente no es un flag, es su valor
            if (($i + 1) -lt $Arguments.Count -and -not $Arguments[$i + 1].StartsWith("-")) {
                $i++
                $val = $Arguments[$i]
                # Truncar rutas largas para la visualizacion
                if ($val.Length -gt 60) {
                    $val = "..." + $val.Substring($val.Length - 55)
                }
                Write-Host " " -NoNewline
                Write-Host $val -NoNewline -ForegroundColor "Yellow"
            }
        } else {
            Write-Host " " -NoNewline
            $val = $arg
            if ($val.Length -gt 60) {
                $val = "..." + $val.Substring($val.Length - 55)
            }
            Write-Host $val -NoNewline -ForegroundColor "Yellow"
        }
        $i++
    }
    Write-Host ""
    Write-Host ""
    Show-Separator
}

function Invoke-LlamaCommand {
    <#
    .SYNOPSIS
        Ejecuta un comando de llama.cpp con streaming de salida en tiempo real.
        Captura errores y muestra diagnosticos si falla.
    #>
    param(
        [string]$ExePath,
        [string[]]$Arguments
    )

    # Verificar que el ejecutable existe
    if (-not (Test-Path $ExePath)) {
        Write-Color "  [X] ERROR: Ejecutable no encontrado: $ExePath" "Red"
        return
    }

    Write-Host ""
    Show-Separator "EJECUTANDO" "Green"
    $exeFileName = [System.IO.Path]::GetFileName($ExePath)
    $startTime = (Get-Date).ToString("HH:mm:ss")
    Write-ColorLine @(
        @{ Text = "  Herramienta: "; Color = "DarkGray" },
        @{ Text = $exeFileName; Color = "Green" }
    )
    Write-ColorLine @(
        @{ Text = "  Inicio: "; Color = "DarkGray" },
        @{ Text = $startTime; Color = "White" }
    )
    Show-Separator
    Write-Host ""

    try {
        # Construir la linea de argumentos como un solo string
        $argString = $Arguments -join " "

        # Usar Start-Process con -NoNewWindow para streaming interactivo
        # Esto permite que llama-cli capture input del teclado en modo conversacion
        $process = Start-Process -FilePath $ExePath `
                                 -ArgumentList $argString `
                                 -NoNewWindow `
                                 -Wait `
                                 -PassThru

        Write-Host ""
        Show-Separator "FIN DE EJECUCION" "DarkYellow"

        if ($process.ExitCode -eq 0) {
            Write-ColorLine @(
                @{ Text = "  [OK] Proceso termino correctamente "; Color = "Green" },
                @{ Text = "(codigo: 0)"; Color = "DarkGray" }
            )
            # Save to history safely
            try {
                # Infer ToolName from ExePath if needed
                $tName = [System.IO.Path]::GetFileNameWithoutExtension($ExePath)
                Add-History -CommandInfo @{ Executable = $ExePath; Arguments = $Arguments; ToolName = $tName }
            } catch {}
        } else {
            $exitCodeStr = $process.ExitCode.ToString()
            Write-ColorLine @(
                @{ Text = "  [!] Proceso termino con codigo: "; Color = "Yellow" },
                @{ Text = $exitCodeStr; Color = "Red" }
            )

            # Diagnostico segun codigo de error
            switch ($process.ExitCode) {
                1   { Write-Color "  -> Error general. Verifica los argumentos." "DarkGray" }
                -1  { Write-Color "  -> Error interno de llama.cpp." "DarkGray" }
                137 { Write-Color "  -> Proceso terminado por el sistema (sin memoria)." "DarkGray" }
                139 { Write-Color "  -> Segmentation fault - posible modelo corrupto." "DarkGray" }
                default { Write-Color "  -> Consulta la documentacion de llama.cpp para este codigo." "DarkGray" }
            }
        }
    } catch {
        Write-Host ""
        Write-Color "  [X] ERROR AL EJECUTAR:" "Red"
        Write-Color "  $($_.Exception.Message)" "Red"
        Write-Host ""
        Write-Color "  Sugerencias:" "Yellow"
        Write-Color "  - Verifica que el modelo .gguf no este corrupto" "DarkGray"
        Write-Color "  - Asegurate de tener suficiente RAM disponible" "DarkGray"
        Write-Color "  - Revisa que los DLLs de ggml esten en la carpeta bin/" "DarkGray"
    }

    Write-Host ""
    Write-Host "  Presiona cualquier tecla para volver al menu..." -ForegroundColor "DarkGray"
    [Console]::ReadKey($true) | Out-Null
}


# ================================================================
#              VALIDACION Y MANEJO DE ERRORES
# ================================================================

function Test-Prerequisites {
    <#
    .SYNOPSIS
        Valida que el entorno este correctamente configurado antes de iniciar.
        Retorna $true si todo esta bien, $false si hay errores criticos.
    #>

    $errors   = @()
    $warnings = @()

    # 1. Verificar carpeta bin
    if (-not (Test-Path $Script:BinPath)) {
        $errors += "No se encontro la carpeta bin/ en: $Script:BinPath"
    } else {
        $exeCount = (Get-ChildItem -Path $Script:BinPath -Filter "*.exe" -File).Count
        if ($exeCount -eq 0) {
            $errors += "La carpeta bin/ existe pero no contiene ejecutables .exe"
        }
    }

    # 2. Verificar DLLs criticos
    $criticalDlls = @("ggml-base.dll", "ggml.dll", "llama.dll", "llama-common.dll")
    foreach ($dll in $criticalDlls) {
        $dllPath = Join-Path $Script:BinPath $dll
        if (-not (Test-Path $dllPath)) {
            $warnings += "DLL faltante: $dll (algunas herramientas podrian fallar)"
        }
    }

    # 3. Verificar espacio en disco
    try {
        $drive = (Get-Item $Script:BasePath).PSDrive
        $freeGB = [Math]::Round(($drive.Free / 1GB), 1)
        if ($freeGB -lt 2) {
            $warnings += "Poco espacio en disco: $freeGB GB disponibles"
        }
    } catch {
        # No es critico
    }

    # Mostrar resultados
    if ($errors.Count -gt 0) {
        Write-Host ""
        Write-Color "  === ERRORES CRITICOS ===" "Red"
        foreach ($e in $errors) {
            Write-Color "  [X] $e" "Red"
        }
        Write-Host ""
        return $false
    }

    if ($warnings.Count -gt 0) {
        Write-Host ""
        foreach ($w in $warnings) {
            Write-Color "  [!] $w" "Yellow"
        }
    }

    return $true
}


# ================================================================
#                   FLUJO POST-EJECUCION
# ================================================================

function Show-PostWizardMenu {
    <#
    .SYNOPSIS
        Despues de configurar un comando, muestra opciones: ejecutar, guardar, copiar, cancelar.
    #>
    param([hashtable]$CommandInfo)

    if (-not $CommandInfo) { return }

    # Vista previa
    Show-CommandPreview -ExePath $CommandInfo.Executable -Arguments $CommandInfo.Arguments

    $choice = Show-Menu -Title "Que deseas hacer?" -Options @(
        ">>  Ejecutar ahora"
        "[SAVE+RUN] Guardar como perfil y ejecutar"
        "[BACKGROUND] Iniciar en segundo plano (solo servidor)",
        "[SAVE] Guardar como perfil (sin ejecutar)",
        "[COPY] Copiar comando al portapapeles",
        "[EDIT] Editar argumentos manualmente",
        "[X]  Cancelar"
    )

    switch ($choice) {
        0 {
            Invoke-LlamaCommand -ExePath $CommandInfo.Executable -Arguments $CommandInfo.Arguments
        }
        1 {
            Save-Profile -CommandInfo $CommandInfo
            Invoke-LlamaCommand -ExePath $CommandInfo.Executable -Arguments $CommandInfo.Arguments
        }
        2 {
            if ($CommandInfo.ToolName -eq "llama-server") {
                Start-ServerBackground -ExePath $CommandInfo.Executable -Arguments $CommandInfo.Arguments
            } else {
                Write-Color "  [X] Opcion solo valida para llama-server." "Red"
            }
        }
        3 {
            Save-Profile -CommandInfo $CommandInfo
        }
        4 {
            $cmdText = '"' + $CommandInfo.Executable + '" ' + ($CommandInfo.Arguments -join ' ')
            Set-Clipboard -Value $cmdText
            Write-Color "  [OK] Comando copiado al portapapeles." "Green"
            Start-Sleep -Milliseconds 1000
        }
        5 {
            # Edicion manual
            $cmdText = $CommandInfo.Arguments -join " "
            Write-Host ""
            Write-Color "  Argumentos actuales:" "Cyan"
            Write-Color "  $cmdText" "White"
            $newArgs = Read-ValidInput -Prompt "Nuevos argumentos (Enter para mantener)"
            if ($newArgs) {
                $CommandInfo.Arguments = $newArgs.Split(" ")
            }
            # Volver a mostrar el menu post-wizard con los argumentos actualizados
            Show-PostWizardMenu -CommandInfo $CommandInfo
        }
        default {
            Write-Color "  Operacion cancelada." "DarkGray"
        }
    }
}


# ================================================================
#                      MENU PRINCIPAL
# ================================================================

function Show-ToolSelectionMenu {
    <#
    .SYNOPSIS
        Menu principal que lista todas las herramientas encontradas en bin/,
        agrupadas por categoria, junto con opciones de gestion de perfiles.
    #>

    $executables = Find-LlamaExecutables

    if ($executables.Count -eq 0) {
        Write-Color "  [X] No se encontraron herramientas en la carpeta bin/." "Red"
        return $null
    }

    # Agrupar por categoria para orden visual
    $categories = @("Generacion", "Servidor", "Cuantizacion", "Benchmark", "Evaluacion", "Multimodal", "Audio", "Utilidad", "Debug", "Otro")
    $ordered = @()

    foreach ($cat in $categories) {
        $inCat = $executables | Where-Object { $_.Category -eq $cat }
        if ($inCat) {
            $ordered += $inCat
        }
    }

    # Construir arrays para el menu
    $optNames = @()
    $optDescs = @()
    $optColors = @()

    $lastCat = ""
    foreach ($exe in $ordered) {
        $catLabel = "             "
        if ($exe.Category -ne $lastCat) { $catLabel = "[$($exe.Category)] " }
        $lastCat = $exe.Category
        $optNames += "$catLabel$($exe.BaseName)"
        $sizeTxt = $exe.SizeMB
        $optDescs += "$($exe.Description) ($sizeTxt MB)"

        # Color por categoria
        $color = switch ($exe.Category) {
            "Generacion"   { "Green" }
            "Servidor"     { "Cyan" }
            "Cuantizacion" { "Magenta" }
            "Benchmark"    { "Yellow" }
            "Evaluacion"   { "Yellow" }
            "Multimodal"   { "Blue" }
            "Audio"        { "DarkCyan" }
            "Utilidad"     { "White" }
            "Debug"        { "DarkGray" }
            default        { "White" }
        }
        $optColors += $color
    }

    $sel = Show-Menu -Title "Selecciona una herramienta" -Options $optNames -Descriptions $optDescs -Colors $optColors -ShowBack

    if ($sel -eq -1) { return $null }
    return $ordered[$sel]
}

function Start-MainLoop {
    <#
    .SYNOPSIS
        Bucle principal del programa. Muestra el menu principal y despacha
        a los sub-menus correspondientes.
    #>

    # Validacion inicial
    if (-not (Test-Prerequisites)) {
        Write-Host ""
        Write-Color "  No se puede continuar. Corrige los errores y vuelve a ejecutar." "Red"
        Write-Host ""
        return
    }

    while ($true) {
        Show-Banner

        $mainChoice = Show-Menu -Title "MENU PRINCIPAL" -Options @(
            "[1] Ejecutar una herramienta"
            "[2] Perfiles guardados (Presets)"
            "[3] Buscar modelos .gguf"
            "[4] Historial de comandos",
            "[5] Chat rapido (Ultimo modelo)",
            "[6] Monitor de Servidor en Background",
            "[7] Analizar Documento de Texto (Lotes)",
            "[8] Descargar modelo de HuggingFace",
            "[9] Leer Metadata GGUF",
            "[10] Actualizar llama.cpp",
            "[11] Info del sistema",
            "[12] Salir"
        ) -Descriptions @(
            "Seleccionar y configurar una herramienta"
            "Ejecutar configuraciones guardadas"
            "Explorar modelos disponibles"
            "Ver comandos pasados"
            "Lanzar ultimo modelo"
            "Revisar/Apagar servidor / Web UI"
            "Resumir documentos con modelo"
            "Bajar GGUF directo"
            "Ver parametros internos"
            "Bajar ultima version Github"
            "Rutas y Hardware"
            "Cerrar"
        )

        switch ($mainChoice) {
            0 {
                # --- Ejecutar herramienta ---
                $tool = Show-ToolSelectionMenu
                if (-not $tool) { continue }

                # Despachar al wizard apropiado segun el ejecutable
                $commandInfo = $null
                switch ($tool.BaseName) {
                    "llama-cli"    { $commandInfo = Invoke-CliWizard    -ExePath $tool.Path }
                    "llama"        { $commandInfo = Invoke-CliWizard    -ExePath $tool.Path }
                    "llama-server" { $commandInfo = Invoke-ServerWizard -ExePath $tool.Path }
                    "llama-quantize" { $commandInfo = Invoke-QuantizeWizard -ExePath $tool.Path }
                    "llama-bench"    { $commandInfo = Invoke-BenchWizard  -ExePath $tool.Path }
                    "llama-batched-bench" { $commandInfo = Invoke-BenchWizard -ExePath $tool.Path }

                    "llama-llava-cli" { $commandInfo = Invoke-MultimodalWizard -ExePath $tool.Path }
                    "llama-qwen2vl-cli" { $commandInfo = Invoke-MultimodalWizard -ExePath $tool.Path }
                    "llama-minicpmv-cli" { $commandInfo = Invoke-MultimodalWizard -ExePath $tool.Path }
                    "llama-gemma3-cli" { $commandInfo = Invoke-MultimodalWizard -ExePath $tool.Path }
                    "llama-mtmd-cli" { $commandInfo = Invoke-MultimodalWizard -ExePath $tool.Path }
                    "llama-tts" { $commandInfo = Invoke-TtsWizard -ExePath $tool.Path }

                    default {
                        $commandInfo = Invoke-GenericWizard -ExePath $tool.Path -ToolName $tool.BaseName
                    }
                }

                if ($commandInfo) {
                    Show-PostWizardMenu -CommandInfo $commandInfo
                }
            }
            1 {
                # --- Perfiles ---
                Show-ProfileManager
            }
            2 {
                # --- Buscar modelos ---
                Write-Host ""
                Show-Separator "Modelos .gguf encontrados" "Cyan"
                $models = Find-GgufModels -Recursive

                if ($models.Count -gt 0) {
                    Write-Host ""
                    foreach ($m in $models) {
                        $sizeLabel = "$($m.SizeMB) MB"
                        if ($m.SizeGB -ge 1) { $sizeLabel = "$($m.SizeGB) GB" }
                        Write-ColorLine @(
                            @{ Text = "  * "; Color = "Green" },
                            @{ Text = $m.Name; Color = "White" },
                            @{ Text = " ($sizeLabel)"; Color = "DarkGray" }
                        )
                        Write-Color "    $($m.Dir)" "DarkGray"
                    }
                    Write-Host ""
                    $totalModels = $models.Count
                    Write-Color "  Total: $totalModels modelo(s) encontrado(s)" "Cyan"
                } else {
                    Write-Color "  No se encontraron modelos .gguf." "Yellow"
                    Write-Color "  Rutas de busqueda:" "DarkGray"
                    foreach ($p in $Script:ModelSearchPaths) {
                        $existsMark = "[X]"
                        if (Test-Path $p) { $existsMark = "[OK]" }
                        Write-Color "    $existsMark $p" "DarkGray"
                    }
                }

                Write-Host ""
                Write-Host "  Presiona cualquier tecla para volver..." -ForegroundColor "DarkGray"
                [Console]::ReadKey($true) | Out-Null
            }
            3 {
                # --- Historial ---
                Show-History
            }
            4 {
                # --- Chat rapido ---
                Invoke-QuickChat
            }
            5 {
                # --- Monitor Servidor ---
                Show-ServerMonitor
            }
            6 {
                # --- Document Analyzer ---
                Invoke-DocumentAnalyzer
            }
            7 {
                # --- HF Downloader ---
                Invoke-HfDownloader
            }
            8 {
                # --- Metadata Reader ---
                $modelPath = Browse-ForModel
                if ($modelPath) {
                    Show-GgufMetadata -ModelPath $modelPath
                }
            }
            9 {
                # --- Auto Updater ---
                Invoke-AutoUpdater
            }
            10 {
                # --- Info del sistema ---
                Write-Host ""
                Show-Separator "Informacion del Sistema" "Cyan"
                Write-Host ""

                $osVer = [System.Environment]::OSVersion.VersionString
                $psVer = $PSVersionTable.PSVersion.ToString()
                Write-ColorLine @(
                    @{ Text = "  Sistema Operativo:  "; Color = "DarkGray" },
                    @{ Text = $osVer; Color = "White" }
                )
                Write-ColorLine @(
                    @{ Text = "  PowerShell:         "; Color = "DarkGray" },
                    @{ Text = $psVer; Color = "White" }
                )
                Write-ColorLine @(
                    @{ Text = "  Procesadores (CPU): "; Color = "DarkGray" },
                    @{ Text = "$Script:DefaultThreads hilos logicos"; Color = "White" }
                )

                # RAM disponible
                try {
                    $os = Get-CimInstance -ClassName Win32_OperatingSystem -ErrorAction SilentlyContinue
                    if ($os) {
                        $totalRAM = [Math]::Round($os.TotalVisibleMemorySize / 1MB, 1)
                        $freeRAM  = [Math]::Round($os.FreePhysicalMemory / 1MB, 1)
                        Write-ColorLine @(
                            @{ Text = "  RAM Total:          "; Color = "DarkGray" },
                            @{ Text = "$totalRAM GB"; Color = "White" }
                        )
                        $ramColor = "Green"
                        if ($freeRAM -lt 2) { $ramColor = "Red" }
                        elseif ($freeRAM -lt 4) { $ramColor = "Yellow" }
                        Write-ColorLine @(
                            @{ Text = "  RAM Disponible:     "; Color = "DarkGray" },
                            @{ Text = "$freeRAM GB"; Color = $ramColor }
                        )
                    }
                } catch {}

                # Disco
                try {
                    $drive = (Get-Item $Script:BasePath).PSDrive
                    $freeGB = [Math]::Round(($drive.Free / 1GB), 1)
                    $diskColor = "Green"
                    if ($freeGB -lt 5) { $diskColor = "Red" }
                    elseif ($freeGB -lt 20) { $diskColor = "Yellow" }
                    Write-ColorLine @(
                        @{ Text = "  Disco Disponible:   "; Color = "DarkGray" },
                        @{ Text = "$freeGB GB"; Color = $diskColor }
                    )
                } catch {}

                Write-Host ""
                Show-Separator "Rutas Configuradas" "Cyan"
                Write-Host ""
                Write-ColorLine @(
                    @{ Text = "  Script:    "; Color = "DarkGray" },
                    @{ Text = $Script:BasePath; Color = "White" }
                )
                Write-ColorLine @(
                    @{ Text = "  Bin:       "; Color = "DarkGray" },
                    @{ Text = $Script:BinPath; Color = "White" }
                )
                Write-ColorLine @(
                    @{ Text = "  Perfiles:  "; Color = "DarkGray" },
                    @{ Text = $Script:ProfilePath; Color = "White" }
                )

                Write-Host ""
                Show-Separator "Ejecutables Detectados" "Cyan"
                Write-Host ""
                $exes = Find-LlamaExecutables
                $grouped = $exes | Group-Object Category
                foreach ($group in $grouped) {
                    Write-Color "  [$($group.Name)]" "Yellow"
                    foreach ($exe in $group.Group) {
                        $bname = $exe.BaseName
                        $bsize = $exe.SizeMB
                        Write-Color "    - $bname ($bsize MB)" "DarkGray"
                    }
                }
                Write-Host ""
                $totalExes = $exes.Count
                Write-Color "  Total: $totalExes ejecutable(s)" "Cyan"

                Write-Host ""
                Write-Host "  Presiona cualquier tecla para volver..." -ForegroundColor "DarkGray"
                [Console]::ReadKey($true) | Out-Null
            }
            11 {
                # --- Salir ---
                Write-Host ""
                Write-Color "  Hasta luego!" "Cyan"
                Write-Host ""
                return
            }
            default {
                # -1 (Escape) en menu principal = salir
                Write-Host ""
                Write-Color "  Hasta luego!" "Cyan"
                Write-Host ""
                return
            }
        }
    }
}


# ================================================================
#                       PUNTO DE ENTRADA
# ================================================================

# Ejecutar el bucle principal al lanzar el script
Start-MainLoop


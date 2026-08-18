package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"llamamanager/internal/hardware"
	"llamamanager/internal/models"
	"llamamanager/internal/process"
	"llamamanager/internal/tester"

	"github.com/charmbracelet/huh"
)

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func RunMainMenu() error {
	clearScreen()
	var action string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Gestor Universal de Llama.cpp").
				Description("Elige una opción").
				Options(
					huh.NewOption("1. Iniciar Chat en Terminal (CLI)", "cli"),
					huh.NewOption("2. Iniciar Servidor Local (Web/API)", "servidor"),
					huh.NewOption("3. Descargar Nuevo Modelo", "descargar"),
					huh.NewOption("4. Ajustes de Rendimiento", "ajustes"),
					huh.NewOption("5. Benchmark: Hardware y Modelo", "benchmark"),
					huh.NewOption("9. Salir", "salir"),
				).
				Value(&action),
		),
	)

	err := form.Run()
	if err != nil {
		return err
	}

	fmt.Printf("Has seleccionado: %s\n", action)

	if action == "salir" {
		fmt.Println("Saliendo del Gestor de Llama.cpp...")
		os.Exit(0)
	}

	if action == "cli" || action == "servidor" || action == "benchmark" {
		// 1. Escanear Modelos
		foundModels, _ := models.FindGGUFModels(".")
		var selectedModel string

		if len(foundModels) > 0 {
			opts := make([]huh.Option[string], len(foundModels))
			for i, m := range foundModels {
				opts[i] = huh.NewOption(m, m)
			}
			
			huh.NewSelect[string]().
				Title("Selecciona un modelo detectado").
				Options(opts...).
				Value(&selectedModel).Run()
		} else {
			huh.NewInput().
				Title("Ruta al archivo .gguf").
				Value(&selectedModel).Run()
			selectedModel = strings.Trim(selectedModel, "\"")
		}

		if selectedModel == "" {
			fmt.Println("No se seleccionó un modelo válido.")
			return nil
		}

		if action == "benchmark" {
			tester.RunDiagnostics(selectedModel)
		} else {
			// 2. Autoconfigurar
			fmt.Println("\nAnalizando Hardware (CPU/RAM/GPU)...")
			
			settings := LoadSettings()
			
			cfg, err := hardware.BuildOptimalConfig(selectedModel, settings.ForceCPU, settings.CustomGPU, settings.CustomContext, settings.CustomThreads)
			if err != nil {
				return fmt.Errorf("error construyendo config: %v", err)
			}

			fmt.Printf("Configuración Óptima: RAM[%s] HILOS[%s] GPU[%s]\n", cfg.CtxFlag, cfg.ThreadFlag, cfg.GPUFlag)

			// 3. Ejecutar
			if action == "cli" {
				baseArgs := []string{"-m", selectedModel, "-c", cfg.CtxFlag, "-t", cfg.ThreadFlag, "-ngl", cfg.GPUFlag, "-n", "-1", "--color", "-i"}
				// Inyectar System Prompt si no está vacío
				if settings.SystemPrompt != "" {
					baseArgs = append(baseArgs, "-p", settings.SystemPrompt)
				}
				
				exe, finalArgs := hardware.GetCLIArgs(cfg.Executable, baseArgs)
				err := process.RunLlamaInteractive(exe, finalArgs)
				if err != nil {
					fmt.Printf("\n[Error de Ejecución]: %v\n", err)
				}
			} else {
				baseArgs := []string{"-m", selectedModel, "--port", "8080", "-c", cfg.CtxFlag, "-t", cfg.ThreadFlag, "-ngl", cfg.GPUFlag}
				exe, finalArgs := hardware.GetServerArgs(cfg.Executable, baseArgs)
				err := process.RunLlamaServer(exe, finalArgs)
				if err != nil {
					fmt.Printf("\n[Error de Servidor]: %v\n", err)
				}
			}
		}

		// Pausar para que la consola no se cierre al terminar
		fmt.Println("\n\nEl proceso de Llama ha finalizado.")
		fmt.Println("Presiona ENTER para volver al menú o salir...")
		var dummy string
		fmt.Scanln(&dummy)
	} else if action == "descargar" {
		fmt.Println("\n¡La función de descarga estará disponible muy pronto!")
		fmt.Println("Presiona ENTER para volver...")
		var dummy string
		fmt.Scanln(&dummy)
	} else if action == "ajustes" {
		err := RunSettingsMenu()
		if err != nil {
			fmt.Println("Error abriendo ajustes:", err)
		}
	}

	// Reiniciar el menú recursivamente
	return RunMainMenu()
}

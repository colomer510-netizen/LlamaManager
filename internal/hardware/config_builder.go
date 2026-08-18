package hardware

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type ConfigFlags struct {
	CtxFlag    string
	ThreadFlag string
	GPUFlag    string
	Executable string
}

// FindLlamaExecutable busca el ejecutable correcto de llama.cpp (llama.exe, llama-cli.exe, etc)
func FindLlamaExecutable() (string, error) {
	// Intentamos buscar "llama", "llama-cli" en el PATH o en la carpeta local
	candidates := []string{"llama", "llama-cli"}
	
	for _, candidate := range candidates {
		path, err := exec.LookPath(candidate)
		if err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no se encontró ningún ejecutable de llama.cpp en el PATH")
}

// BuildOptimalConfig calcula los mejores flags basándose en los recursos actuales y prueba la GPU
func BuildOptimalConfig(modelPath string, forceCPU bool, customGPU int, customCtx int, customThr int) (*ConfigFlags, error) {
	specs, err := GetSystemSpecs()
	if err != nil {
		return nil, err
	}

	cfg := &ConfigFlags{}

	// 1. Configuración de RAM (Contexto)
	if customCtx > 0 {
		cfg.CtxFlag = fmt.Sprintf("%d", customCtx)
	} else {
		if specs.TotalRAMGB >= 16 {
			cfg.CtxFlag = "4096"
		} else {
			cfg.CtxFlag = "2048"
		}
	}

	// 2. Configuración de CPU (Hilos)
	if customThr > 0 {
		cfg.ThreadFlag = fmt.Sprintf("%d", customThr)
	} else {
		threads := specs.LogicalCores
		if threads > 2 {
			threads -= 1 // Reservar 1 núcleo para el SO
		}
		cfg.ThreadFlag = fmt.Sprintf("%d", threads)
	}

	// 3. Encontrar Ejecutable
	exe, err := FindLlamaExecutable()
	if err != nil {
		return nil, err
	}
	cfg.Executable = exe

	// 4. Prueba de Fuego GPU o Valores Custom
	if forceCPU {
		cfg.GPUFlag = "0"
	} else if customGPU >= 0 {
		cfg.GPUFlag = fmt.Sprintf("%d", customGPU)
	} else if modelPath != "" {
		if testGPU(exe, modelPath) {
			cfg.GPUFlag = "99"
		} else {
			cfg.GPUFlag = "0"
		}
	} else {
		cfg.GPUFlag = "99" // Predeterminado si no hay modelo para probar
	}

	return cfg, nil
}

func testGPU(executable string, modelPath string) bool {
	// Si el binario es el nuevo unificado ("llama.exe"), necesita el comando "cli"
	args := []string{}
	base := filepath.Base(strings.ToLower(executable))
	if base == "llama.exe" || base == "llama" {
		args = append(args, "cli")
	}
	args = append(args, "-m", modelPath, "-p", "test", "-n", "1", "-ngl", "99")

	// Ejecuta un test ultra rápido para verificar si Vulkan/CUDA fallan
	cmd := exec.Command(executable, args...)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func GetServerArgs(executable string, baseArgs []string) (string, []string) {
	base := filepath.Base(strings.ToLower(executable))
	if base == "llama.exe" || base == "llama" {
		// Nuevo formato: llama serve <args>
		return executable, append([]string{"serve"}, baseArgs...)
	} else if strings.Contains(base, "llama-cli") {
		// Viejo formato: usar llama-server en su lugar
		serverExe := strings.Replace(executable, "llama-cli", "llama-server", 1)
		return serverExe, baseArgs
	}
	return executable, baseArgs
}

func GetCLIArgs(executable string, baseArgs []string) (string, []string) {
	base := filepath.Base(strings.ToLower(executable))
	if base == "llama.exe" || base == "llama" {
		// Nuevo formato: llama cli <args>
		return executable, append([]string{"cli"}, baseArgs...)
	}
	return executable, baseArgs
}

package tester

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"llamamanager/internal/hardware"
)

// RunDiagnostics ejecuta un test progresivo en CPU y GPU usando llama-bench
func RunDiagnostics(modelPath string) {
	fmt.Println("\n=====================================================")
	fmt.Println("       INICIANDO TEST DE ESTRÉS (BENCHMARK)          ")
	fmt.Println("=====================================================")
	
	exe, err := hardware.FindLlamaExecutable()
	if err != nil {
		fmt.Printf("[Error] No se encontró llama.cpp en tu PC.\n")
		return
	}

	// Averiguar si debemos usar "llama bench" o "llama-bench"
	benchCmd := []string{}
	if strings.HasSuffix(strings.ToLower(exe), "llama.exe") || strings.HasSuffix(strings.ToLower(exe), "llama") {
		benchCmd = append(benchCmd, "bench")
	} else {
		exe = strings.Replace(exe, "llama-cli", "llama-bench", 1) // Fallback asumiendo binaries antiguos
	}

	// 1. TEST DE SUBSISTENCIA CPU
	fmt.Println("\n[1/2] Probando estabilidad en Procesador (CPU)...")
	argsCPU := append(benchCmd, "-m", modelPath, "-p", "128", "-n", "16", "-ngl", "0")
	cmdCPU := exec.Command(exe, argsCPU...)
	
	start := time.Now()
	errCPU := cmdCPU.Run()
	elapsedCPU := time.Since(start)

	if errCPU != nil {
		fmt.Printf("❌ ERROR FATAL: El modelo está corrupto o es incompatible. (%v)\n", errCPU)
		return
	}
	fmt.Printf("✅ CPU: ¡Aprobado! (Tiempo: %v)\n", elapsedCPU)


	// 2. TEST DE CARGA VRAM Y VULKAN (GPU)
	fmt.Println("\n[2/2] Probando Tarjeta de Video y Drivers (GPU)...")
	argsGPU := append(benchCmd, "-m", modelPath, "-p", "128", "-n", "16", "-ngl", "99")
	cmdGPU := exec.Command(exe, argsGPU...)
	
	start = time.Now()
	errGPU := cmdGPU.Run()
	elapsedGPU := time.Since(start)

	if errGPU != nil {
		fmt.Printf("❌ GPU FALLÓ: %v\n", errGPU)
		fmt.Println("    -> Diagnóstico: Tus drivers de video (Vulkan) se desbordaron.")
		fmt.Println("    -> Solución Recomendada: Ve al Menú '4. Ajustes', y activa 'Forzar Modo CPU'.")
	} else {
		fmt.Printf("✅ GPU: ¡Aprobada! (Tiempo: %v)\n", elapsedGPU)
		
		if elapsedGPU < elapsedCPU {
			fmt.Println("\n💡 CONCLUSIÓN: Tu Tarjeta de Video es MUCHO más rápida. Puedes usar configuraciones normales (-ngl 99).")
		} else {
			fmt.Println("\n💡 CONCLUSIÓN: Curiosamente tu CPU es más rápida o igual que tu GPU.")
		}
	}
}

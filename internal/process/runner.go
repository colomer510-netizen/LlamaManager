package process

import (
	"fmt"
	"os"
	"os/exec"
)

// RunLlamaInteractive lanza el proceso CLI enganchándolo a la consola actual
func RunLlamaInteractive(executable string, args []string) error {
	cmd := exec.Command(executable, args...)
	
	// Conectar entradas y salidas a la terminal actual
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("\n--- Iniciando Llama CLI ---\n")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar llama: %v", err)
	}
	
	return nil
}

// RunLlamaServer lanza el proceso servidor (puede ser en un modo background o foreground)
func RunLlamaServer(executable string, args []string) error {
	cmd := exec.Command(executable, args...)
	
	// Para el servidor, simplemente mostramos el stdout/stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("\n--- Iniciando Llama Server ---\n")
	err := cmd.Run() // Se queda bloqueado hasta que el usuario presiona Ctrl+C
	if err != nil {
		return fmt.Errorf("el servidor se detuvo: %v", err)
	}

	return nil
}

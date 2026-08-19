package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ResolveBinPath busca el binario en posibles ubicaciones relativas.
func ResolveBinPath(exeName string) (string, error) {
	paths := []string{
		filepath.Join("bin", exeName),
		filepath.Join("..", "..", "bin", exeName),
		filepath.Join("build", "bin", "bin", exeName), // Para casos de dev anidados
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs, nil
		}
	}
	return "", fmt.Errorf("no se encontró %s en las rutas esperadas", exeName)
}

// RunInteractive abre una nueva ventana de consola CMD para ejecutar la herramienta
// y la mantiene abierta (/K) para que el usuario pueda ver el resultado.
func RunInteractive(exePath string, args []string) error {
	// Construimos el comando completo escapando la ruta del ejecutable
	cmdParts := []string{fmt.Sprintf(`"%s"`, exePath)}
	for _, arg := range args {
		// Envolver argumentos con espacios en comillas
		if strings.Contains(arg, " ") {
			cmdParts = append(cmdParts, fmt.Sprintf(`"%s"`, arg))
		} else {
			cmdParts = append(cmdParts, arg)
		}
	}
	
	fullCommand := strings.Join(cmdParts, " ")
	
	// Comando de Windows para abrir una nueva ventana y ejecutar el fullCommand
	// /c: corre el start. 
	// start "Titulo" cmd.exe /K
	c := exec.Command("cmd.exe", "/c", "start", "LlamaManager Tool", "cmd.exe", "/K", fullCommand)
	
	err := c.Start()
	if err != nil {
		return fmt.Errorf("error iniciando comando interactivo: %w", err)
	}
	return nil
}

package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ResolveBinPath busca el binario en posibles ubicaciones relativas o en el PATH del sistema.
func ResolveBinPath(exeName string) (string, error) {
	nameNoExt := strings.TrimSuffix(exeName, ".exe")
	variations := []string{exeName}
	if nameNoExt != exeName {
		variations = append(variations, nameNoExt)
	} else {
		variations = append(variations, exeName+".exe")
	}

	searchDirs := []string{
		"bin",
		filepath.Join("..", "..", "bin"),
		filepath.Join("build", "bin", "bin"),
		".",
	}

	for _, dir := range searchDirs {
		for _, name := range variations {
			p := filepath.Join(dir, name)
			if _, err := os.Stat(p); err == nil {
				abs, _ := filepath.Abs(p)
				return abs, nil
			}
		}
		// Buscar en subdirectorios (por ej: bin/llama-bXXXX/llama-cli)
		if entries, err := os.ReadDir(dir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					for _, name := range variations {
						subP := filepath.Join(dir, entry.Name(), name)
						if _, err := os.Stat(subP); err == nil {
							abs, _ := filepath.Abs(subP)
							return abs, nil
						}
					}
				}
			}
		}
	}

	// Buscar en PATH del sistema
	for _, name := range variations {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no se encontró %s en las rutas esperadas ni en el PATH", exeName)
}

// RunInteractive abre una nueva ventana de consola o terminal para ejecutar la herramienta
// y la mantiene abierta para que el usuario pueda ver el resultado.
func RunInteractive(exePath string, args []string) error {
	if runtime.GOOS == "windows" {
		cmdParts := []string{fmt.Sprintf(`"%s"`, exePath)}
		for _, arg := range args {
			if strings.Contains(arg, " ") {
				cmdParts = append(cmdParts, fmt.Sprintf(`"%s"`, arg))
			} else {
				cmdParts = append(cmdParts, arg)
			}
		}
		fullCommand := strings.Join(cmdParts, " ")
		c := exec.Command("cmd.exe", "/c", "start", "LlamaManager Tool", "cmd.exe", "/K", fullCommand)
		return c.Start()
	}

	// En Linux: intentar terminales comunes
	fullArgs := append([]string{exePath}, args...)
	terminals := [][]string{
		{"gnome-terminal", "--"},
		{"x-terminal-emulator", "-e"},
		{"konsole", "-e"},
		{"xfce4-terminal", "-e"},
		{"xterm", "-e"},
	}

	for _, term := range terminals {
		if _, err := exec.LookPath(term[0]); err == nil {
			cmdArgs := append(term[1:], fullArgs...)
			c := exec.Command(term[0], cmdArgs...)
			if err := c.Start(); err == nil {
				return nil
			}
		}
	}
	c := exec.Command(exePath, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Start()
}

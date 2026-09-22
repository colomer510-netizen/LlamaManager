package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	pidFile      = "/tmp/gestor_llama_server.pid"
	logFile      = "/tmp/gestor_llama_server.log"
	settingsFile = "llama_settings.json"
)

// Configuración Persistente
type AppSettings struct {
	ModelPath    string `json:"model_path"`
	APIPort      string `json:"api_port"`
	ContextSize  int    `json:"context_size"`
	GPULayers    int    `json:"gpu_layers"`
	ForceCPU     bool   `json:"force_cpu"`
	RPCServers   string `json:"rpc_servers"`
	IdleTimerMin int    `json:"idle_timer_min"`
}

func defaultSettings() AppSettings {
	home, _ := os.UserHomeDir()
	return AppSettings{
		ModelPath:    filepath.Join(home, "Documentos"),
		APIPort:      "8080",
		ContextSize:  4096,
		GPULayers:    99,
		ForceCPU:     false,
		RPCServers:   "",
		IdleTimerMin: 0,
	}
}

func loadSettings() AppSettings {
	s := defaultSettings()
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return s
	}
	json.Unmarshal(data, &s)
	return s
}

func saveSettings(s AppSettings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsFile, data, 0644)
}

func readInput(prompt string, defaultVal string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

func getTotalRAMGB() int {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 8
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseUint(fields[1], 10, 64)
				if err == nil {
					return int(kb / (1024 * 1024))
				}
			}
		}
	}
	return 8
}

func getOptimalContext(ramGB int) int {
	if ramGB >= 32 {
		return 16384
	}
	if ramGB >= 16 {
		return 8192
	}
	return 4096
}

func getRunningPID() int {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return 0
	}
	err = process.Signal(syscall.Signal(0))
	if err == nil {
		return pid
	}
	return 0
}

func checkServerHealth(port string) string {
	client := http.Client{Timeout: 800 * time.Millisecond}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%s/health", port))
	if err == nil && (resp.StatusCode == 200 || resp.StatusCode == 503) {
		return "🟢 [API Activa y Saludable]"
	}
	return "🟡 [Iniciando o cargando modelo...]"
}

func scanModels(rootDir string) []string {
	var models []string
	baseDepth := strings.Count(filepath.Clean(rootDir), string(os.PathSeparator))
	maxDepth := 4

	filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "bin" {
				if path != rootDir {
					return filepath.SkipDir
				}
			}
			depth := strings.Count(filepath.Clean(path), string(os.PathSeparator)) - baseDepth
			if depth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(strings.ToLower(d.Name()), ".gguf") {
			models = append(models, path)
		}
		return nil
	})

	sort.Strings(models)
	return models
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// -------------------------------------------------------------
// TELEMETRÍA Y AGENTE AUTÓNOMO
// -------------------------------------------------------------

func getSystemTelemetry() string {
	var sb strings.Builder
	sb.WriteString("=== TELEMETRÍA EN VIVO DEL SISTEMA LINUX ===\n")

	if out, err := exec.Command("uname", "-snrvm").Output(); err == nil {
		sb.WriteString("SO / Kernel: " + strings.TrimSpace(string(out)) + "\n")
	}
	if out, err := exec.Command("uptime", "-p").Output(); err == nil {
		sb.WriteString("Tiempo Activo: " + strings.TrimSpace(string(out)) + "\n")
	}
	if out, err := exec.Command("free", "-h").Output(); err == nil {
		sb.WriteString("Memoria RAM:\n" + strings.TrimSpace(string(out)) + "\n")
	}
	if out, err := exec.Command("df", "-h", "/").Output(); err == nil {
		sb.WriteString("Espacio en Disco Principal (/):\n" + strings.TrimSpace(string(out)) + "\n")
	}
	if out, err := exec.Command("nproc").Output(); err == nil {
		sb.WriteString("Núcleos CPU: " + strings.TrimSpace(string(out)) + "\n")
	}
	sb.WriteString("============================================")
	return sb.String()
}

func isDangerousCommand(cmd string) bool {
	dangerousPatterns := []string{
		"rm -rf /", "rm -rf /*", "mkfs", "dd if=", ":(){ :|:& };:",
		"> /dev/sda", "> /dev/nvme", "chmod -R 777 /", "shutdown", "reboot",
	}
	lower := strings.ToLower(cmd)
	for _, p := range dangerousPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func extractBashCommand(response string) string {
	re := regexp.MustCompile("(?s)```(?:bash|sh)?\\s*\\n?(.*?)\\n?```")
	matches := re.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	Stream      bool          `json:"stream"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

func autoStartServer(cfg AppSettings, llamaCmd string) (int, error) {
	homeDir, _ := os.UserHomeDir()
	models := scanModels(cfg.ModelPath)
	if len(models) == 0 {
		models = scanModels(filepath.Join(homeDir, "Descargas"))
	}
	if len(models) == 0 {
		return 0, fmt.Errorf("no se encontraron modelos .gguf en Documentos ni Descargas")
	}
	selectedModel := models[0]
	ctxSize := cfg.ContextSize
	gpuLayers := cfg.GPULayers
	if cfg.ForceCPU {
		gpuLayers = 0
	}

	args := []string{"serve", "-m", selectedModel, "-c", fmt.Sprintf("%d", ctxSize), "-ngl", fmt.Sprintf("%d", gpuLayers), "-fa", "on", "--host", "127.0.0.1", "--port", cfg.APIPort}
	if cfg.RPCServers != "" {
		args = append(args, "--rpc", cfg.RPCServers)
	}

	cmd := exec.Command(llamaCmd, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = os.Environ()
	if gpuLayers == 0 {
		cmd.Env = append(cmd.Env, "GGML_VK_VISIBLE_DEVICES=", "CUDA_VISIBLE_DEVICES=")
	}

	outFile, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		cmd.Stdout = outFile
		cmd.Stderr = outFile
	}

	if err := cmd.Start(); err != nil {
		return 0, err
	}

	os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644)
	return cmd.Process.Pid, nil
}

func streamLLMResponse(apiURL string, history []ChatMessage, temp float64) (string, error) {
	reqBody, err := json.Marshal(ChatRequest{
		Messages:    history,
		Temperature: temp,
		Stream:      true,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var fullText strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			payload := strings.TrimPrefix(line, "data: ")
			if strings.TrimSpace(payload) == "[DONE]" {
				break
			}
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(payload), &chunk); err == nil {
				if len(chunk.Choices) > 0 {
					token := chunk.Choices[0].Delta.Content
					fmt.Print(token)
					fullText.WriteString(token)
				}
			}
		}
	}
	fmt.Println()
	return fullText.String(), scanner.Err()
}

func runSuperAgent(cfg AppSettings, llamaCmd string) {
	fmt.Println("\n=======================================================")
	fmt.Println("   🤖 INICIANDO SUPER AGENTE AUTÓNOMO LINUX 🤖         ")
	fmt.Println("=======================================================")

	pid := getRunningPID()
	startedByAgent := false
	if pid == 0 {
		fmt.Println("⏳ El motor API no está activo. Iniciándolo en segundo plano...")
		newPid, err := autoStartServer(cfg, llamaCmd)
		if err != nil {
			fmt.Printf("❌ Error al arrancar el motor: %v\n", err)
			return
		}
		startedByAgent = true
		fmt.Printf("👻 Motor iniciado (PID: %d). Cargando modelo en memoria", newPid)
		for i := 0; i < 20; i++ {
			time.Sleep(1 * time.Second)
			if checkServerHealth(cfg.APIPort) == "🟢 [API Activa y Saludable]" {
				break
			}
			fmt.Print(".")
		}
		fmt.Println("\n✅ ¡Motor en línea y listo para recibir preguntas!")
	}

	apiURL := fmt.Sprintf("http://127.0.0.1:%s/v1/chat/completions", cfg.APIPort)
	telemetry := getSystemTelemetry()

	systemPrompt := fmt.Sprintf(`Eres un Super Agente de Diagnóstico y Asistencia de Linux para este computador.
Tienes acceso a la telemetría en tiempo real:
%s

Instrucciones:
1. Responde de forma clara y profesional en español.
2. Puedes diagnosticar problemas de rendimiento, memoria, red y espacio en disco.
3. Si para responder o resolver una tarea necesitas ejecutar un comando en Linux, escribe el comando exacto encerrado en un bloque de código:
`+"```bash"+`
tu_comando_aqui
`+"```"+`
Solo incluye un comando por turno y explica brevemente al usuario qué hace antes del bloque.`, telemetry)

	history := []ChatMessage{
		{Role: "system", Content: systemPrompt},
	}

	fmt.Println("\n✅ Super Agente en línea con Streaming en vivo.")
	fmt.Println("👉 Puedes hacerle preguntas sobre tu PC, pedirle diagnósticos o solicitar tareas.")
	fmt.Println("👉 Escribe 'salir' para volver al menú principal.\n")

	for {
		userMsg := readInput("\n👤 Tú: ", "")
		if userMsg == "" {
			continue
		}
		if strings.ToLower(userMsg) == "salir" {
			fmt.Println("Saliendo de la sesión del Super Agente...")
			break
		}

		history = append(history, ChatMessage{Role: "user", Content: userMsg})
		fmt.Print("\n🤖 Agente: ")

		agentReply, err := streamLLMResponse(apiURL, history, 0.7)
		if err != nil {
			fmt.Printf("\n❌ Error de conexión con el motor: %v\n", err)
			break
		}

		history = append(history, ChatMessage{Role: "assistant", Content: agentReply})

		// Detección y ejecución de comandos propuesta por el Agente
		cmdToRun := extractBashCommand(agentReply)
		if cmdToRun != "" {
			fmt.Println("\n-------------------------------------------------------")
			fmt.Printf("⚙️  COMANDO SUGERIDO POR EL AGENTE:\n> %s\n", cmdToRun)
			
			if isDangerousCommand(cmdToRun) {
				fmt.Println("🛡️ [ALERTA DE SEGURIDAD]: El comando parece destructivo y fue bloqueado automáticamente.")
				continue
			}

			execConfirm := readInput("¿Deseas autorizar la ejecución de este comando? (s/N): ", "n")
			if strings.ToLower(execConfirm) == "s" {
				fmt.Println("⏳ Ejecutando en tu terminal...")
				execCmd := exec.Command("bash", "-c", cmdToRun)
				outBytes, execErr := execCmd.CombinedOutput()
				output := string(outBytes)

				if len(output) > 4096 {
					output = output[:4096] + "\n... (salida recortada)"
				}

				fmt.Printf("\n--- Salida del Comando ---\n%s\n--------------------------\n", output)

				// Realimentar al agente con la salida
				feedbackMsg := fmt.Sprintf("El comando '%s' se ejecutó.\nSalida obtenida:\n%s", cmdToRun, output)
				if execErr != nil {
					feedbackMsg = fmt.Sprintf("El comando falló con error: %v\nSalida:\n%s", execErr, output)
				}
				history = append(history, ChatMessage{Role: "system", Content: feedbackMsg})

				// El agente analiza la salida en vivo con streaming
				fmt.Print("\n🤖 Análisis del Resultado: ")
				finalAnalysis, err2 := streamLLMResponse(apiURL, history, 0.5)
				if err2 == nil {
					history = append(history, ChatMessage{Role: "assistant", Content: finalAnalysis})
				}
			} else {
				fmt.Println("ℹ️  Comando omitido por el usuario.")
			}
		}
	}

	if startedByAgent {
		shutdown := readInput("\n¿Deseas apagar el motor en segundo plano para liberar RAM? (S/n): ", "s")
		if strings.ToLower(shutdown) != "n" {
			currentPid := getRunningPID()
			if currentPid > 0 {
				killServer(currentPid)
			}
		}
	}
}

// -------------------------------------------------------------
// MENÚ PRINCIPAL
// -------------------------------------------------------------

func main() {
	homeDir, _ := os.UserHomeDir()
	llamaCmd := filepath.Join(homeDir, ".local", "bin", "llama")
	totalRAM := getTotalRAMGB()
	recCtx := getOptimalContext(totalRAM)

	for {
		cfg := loadSettings()
		pid := getRunningPID()

		fmt.Println("\n=======================================================")
		fmt.Println("  🦙 GESTOR PROFESIONAL LLAMA.CPP (V6.0 Super Agente) 🦙")
		fmt.Println("=======================================================")
		fmt.Printf(" 🖥️  Hardware: %d GB RAM | Contexto Óptimo: %d\n", totalRAM, recCtx)
		fmt.Println("-------------------------------------------------------")
		fmt.Println(" [1] 🌐 Iniciar Servidor API (Normal o Silencioso)")
		fmt.Println(" [2] 💬 Chatear en la Terminal (Modo CLI Estándar)")
		fmt.Println(" [3] 🤖 Iniciar Sesión con SUPER AGENTE Copiloto Linux")
		fmt.Println(" [4] ⚡ Test de Estrés y Benchmark (CPU vs GPU)")
		fmt.Println(" [5] 📥 Descargar nuevo modelo desde HuggingFace")
		fmt.Println(" [6] ⚙️  Configuración y Preferencias Guardadas")
		fmt.Println(" [7] 🔄 Actualizar motor oficial llama.cpp")
		fmt.Println(" [8] 🖥️  Crear Acceso Directo de Escritorio (.desktop y .sh)")
		
		if pid > 0 {
			health := checkServerHealth(cfg.APIPort)
			fmt.Println("-------------------------------------------------------")
			fmt.Printf(" 🟢 SERVIDOR ACTIVO (PID: %d) en :%s %s\n", pid, cfg.APIPort, health)
			fmt.Println(" [9] 📜 Ver consola y logs en tiempo real")
			fmt.Println(" [10] 🛑 APAGAR Servidor en Segundo Plano")
		}
		
		fmt.Println(" [0] Salir del gestor")
		fmt.Println("=======================================================")

		choice := readInput("Selecciona una opción: ", "")

		switch choice {
		case "1":
			if pid > 0 {
				fmt.Println("\n[!] Ya hay un servidor corriendo. Apágalo primero (Opción 10).")
				continue
			}
			runModel("1", cfg, llamaCmd, recCtx)
		case "2":
			runModel("2", cfg, llamaCmd, recCtx)
		case "3":
			runSuperAgent(cfg, llamaCmd)
		case "4":
			runDiagnostics(cfg, llamaCmd)
		case "5":
			downloadModel(cfg.ModelPath, llamaCmd)
		case "6":
			manageSettings(&cfg)
		case "7":
			updateLlama()
		case "8":
			setupLaunchers(homeDir)
		case "9":
			if pid > 0 {
				viewLogs()
			} else {
				fmt.Println("Opción no válida.")
			}
		case "10":
			if pid > 0 {
				killServer(pid)
			} else {
				fmt.Println("Opción no válida.")
			}
		case "0":
			fmt.Println("¡Hasta pronto! Gracias por usar Gestor Llama.")
			return
		default:
			fmt.Println("Opción inválida. Intenta nuevamente.")
		}
	}
}

// -------------------------------------------------------------
// EJECUCIÓN DEL MODELO
// -------------------------------------------------------------

func runModel(mode string, cfg AppSettings, llamaCmd string, recCtx int) {
	homeDir, _ := os.UserHomeDir()
	models := scanModels(cfg.ModelPath)
	descargasModels := scanModels(filepath.Join(homeDir, "Descargas"))

	for _, m := range descargasModels {
		exists := false
		for _, em := range models {
			if em == m { exists = true; break }
		}
		if !exists { models = append(models, m) }
	}

	if len(models) == 0 {
		fmt.Println("\n[!] No se encontraron archivos .gguf.")
		return
	}

	fmt.Println("\n--- Modelos Disponibles ---")
	for i, m := range models {
		info, err := os.Stat(m)
		sizeStr := "N/A"
		if err == nil {
			sizeStr = formatFileSize(info.Size())
		}
		fmt.Printf(" [%d] %s  (%s)\n", i+1, filepath.Base(m), sizeStr)
	}

	selStr := readInput(fmt.Sprintf("\nElige el modelo (1-%d): ", len(models)), "1")
	selIdx, err := strconv.Atoi(selStr)
	if err != nil || selIdx < 1 || selIdx > len(models) {
		fmt.Println("Selección cancelada.")
		return
	}
	selectedModel := models[selIdx-1]

	ctxPrompt := fmt.Sprintf("Tamaño de contexto [Guardado: %d | Óptimo: %d]: ", cfg.ContextSize, recCtx)
	ctxStr := readInput(ctxPrompt, fmt.Sprintf("%d", cfg.ContextSize))
	ctxSize, _ := strconv.Atoi(ctxStr)
	if ctxSize == 0 { ctxSize = cfg.ContextSize }

	gpuLayers := cfg.GPULayers
	if !cfg.ForceCPU {
		gpuPrompt := fmt.Sprintf("Capas a GPU (0 para Solo CPU) [Guardado: %d]: ", cfg.GPULayers)
		gpuStr := readInput(gpuPrompt, fmt.Sprintf("%d", cfg.GPULayers))
		gpuLayers, _ = strconv.Atoi(gpuStr)
	} else {
		fmt.Println("ℹ️  Modo 'Forzar CPU' activo. Capas GPU fijadas en 0.")
		gpuLayers = 0
	}

	args := []string{}
	var cmd *exec.Cmd
	silentMode := false

	if mode == "1" {
		portPrompt := fmt.Sprintf("Puerto de escucha [Guardado: %s]: ", cfg.APIPort)
		port := readInput(portPrompt, cfg.APIPort)

		hostInput := readInput("¿Permitir conexiones en red Wi-Fi? (s/N): ", "n")
		host := "127.0.0.1"
		if strings.ToLower(hostInput) == "s" { host = "0.0.0.0" }

		apiKey := readInput("API Key (opcional, Enter para omitir): ", "")
		silentInput := readInput("¿Iniciar de forma SILENCIOSA en segundo plano? (s/N): ", "n")
		if strings.ToLower(silentInput) == "s" { silentMode = true }

		args = append(args, "serve", "-m", selectedModel, "-c", fmt.Sprintf("%d", ctxSize), "-ngl", fmt.Sprintf("%d", gpuLayers), "-fa", "on", "--host", host, "--port", port)
		if apiKey != "" { args = append(args, "--api-key", apiKey) }
		if cfg.RPCServers != "" { args = append(args, "--rpc", cfg.RPCServers) }

		cmd = exec.Command(llamaCmd, args...)
		fmt.Printf("\n🚀 Servidor API en http://localhost:%s ...\n", port)
	} else {
		args = append(args, "cli", "-m", selectedModel, "-c", fmt.Sprintf("%d", ctxSize), "-ngl", fmt.Sprintf("%d", gpuLayers), "-fa", "on", "--color", "on")
		if cfg.RPCServers != "" { args = append(args, "--rpc", cfg.RPCServers) }
		cmd = exec.Command(llamaCmd, args...)
		fmt.Println("\n💬 Sesión CLI iniciada. (Presiona Ctrl+C para salir).")
	}

	cmd.Env = os.Environ()
	if gpuLayers == 0 {
		fmt.Println("🛡️  Inyectando protección: Desactivando Vulkan y CUDA para CPU pura.")
		cmd.Env = append(cmd.Env, "GGML_VK_VISIBLE_DEVICES=", "CUDA_VISIBLE_DEVICES=")
	}

	if silentMode {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		outFile, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err == nil {
			cmd.Stdout = outFile
			cmd.Stderr = outFile
		}
		if err := cmd.Start(); err != nil {
			fmt.Println("\n[!] Error al iniciar en segundo plano:", err)
			return
		}
		os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644)
		fmt.Printf("👻 Servidor corriendo con PID: %d. Logs en %s\n", cmd.Process.Pid, logFile)
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
	}
}

// -------------------------------------------------------------
// BENCHMARK
// -------------------------------------------------------------

func runDiagnostics(cfg AppSettings, llamaCmd string) {
	fmt.Println("\n=====================================================")
	fmt.Println("   ⚡ INICIANDO TEST DE ESTRÉS Y RENDIMIENTO ⚡    ")
	fmt.Println("=====================================================")

	models := scanModels(cfg.ModelPath)
	if len(models) == 0 {
		fmt.Println("[!] No se encontraron modelos para el test.")
		return
	}
	model := models[0]
	fmt.Printf("Probando con: %s\n", filepath.Base(model))

	fmt.Println("\n[1/2] Probando estabilidad en Procesador (CPU)...")
	cmdCPU := exec.Command(llamaCmd, "bench", "-m", model, "-p", "128", "-n", "16", "-ngl", "0")
	start := time.Now()
	errCPU := cmdCPU.Run()
	elapsedCPU := time.Since(start)

	if errCPU != nil {
		fmt.Println("❌ Fallo en CPU:", errCPU)
		return
	}
	fmt.Printf("✅ CPU Aprobado. Tiempo: %v\n", elapsedCPU)

	fmt.Println("\n[2/2] Probando Tarjeta Gráfica y VRAM (GPU)...")
	cmdGPU := exec.Command(llamaCmd, "bench", "-m", model, "-p", "128", "-n", "16", "-ngl", "99")
	start = time.Now()
	errGPU := cmdGPU.Run()
	elapsedGPU := time.Since(start)

	if errGPU != nil {
		fmt.Printf("❌ GPU FALLÓ: %v\n", errGPU)
		fmt.Println("   -> Drivers Vulkan sobrecargados o VRAM insuficiente.")
		fmt.Println("   -> Sugerencia: En Ajustes (Opción 6), activa 'Forzar Modo CPU'.")
	} else {
		fmt.Printf("✅ GPU Aprobada. Tiempo: %v\n", elapsedGPU)
		if elapsedGPU < elapsedCPU {
			fmt.Println("\n💡 CONCLUSIÓN: Tu GPU es más rápida. Usa -ngl 99.")
		} else {
			fmt.Println("\n💡 CONCLUSIÓN: Tu CPU rinde igual o mejor.")
		}
	}
}

// -------------------------------------------------------------
// AJUSTES
// -------------------------------------------------------------

func manageSettings(cfg *AppSettings) {
	for {
		fmt.Println("\n-----------------------------------------------------")
		fmt.Println("       ⚙️  CONFIGURACIÓN Y PREFERENCIAS GUARDADAS     ")
		fmt.Println("-----------------------------------------------------")
		fmt.Printf(" [1] Carpeta de Modelos: %s\n", cfg.ModelPath)
		fmt.Printf(" [2] Puerto API: %s\n", cfg.APIPort)
		fmt.Printf(" [3] Contexto Predeterminado: %d\n", cfg.ContextSize)
		fmt.Printf(" [4] Capas GPU: %d\n", cfg.GPULayers)
		fmt.Printf(" [5] Forzar Solo CPU: %v\n", cfg.ForceCPU)
		fmt.Printf(" [6] Servidores RPC: %s\n", cfg.RPCServers)
		fmt.Println(" [7] 💾 Guardar cambios y volver")
		fmt.Println(" [0] Cancelar sin guardar")
		fmt.Println("-----------------------------------------------------")

		opt := readInput("Opción a editar: ", "7")
		switch opt {
		case "1":
			cfg.ModelPath = readInput("Nueva ruta: ", cfg.ModelPath)
		case "2":
			cfg.APIPort = readInput("Nuevo puerto: ", cfg.APIPort)
		case "3":
			val, _ := strconv.Atoi(readInput("Nuevo contexto: ", fmt.Sprintf("%d", cfg.ContextSize)))
			if val > 0 { cfg.ContextSize = val }
		case "4":
			val, _ := strconv.Atoi(readInput("Capas GPU: ", fmt.Sprintf("%d", cfg.GPULayers)))
			cfg.GPULayers = val
		case "5":
			cfg.ForceCPU = !cfg.ForceCPU
		case "6":
			cfg.RPCServers = readInput("Servidores RPC: ", cfg.RPCServers)
		case "7":
			saveSettings(*cfg)
			fmt.Println("✅ Preferencias guardadas.")
			return
		case "0":
			return
		}
	}
}

// -------------------------------------------------------------
// UTILIDADES Y LANZADORES
// -------------------------------------------------------------

func downloadModel(searchDir, llamaCmd string) {
	fmt.Println("\n--- Descargar Modelo ---")
	repo := readInput("Repositorio HuggingFace: ", "")
	if repo == "" { return }
	file := readInput("Nombre de archivo (.gguf): ", "")
	if file == "" { return }

	destDir := filepath.Join(searchDir, "MODELOS IA")
	os.MkdirAll(destDir, 0755)

	cmd := exec.Command(llamaCmd, "download", "--repo", repo, "--file", file, "--outdir", destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Println("\n⬇️ Descargando...")
	if err := cmd.Run(); err == nil {
		fmt.Println("✅ Descarga completada en:", destDir)
	}
}

func updateLlama() {
	fmt.Println("\n🔄 Actualizando motor llama.cpp...")
	cmd := exec.Command("sh", "-c", "curl -Lsf https://llama.app/install.sh | sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err == nil {
		fmt.Println("✅ Actualización completada.")
	}
}

func viewLogs() {
	fmt.Println("\n--- Logs en Vivo (Ctrl+C para volver) ---")
	cmd := exec.Command("tail", "-n", "40", "-f", logFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func setupLaunchers(homeDir string) {
	execDir := filepath.Join(homeDir, "Documentos", "Motores IA Local")
	execPath := filepath.Join(execDir, "gestor-llama")
	shPath := filepath.Join(execDir, "iniciar.sh")
	desktopAppPath := filepath.Join(homeDir, ".local", "share", "applications", "gestor-llama.desktop")
	desktopEscritorioPath := filepath.Join(homeDir, "Escritorio", "Gestor-Llama.desktop")

	// 1. Crear iniciar.sh con auto-detección de terminal
	shContent := fmt.Sprintf(`#!/bin/bash
cd "%s"
if command -v gnome-terminal &>/dev/null; then
    gnome-terminal --title="Gestor Llama.cpp" -- bash -c './gestor-llama; exec bash'
elif command -v xterm &>/dev/null; then
    xterm -title "Gestor Llama.cpp" -e './gestor-llama'
else
    ./gestor-llama
fi
`, execDir)
	os.WriteFile(shPath, []byte(shContent), 0755)

	// 2. Crear archivo .desktop
	desktopContent := fmt.Sprintf(`[Desktop Entry]
Version=1.0
Type=Application
Name=Gestor Llama.cpp Super Agente
Comment=Lanzador profesional de IA local y Super Agente Linux
Exec=gnome-terminal --title="Gestor Llama.cpp" -- bash -c '"%s"; exec bash'
Icon=utilities-terminal
Terminal=false
Categories=Development;Utility;
`, execPath)

	os.WriteFile(desktopAppPath, []byte(desktopContent), 0755)
	os.WriteFile(desktopEscritorioPath, []byte(desktopContent), 0755)
	exec.Command("gio", "set", desktopEscritorioPath, "metadata::trusted", "true").Run()

	fmt.Println("\n✅ ¡Lanzadores creados con éxito!")
	fmt.Printf(" 📁 Script ejecutable directo: %s\n", shPath)
	fmt.Printf(" 🖥️  Acceso directo en el Escritorio: %s\n", desktopEscritorioPath)
	fmt.Printf(" 📱 Acceso directo en el Menú de Aplicaciones: %s\n", desktopAppPath)
	fmt.Println(" 👉 ¡Ahora puedes hacer doble clic en 'iniciar.sh' o en el icono de tu Escritorio para abrirlo!")
}

func killServer(pid int) {
	process, err := os.FindProcess(pid)
	if err == nil {
		process.Kill()
		os.Remove(pidFile)
		fmt.Println("\n🛑 Servidor detenido con éxito. RAM liberada.")
	}
}

package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"llamamanager/internal/agent"
	"llamamanager/internal/config"
)

// agentSession mantiene el estado de la sesión del agente en memoria
type agentSession struct {
	mu       sync.Mutex
	messages []agent.ChatMessage
	client   *agent.LLMClient
}

var session *agentSession

// getSystemContext recolecta información básica del sistema para inyectar en el prompt
func getSystemContext() string {
	var ctx strings.Builder
	ctx.WriteString("\n\n--- CONTEXTO DEL SISTEMA AL INICIAR ---\n")
	
	if out, err := exec.Command("uname", "-snrvm").Output(); err == nil {
		ctx.WriteString("OS: " + strings.TrimSpace(string(out)) + "\n")
	}
	
	if out, err := exec.Command("free", "-m").Output(); err == nil {
		ctx.WriteString("Memoria:\n" + strings.TrimSpace(string(out)) + "\n")
	}
	
	if out, err := exec.Command("df", "-h", "/").Output(); err == nil {
		ctx.WriteString("Disco (/):\n" + strings.TrimSpace(string(out)) + "\n")
	}
	
	return ctx.String()
}

func initAgentSession() {
	conf := config.LoadSettings()
	port := 8080
	if conf.APIPort != "" {
		fmt.Sscanf(conf.APIPort, "%d", &port)
	}

	session = &agentSession{
		messages: []agent.ChatMessage{
			{Role: "system", Content: agent.SystemPrompt + getSystemContext()},
		},
		client: agent.NewLLMClient(port),
	}
}

// Petición del frontend
type agentChatRequest struct {
	Message string `json:"message"`
}

// Respuesta al frontend
type agentChatResponse struct {
	Reply     string `json:"reply"`
	Command   string `json:"command,omitempty"`
	Dangerous bool   `json:"dangerous,omitempty"`
	ReadOnly  bool   `json:"read_only,omitempty"`
}

// agentExecuteRequest para ejecutar un comando aprobado
type agentExecuteRequest struct {
	Command string `json:"command"`
}

type agentExecuteResponse struct {
	Output  string `json:"output"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// handleAgentChat procesa un mensaje del usuario y devuelve la respuesta del LLM
func handleAgentChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	var req agentChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "JSON inválido"})
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "El mensaje no puede estar vacío"})
		return
	}

	// Inicializar sesión si no existe
	if session == nil {
		initAgentSession()
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	// Añadir mensaje del usuario al historial
	session.messages = append(session.messages, agent.ChatMessage{Role: "user", Content: req.Message})

	// Consultar al LLM
	response, err := session.client.Ask(session.messages)
	if err != nil {
		// Retirar el mensaje fallido
		session.messages = session.messages[:len(session.messages)-1]
		json.NewEncoder(w).Encode(map[string]interface{}{"error": fmt.Sprintf("Error conectando con el modelo: %v", err)})
		return
	}

	// Guardar respuesta en historial
	session.messages = append(session.messages, agent.ChatMessage{Role: "assistant", Content: response})

	// Analizar si la respuesta contiene un comando
	cmdStr := agent.ParseCommand(response)

	resp := agentChatResponse{
		Reply: response,
	}

	if cmdStr != "" {
		resp.Command = cmdStr
		resp.Dangerous = agent.IsDangerous(cmdStr)
		resp.ReadOnly = agent.IsReadOnly(cmdStr)
	}

	json.NewEncoder(w).Encode(resp)
}

// handleAgentExecute ejecuta un comando aprobado por el usuario
func handleAgentExecute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	var req agentExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "JSON inválido"})
		return
	}

	if strings.TrimSpace(req.Command) == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "El comando no puede estar vacío"})
		return
	}

	// Verificación de seguridad adicional en el backend
	if agent.IsDangerous(req.Command) {
		json.NewEncoder(w).Encode(agentExecuteResponse{
			Success: false,
			Error:   "Comando bloqueado por seguridad: parece altamente destructivo.",
		})
		return
	}

	// Ejecutar el comando de forma multiplataforma
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", req.Command)
	} else {
		cmd = exec.Command("bash", "-c", req.Command)
	}
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)

	// Limitar la salida a 8KB para no sobrecargar la interfaz
	if len(output) > 8192 {
		output = output[:8192] + "\n... (salida truncada)"
	}

	resp := agentExecuteResponse{
		Output:  output,
		Success: err == nil,
	}

	if err != nil {
		resp.Error = err.Error()
	}

	// Informar al historial del agente sobre el resultado
	if session != nil {
		session.mu.Lock()
		if err != nil {
			session.messages = append(session.messages, agent.ChatMessage{
				Role:    "system",
				Content: fmt.Sprintf("El comando '%s' falló con error: %v\nSalida: %s", req.Command, err, output),
			})
		} else {
			session.messages = append(session.messages, agent.ChatMessage{
				Role:    "system",
				Content: fmt.Sprintf("El comando '%s' se ejecutó exitosamente.\nSalida: %s", req.Command, output),
			})
		}
		session.mu.Unlock()
	}

	json.NewEncoder(w).Encode(resp)
}

// handleAgentReset reinicia la sesión del agente
func handleAgentReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	initAgentSession()
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "msg": "Sesión del agente reiniciada."})
}

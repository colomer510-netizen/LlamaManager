package ui

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
)

type AppSettings struct {
	ForceCPU      bool   `json:"force_cpu"`
	CustomGPU     int    `json:"custom_gpu"`
	CustomContext int    `json:"custom_context"`
	CustomThreads int    `json:"custom_threads"`
	SystemPrompt  string `json:"system_prompt"`
}

var SettingsFile = "llama_settings.json"

func LoadSettings() AppSettings {
	var s AppSettings
	// Valores predeterminados
	s.CustomGPU = -1
	s.CustomContext = 0
	s.CustomThreads = 0
	s.SystemPrompt = "Eres un asistente virtual experto."

	data, err := os.ReadFile(SettingsFile)
	if err == nil {
		json.Unmarshal(data, &s)
	}
	return s
}

func SaveSettings(s AppSettings) {
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(SettingsFile, data, 0644)
}

func RunSettingsMenu() error {
	settings := LoadSettings()
	clearScreen()

	// Convertir enteros a string para los inputs
	gpuStr := fmt.Sprintf("%d", settings.CustomGPU)
	ctxStr := fmt.Sprintf("%d", settings.CustomContext)
	thrStr := fmt.Sprintf("%d", settings.CustomThreads)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("1. Forzar Modo CPU").
				Description("Si Llama crashea en tu GPU (Error 0xc0000409), activa esto.").
				Value(&settings.ForceCPU),

			huh.NewInput().
				Title("2. Capas GPU (-ngl)").
				Description("Pon -1 para Auto, 0 para desactivar, o ej. 15 para gráficas débiles.").
				Value(&gpuStr),

			huh.NewInput().
				Title("3. Tamaño de Contexto (-c)").
				Description("0 para Auto. Usa 1024, 2048, 4096, u 8192.").
				Value(&ctxStr),

			huh.NewInput().
				Title("4. Hilos de Procesador (-t)").
				Description("0 para Auto. Número de núcleos de CPU a usar.").
				Value(&thrStr),

			huh.NewText().
				Title("5. System Prompt (Personalidad)").
				Description("Instrucciones ocultas para la IA.").
				Value(&settings.SystemPrompt),
		),
	)

	err := form.Run()
	if err != nil {
		return err
	}

	// Guardar de vuelta
	fmt.Sscanf(gpuStr, "%d", &settings.CustomGPU)
	fmt.Sscanf(ctxStr, "%d", &settings.CustomContext)
	fmt.Sscanf(thrStr, "%d", &settings.CustomThreads)

	SaveSettings(settings)
	return nil
}

package main

import (
    "io/ioutil"
    "strings"
)

func main() {
    content, _ := ioutil.ReadFile("internal/web/server.go")
    str := string(content)
    
    start := strings.Index(str, "exePath, err := tools.ResolveBinPath(\"llama\")")
    end := strings.Index(str, "json.NewEncoder(w).Encode(map[string]interface{}{\"error\": \"No se pudo leer la recomendación del motor.\"})\n}")
    
    if start != -1 && end != -1 {
        newFuncBody := `
	specs, _ := hardware.GetSystemSpecs()
	
	// Determinar valores seguros basados en la RAM
	ctxSize := 4096
	gpuLayers := 0 // Por defecto 0 si no hay mucha RAM, pero podemos poner 99 para auto-offload
	
	if specs != nil {
	    if specs.TotalRAMGB >= 16 {
	        ctxSize = 8192
	    }
	    if specs.TotalRAMGB >= 32 {
	        ctxSize = 16384
	    }
	    // Asumir que pueden tener GPU y usar auto-offload (llama.cpp lo maneja con -ngl 99 limitando a la VRAM disponible)
	    gpuLayers = 99
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"context_size": ctxSize,
		"gpu_layers": gpuLayers,
		"msg": "Configuración óptima calculada (basada en RAM disponible).",
	})
}
`
        // Replace from start to end + length of the end string
        replaced := str[:start] + newFuncBody + str[end + len("json.NewEncoder(w).Encode(map[string]interface{}{\"error\": \"No se pudo leer la recomendación del motor.\"})\n}"):]
        ioutil.WriteFile("internal/web/server.go", []byte(replaced), 0644)
    }
}

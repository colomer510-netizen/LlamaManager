package hardware

import (
	"runtime"
)

// SystemSpecs contiene la información de hardware detectada
type SystemSpecs struct {
	TotalRAMGB   int    `json:"total_ram_gb"`
	LogicalCores int    `json:"logical_cores"`
	CpuModel     string `json:"cpu_model"`
}

// GetSystemSpecs escanea y devuelve las especificaciones del hardware (RAM y CPU)
func GetSystemSpecs() (*SystemSpecs, error) {
	ram, err := getTotalRAMGB()
	if err != nil {
		return nil, err
	}

	return &SystemSpecs{
		TotalRAMGB:   ram,
		LogicalCores: runtime.NumCPU(),
		CpuModel:     "Intel/AMD CPU",
	}, nil
}

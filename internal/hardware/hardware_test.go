package hardware

import (
	"runtime"
	"testing"
)

func TestGetSystemSpecs(t *testing.T) {
	specs, err := GetSystemSpecs()
	if err != nil {
		t.Fatalf("GetSystemSpecs falló en Linux: %v", err)
	}

	if specs.TotalRAMGB <= 0 {
		t.Errorf("La memoria RAM detectada debe ser > 0, obtenido: %d GB", specs.TotalRAMGB)
	}

	if specs.LogicalCores != runtime.NumCPU() {
		t.Errorf("Hilos lógicos detectados (%d) no coinciden con NumCPU (%d)", specs.LogicalCores, runtime.NumCPU())
	}
}

func TestBuildOptimalConfig(t *testing.T) {
	cfg, err := BuildOptimalConfig("", true, 0, 4096, 4)
	if err != nil {
		t.Fatalf("BuildOptimalConfig falló: %v", err)
	}

	if cfg.CtxFlag != "4096" {
		t.Errorf("CtxFlag incorrecto: %s, esperado 4096", cfg.CtxFlag)
	}

	if cfg.ThreadFlag != "4" {
		t.Errorf("ThreadFlag incorrecto: %s, esperado 4", cfg.ThreadFlag)
	}

	if cfg.GPUFlag != "0" {
		t.Errorf("GPUFlag incorrecto: %s, esperado 0 (forceCPU=true)", cfg.GPUFlag)
	}
}

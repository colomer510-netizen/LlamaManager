package main

import (
	"fmt"
	"llamamanager/internal/hardware"
)

func main() {
	specs, err := hardware.GetSystemSpecs()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("RAM: %d GB, CPU: %d cores\n", specs.TotalRAMGB, specs.LogicalCores)
}

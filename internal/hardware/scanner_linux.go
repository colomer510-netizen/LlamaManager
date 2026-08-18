//go:build linux
// +build linux

package hardware

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func getTotalRAMGB() (int, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return 0, err
				}
				return int(kb / (1024 * 1024)), nil
			}
		}
	}
	return 0, scanner.Err()
}

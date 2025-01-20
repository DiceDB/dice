// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package observability

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type HardwareConfig struct {
	OS            string `json:"os"`
	CPU           string `json:"cpu"`
	TotalMemoryMB int64  `json:"total_memory_mb"`
	TotalDiskMB   int64  `json:"total_disk_mb"`
	TotalCores    int    `json:"total_cores"`
}

func GetHardwareMeta() (HardwareConfig, error) {
	var res HardwareConfig
	totalMemory, err := getTotalMemoryMB()
	if err != nil {
		return res, err
	}

	totalDisk, err := getTotalDiskMB()
	if err != nil {
		return res, err
	}

	os := runtime.GOOS
	cpu := runtime.GOARCH
	cores := getTotalCores()

	return HardwareConfig{
		TotalMemoryMB: totalMemory,
		TotalDiskMB:   totalDisk,
		TotalCores:    cores,
		OS:            os,
		CPU:           cpu,
	}, nil
}

func getTotalCores() int {
	return runtime.NumCPU()
}

func getTotalMemoryMB() (int64, error) {
	var res string
	var err error
	switch runtime.GOOS {
	case linux:
		res, err = executeCommand("free -b | grep Mem | awk '{print $2}'")
		if err != nil {
			return 0, fmt.Errorf("error:stats.total_memory_mb failed to capture memory err=%w", err)
		}
	case darwin:
		res, err = executeCommand("sysctl -n hw.memsize")
		if err != nil {
			return 0, fmt.Errorf("error:stats.total_memory_mb failed to capture memory err=%w", err)
		}
	case windows:
		res, err = executeCommand("wmic OS get TotalVisibleMemorySize /Value")
		if err != nil {
			return 0, fmt.Errorf("error:stats.total_memory_mb failed to capture memory err=%w", err)
		}

		parts := strings.Split(res, "=")
		if len(parts) != 2 {
			return 0, fmt.Errorf("error:stats.total_memory_mb unexpected output format: %s", res)
		}

		res = strings.TrimSpace(parts[1])
	default:
		return 0, fmt.Errorf("error:stats.total_memory_mb unsupported platform")
	}

	v, err := strconv.ParseFloat(res, 64)
	if err != nil {
		return 0, fmt.Errorf("error:stats.total_memory_mb not a number: %s", res)
	}

	return int64(v) / 1024 / 1024, nil
}

func getTotalDiskMB() (int64, error) {
	var res string
	var err error
	switch runtime.GOOS {
	case linux:
		res, err = executeCommand("df --block-size=1 / | tail -1 | awk '{print $2}'")
		if err != nil {
			return 0, fmt.Errorf("error:stats.total_disk_mb failed to capture disk usage err=%w", err)
		}
	case darwin:
		res, err = executeCommand("df -k / | tail -1 | awk '{print $2}'")
		if err != nil {
			return 0, fmt.Errorf("error:stats.total_disk_mb failed to capture disk usage err=%w", err)
		}
	case windows:
		res, err = executeCommand("wmic logicaldisk get size /Value")
		if err != nil {
			return 0, fmt.Errorf("error:stats.total_disk_mb failed to capture disk usage err=%w", err)
		}

		parts := strings.Split(res, "=")
		if len(parts) != 2 {
			return 0, fmt.Errorf("error:stats.total_disk_mb unexpected output format: %s", res)
		}

		res = strings.TrimSpace(parts[1])
	default:
		return 0, fmt.Errorf("error:stats.total_disk_mb unsupported platform")
	}

	v, err := strconv.ParseFloat(res, 64)
	if err != nil {
		return 0, fmt.Errorf("error:stats.total_disk_mb not a number: %s", res)
	}

	return int64(v) / 1024 / 1024, nil
}

func executeCommand(cmd string) (string, error) {
	var out []byte
	var err error

	if runtime.GOOS == windows {
		out, err = exec.Command("cmd", "/C", cmd).Output()
	} else {
		out, err = exec.Command("sh", "-c", cmd).Output()
	}

	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

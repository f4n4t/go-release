package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

// IsSSD determines if the storage device of a given file path is an SSD by checking its rotational status.
// Errors are logged, and false is returned on failure.
func IsSSD(filePath string) bool {
	if runtime.GOOS != "linux" {
		return false
	}

	deviceID, err := getDeviceID(filePath)
	if err != nil {
		return false
	}

	blockDevice, err := findBlockDevice(deviceID)
	if err != nil {
		return false
	}

	deviceName := resolveBlockDevice(blockDevice)
	if deviceName == "" {
		return false
	}

	rotationalPath := fmt.Sprintf("/sys/block/%s/queue/rotational", deviceName)
	rotationalData, err := os.ReadFile(rotationalPath)
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(rotationalData)) == "0"
}

// getDeviceID returns the device ID for a given file path
func getDeviceID(filePath string) (uint64, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return 0, fmt.Errorf("resolve absolute path: %w", err)
	}

	fileStat := new(unix.Stat_t)
	if err = unix.Stat(absPath, fileStat); err != nil {
		return 0, fmt.Errorf("stat file: %w", err)
	}

	return fileStat.Dev, nil
}

// findBlockDevice locates the block device associated with a given device ID by scanning /proc/mounts
func findBlockDevice(deviceID uint64) (string, error) {
	procMounts, err := os.Open("/proc/mounts")
	if err != nil {
		return "", fmt.Errorf("open /proc/mounts: %w", err)
	}
	defer procMounts.Close()

	scanner := bufio.NewScanner(procMounts)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 {
			device, mountPoint := fields[0], fields[1]
			if strings.HasPrefix(device, "/dev/") {
				mountStat := new(unix.Stat_t)
				if err := unix.Stat(mountPoint, mountStat); err == nil && mountStat.Dev == deviceID {
					return device, nil
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan /proc/mounts: %w", err)
	}

	return "", fmt.Errorf("block device not found for device ID: %d", deviceID)
}

// resolveBlockDevice resolves the base block device from a symlink or partition path
func resolveBlockDevice(device string) string {
	resolvedDevice, err := filepath.EvalSymlinks(device)
	if err != nil {
		return ""
	}

	// Example: resolvedDevice could be "/dev/nvme0n1p4"
	base := filepath.Base(resolvedDevice)

	// loop until the base is not a partition
	for strings.HasSuffix(base, "p") || len(base) > 3 && base[len(base)-2] == 'p' {
		base = base[:len(base)-1]
	}

	if _, err := os.Stat(filepath.Join("/sys/block", base)); errors.Is(err, os.ErrNotExist) {
		return ""
	}

	return base
}

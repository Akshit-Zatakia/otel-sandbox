package utils

import (
	"fmt"
	"runtime"
)

func GetOSArch() (string, string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	switch goos {
	case "linux", "darwin", "windows":
	default:
		return "", "", fmt.Errorf("unsupported OS: %s", goos)
	}

	switch goarch {
	case "amd64", "arm64":
	default:
		return "", "", fmt.Errorf("unsupported architecture: %s", goarch)
	}

	return goos, goarch, nil
}

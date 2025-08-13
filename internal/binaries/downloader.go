package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Akshit-Zatakia/otel-sandbox/internal/utils"
)

type Binary struct {
	Name string
	URL  string
}

var binaries = []Binary{
	{
		Name: "otelcol",
		URL:  "https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.131.1/otelcol-contrib_0.131.1_%s_%s.tar.gz",
	},
	{
		Name: "jaeger-all-in-one",
		URL:  "https://github.com/jaegertracing/jaeger/releases/download/v1.72.0/jaeger-1.72.0-%s-%s.tar.gz",
	},
	{
		Name: "prometheus",
		URL:  "https://github.com/prometheus/prometheus/releases/download/v3.5.0/prometheus-3.5.0.%s-%s.tar.gz",
	},
}

// DownloadAll downloads all binaries if not present
func DownloadAll(binDir string) error {
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin dir: %w", err)
	}

	goos, goarch, err := utils.GetOSArch()
	if err != nil {
		return err
	}

	for _, bin := range binaries {
		if err := downloadBinary(bin, binDir, goos, goarch); err != nil {
			return err
		}
	}
	return nil
}

func downloadBinary(bin Binary, binDir string, goos, goarch string) error {
	binPath := filepath.Join(binDir, bin.Name)
	if fileExists(binPath) {
		fmt.Printf("✅ %s already exists, skipping\n", bin.Name)
		return nil
	}

	url := fmt.Sprintf(bin.URL, goos, goarch)
	fmt.Printf("⬇️  Downloading %s from %s\n", bin.Name, url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", bin.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status for %s: %s", bin.Name, resp.Status)
	}

	tmpFile := filepath.Join(binDir, bin.Name+".tmp")
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save %s: %w", bin.Name, err)
	}

	if strings.HasSuffix(url, ".tar.gz") {
		if err := extractTarGz(tmpFile, binDir); err != nil {
			return err
		}
	} else if strings.HasSuffix(url, ".zip") {
		if err := extractZip(tmpFile, binDir); err != nil {
			return err
		}
	} else {
		os.Rename(tmpFile, binPath)
	}

	os.Chmod(binPath, 0755)
	fmt.Printf("✅ %s installed at %s\n", bin.Name, binPath)
	return nil
}

func extractTarGz(filePath, dest string) error {
	cmd := exec.Command("tar", "-xzf", filePath, "-C", dest)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract %s: %w", filePath, err)
	}
	return os.Remove(filePath)
}

func extractZip(filePath, dest string) error {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fPath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fPath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return os.Remove(filePath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

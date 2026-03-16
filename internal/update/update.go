package update

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const repo = "tomassar/dioptra"

// Version is set at build time via ldflags.
// Falls back to "dev" when running from source.
var Version = "dev"

type release struct {
	TagName string `json:"tag_name"`
}

// LatestVersion fetches the latest published release tag from GitHub.
func LatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("could not reach GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub returned %d", resp.StatusCode)
	}

	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("could not parse response: %w", err)
	}
	return r.TagName, nil
}

// IsNewer returns true if latest > current (simple string compare on vX.Y.Z tags).
func IsNewer(current, latest string) bool {
	if current == "dev" || current == "" {
		return false
	}
	// Strip leading 'v' for comparison
	c := strings.TrimPrefix(current, "v")
	l := strings.TrimPrefix(latest, "v")
	return l > c
}

// Do downloads the latest release binary and replaces the running executable.
func Do(latest string) error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	archive := fmt.Sprintf("dioptra_%s_%s.tar.gz", goos, goarch)
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, latest, archive)

	fmt.Printf("Downloading dioptra %s (%s/%s)...\n", latest, goos, goarch)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Extract binary from tar.gz
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)

	var binData []byte
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Name == "dioptra" {
			binData, err = io.ReadAll(tr)
			if err != nil {
				return err
			}
			break
		}
	}
	if binData == nil {
		return fmt.Errorf("binary not found in archive")
	}

	// Get path of running executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find current executable: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return err
	}

	// Write to a temp file next to the binary, then rename (atomic on same fs)
	tmpPath := exePath + ".tmp"
	if err := os.WriteFile(tmpPath, binData, 0755); err != nil {
		return fmt.Errorf("could not write update (try with sudo): %w", err)
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("could not replace binary (try with sudo): %w", err)
	}

	fmt.Printf("Updated to %s. Run 'dioptra --version' to confirm.\n", latest)
	return nil
}

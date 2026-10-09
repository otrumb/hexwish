package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type moduleDownload struct {
	Path    string
	Version string
	Dir     string
	Zip     string
}

func downloadModule(path, version string) (moduleDownload, error) {
	output, err := exec.Command("go", "mod", "download", "-json", path+"@"+version).Output()
	if err != nil {
		return moduleDownload{}, fmt.Errorf("download module source %s@%s: %w", path, version, err)
	}
	var download moduleDownload
	if err := json.Unmarshal(output, &download); err != nil {
		return moduleDownload{}, fmt.Errorf("decode module source %s@%s: %w", path, version, err)
	}
	if download.Path != path || download.Version != version || download.Zip == "" {
		return moduleDownload{}, fmt.Errorf("module source mismatch for %s@%s", path, version)
	}
	return download, nil
}

func moduleDir(location moduleLocation, download func(string, string) (moduleDownload, error)) (string, error) {
	if location.Dir != "" {
		return location.Dir, nil
	}
	result, err := download(location.Path, location.Version)
	if err != nil {
		return "", err
	}
	if result.Path != location.Path || result.Version != location.Version || result.Dir == "" {
		return "", fmt.Errorf("module directory mismatch for %s@%s", location.Path, location.Version)
	}
	return result.Dir, nil
}

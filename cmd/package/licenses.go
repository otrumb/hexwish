package main

import (
	"debug/buildinfo"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type moduleLocation struct {
	Path    string
	Version string
	Dir     string
}

type moduleDownload struct {
	Path    string
	Version string
	Zip     string
}

func releaseFiles(binary string) (map[string][]byte, error) {
	files := make(map[string][]byte)
	for _, name := range []string{"LICENSE", "README.md", "RELINKING.md", "THIRD_PARTY_NOTICES.md", "go.mod", "go.sum"} {
		content, err := os.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		files[name] = content
	}
	binaryContent, err := os.ReadFile(binary)
	if err != nil {
		return nil, fmt.Errorf("read binary: %w", err)
	}
	files[filepath.Base(binary)] = binaryContent
	if err := addSource(files); err != nil {
		return nil, err
	}
	if err := addLicenses(files, binary); err != nil {
		return nil, err
	}
	if err := addGoLicense(files); err != nil {
		return nil, err
	}
	return files, nil
}

func sourceFiles(binaries []string) (map[string][]byte, error) {
	files := make(map[string][]byte)
	if err := addProjectSource(files); err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	for _, binary := range binaries {
		info, err := buildinfo.ReadFile(binary)
		if err != nil {
			return nil, fmt.Errorf("read build metadata: %w", err)
		}
		for _, dependency := range info.Deps {
			path, version := dependency.Path, dependency.Version
			if dependency.Replace != nil {
				path, version = dependency.Replace.Path, dependency.Replace.Version
			}
			key := path + "@" + version
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			download, err := downloadModule(path, version)
			if err != nil {
				return nil, err
			}
			content, err := os.ReadFile(download.Zip)
			if err != nil {
				return nil, fmt.Errorf("read module source %s: %w", key, err)
			}
			name := strings.ReplaceAll(path, "/", "_") + "@" + version + ".zip"
			files["modules/"+name] = content
		}
	}
	return files, nil
}

func addProjectSource(files map[string][]byte) error {
	for _, name := range []string{"LICENSE", "README.md", "RELINKING.md", "SECURITY.md", "THIRD_PARTY_NOTICES.md", "THREAT_MODEL.md", "go.mod", "go.sum", "main.go"} {
		content, err := os.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read source %s: %w", name, err)
		}
		files["source/"+name] = content
	}
	rootGo, err := filepath.Glob("*.go")
	if err != nil {
		return err
	}
	for _, path := range rootGo {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files["source/"+filepath.ToSlash(path)] = content
	}
	for _, root := range []string{".github", "cmd", "internal"} {
		if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || (filepath.Ext(path) != ".go" && filepath.Ext(path) != ".yml") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files["source/"+filepath.ToSlash(path)] = content
			return nil
		}); err != nil {
			return fmt.Errorf("collect source %s: %w", root, err)
		}
	}
	return nil
}

func addGoLicense(files map[string][]byte) error {
	for _, name := range []string{"LICENSE", "PATENTS"} {
		content, err := os.ReadFile(filepath.Join(runtime.GOROOT(), name))
		if err != nil {
			return fmt.Errorf("read Go %s: %w", name, err)
		}
		files["licenses/go/"+name] = content
	}
	return nil
}

func downloadModule(path, version string) (moduleDownload, error) {
	command := exec.Command("go", "mod", "download", "-json", path+"@"+version)
	output, err := command.Output()
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

func addSource(files map[string][]byte) error {
	return filepath.WalkDir(".", func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path == ".git" || path == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files["source/"+filepath.ToSlash(path)] = content
		return nil
	})
}

func addLicenses(files map[string][]byte, binary string) error {
	info, err := buildinfo.ReadFile(binary)
	if err != nil {
		return fmt.Errorf("read build metadata: %w", err)
	}
	locations, err := moduleLocations()
	if err != nil {
		return err
	}
	lines := make([]string, 0, len(info.Deps))
	for _, dependency := range info.Deps {
		path, version := dependency.Path, dependency.Version
		if dependency.Replace != nil {
			path, version = dependency.Replace.Path, dependency.Replace.Version
		}
		location, ok := locations[path+"@"+version]
		if !ok {
			return fmt.Errorf("module directory missing for %s@%s", path, version)
		}
		names, err := licenseNames(location.Dir)
		if err != nil {
			return fmt.Errorf("module %s@%s: %w", path, version, err)
		}
		bundleDir := "licenses/" + strings.ReplaceAll(path, "/", "_") + "@" + version
		for _, name := range names {
			content, err := os.ReadFile(filepath.Join(location.Dir, name))
			if err != nil {
				return err
			}
			files[bundleDir+"/"+name] = content
		}
		lines = append(lines, path+" "+version+" "+strings.Join(names, ","))
	}
	sort.Strings(lines)
	files["licenses/MANIFEST.txt"] = []byte(strings.Join(lines, "\n") + "\n")
	return nil
}

func moduleLocations() (map[string]moduleLocation, error) {
	command := exec.Command("go", "list", "-m", "-json", "all")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	locations := make(map[string]moduleLocation)
	for decoder.More() {
		var location moduleLocation
		if err := decoder.Decode(&location); err != nil {
			return nil, fmt.Errorf("decode module: %w", err)
		}
		locations[location.Path+"@"+location.Version] = location
	}
	return locations, nil
}

func licenseNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		upper := strings.ToUpper(entry.Name())
		if !entry.IsDir() && (strings.HasPrefix(upper, "LICENSE") || strings.HasPrefix(upper, "COPYING") || strings.HasPrefix(upper, "NOTICE")) {
			names = append(names, entry.Name())
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no license text in %s", dir)
	}
	sort.Strings(names)
	return names, nil
}

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type binaryPaths []string

func (paths *binaryPaths) String() string { return fmt.Sprint([]string(*paths)) }
func (paths *binaryPaths) Set(value string) error {
	*paths = append(*paths, value)
	return nil
}

func main() {
	var binaries binaryPaths
	flag.Var(&binaries, "binary", "final binary to package; repeat for source union")
	platform := flag.String("platform", "", "linux-amd64, windows-amd64, or source")
	output := flag.String("out", "", "existing output directory")
	flag.Parse()
	if len(binaries) == 0 || *output == "" || (*platform != "linux-amd64" && *platform != "windows-amd64" && *platform != "source") {
		fmt.Fprintln(os.Stderr, "package: -binary, -out, and supported -platform are required")
		os.Exit(2)
	}
	info, err := os.Stat(*output)
	if err != nil || !info.IsDir() {
		fmt.Fprintln(os.Stderr, "package: -out must be an existing directory")
		os.Exit(2)
	}
	if *platform != "source" && len(binaries) != 1 {
		fmt.Fprintln(os.Stderr, "package: binary archive requires exactly one -binary")
		os.Exit(2)
	}
	files, err := releaseFiles(binaries[0])
	if *platform == "source" {
		files, err = sourceFiles(binaries)
	}
	if err == nil {
		if *platform == "linux-amd64" {
			err = writeTarGzip(filepath.Join(*output, "hexwish-linux-amd64.tar.gz"), files)
		} else if *platform == "windows-amd64" {
			err = writeZip(filepath.Join(*output, "hexwish-windows-amd64.zip"), files)
		} else {
			err = writeTarGzip(filepath.Join(*output, "hexwish-source-v0.1.0.tar.gz"), files)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "package: assemble release:", err)
		os.Exit(1)
	}
}

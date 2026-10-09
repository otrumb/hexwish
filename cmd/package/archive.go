package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"sort"
	"time"
)

var archiveTime = time.Unix(0, 0).UTC()

func writeZip(path string, files map[string][]byte) (returnErr error) {
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { returnErr = closeWithError(returnErr, output) }()
	archive := zip.NewWriter(output)
	defer func() { returnErr = closeWithError(returnErr, archive) }()
	for _, name := range sortedNames(files) {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate, Modified: archiveTime}
		header.SetMode(0o644)
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := writer.Write(files[name]); err != nil {
			return err
		}
	}
	return nil
}

func writeTarGzip(path string, files map[string][]byte) (returnErr error) {
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { returnErr = closeWithError(returnErr, output) }()
	compressed, err := gzip.NewWriterLevel(output, gzip.BestCompression)
	if err != nil {
		return err
	}
	compressed.Header.ModTime = archiveTime
	compressed.Header.OS = 255
	defer func() { returnErr = closeWithError(returnErr, compressed) }()
	archive := tar.NewWriter(compressed)
	defer func() { returnErr = closeWithError(returnErr, archive) }()
	for _, name := range sortedNames(files) {
		mode := int64(0o644)
		if name == "hexwish" {
			mode = 0o755
		}
		header := &tar.Header{Name: name, Mode: mode, Size: int64(len(files[name])), ModTime: archiveTime, Format: tar.FormatGNU}
		if err := archive.WriteHeader(header); err != nil {
			return err
		}
		if _, err := archive.Write(files[name]); err != nil {
			return err
		}
	}
	return nil
}

func sortedNames(files map[string][]byte) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func closeWithError(current error, closer io.Closer) error {
	if err := closer.Close(); current == nil {
		return err
	}
	return current
}

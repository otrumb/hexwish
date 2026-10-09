package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestArchives_are_byte_identical_when_repeated(t *testing.T) {
	files := map[string][]byte{"b": []byte("two"), "a": []byte("one")}
	first := filepath.Join(t.TempDir(), "first.zip")
	second := filepath.Join(t.TempDir(), "second.zip")
	if err := writeZip(first, files); err != nil {
		t.Fatal(err)
	}
	if err := writeZip(second, files); err != nil {
		t.Fatal(err)
	}
	firstBytes, err := os.ReadFile(first)
	if err != nil {
		t.Fatal(err)
	}
	secondBytes, err := os.ReadFile(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatal("repeated archives differ")
	}
}

func TestLicenseNames_fails_when_module_has_no_license(t *testing.T) {
	if _, err := licenseNames(t.TempDir()); err == nil {
		t.Fatal("accepted module without license")
	}
}

func TestLicenseNames_includes_all_root_license_material(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"LICENSE", "COPYING", "COPYING.LESSER", "NOTICE"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	names, err := licenseNames(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 4 {
		t.Fatalf("got %v", names)
	}
}

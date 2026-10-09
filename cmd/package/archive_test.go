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

func TestAddGoLicense_includes_toolchain_terms(t *testing.T) {
	files := make(map[string][]byte)
	if err := addGoLicense(files); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"licenses/go/LICENSE", "licenses/go/PATENTS"} {
		if len(files[name]) == 0 {
			t.Fatalf("missing %s", name)
		}
	}
}

func TestModuleDir_downloads_module_when_list_has_no_directory(t *testing.T) {
	want := t.TempDir()
	got, err := moduleDir(moduleLocation{Path: "example.test/module", Version: "v1.2.3"}, func(path, version string) (moduleDownload, error) {
		if path != "example.test/module" || version != "v1.2.3" {
			t.Fatalf("download %s@%s", path, version)
		}
		return moduleDownload{Path: path, Version: version, Dir: want}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

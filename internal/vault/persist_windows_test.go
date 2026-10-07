//go:build windows

package vault

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPersistNoReplace_rejects_existing_junction(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	destination := filepath.Join(dir, "key.json")
	if output, err := exec.Command("cmd.exe", "/c", "mkdir", target).CombinedOutput(); err != nil {
		t.Fatalf("create target: %v: %s", err, output)
	}
	if output, err := exec.Command("cmd.exe", "/c", "mklink", "/J", destination, target).CombinedOutput(); err != nil {
		t.Fatalf("create junction without special privilege: %v: %s", err, output)
	}
	if err := PersistNoReplace(destination, []byte("secret")); err == nil {
		t.Fatal("claimed existing junction")
	}
}

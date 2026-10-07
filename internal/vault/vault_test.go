package vault

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func TestEncryptVerify_roundtrips_standard_scrypt(t *testing.T) {
	key, err := gethcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := Encrypt(key, []byte("TEST ONLY NEVER FUND"))
	if err != nil {
		t.Fatal(err)
	}
	address, err := Verify(encoded, []byte("TEST ONLY NEVER FUND"))
	if err != nil {
		t.Fatal(err)
	}
	if address != gethcrypto.PubkeyToAddress(key.PublicKey) {
		t.Fatalf("got %s", address)
	}
}

func TestVerify_rejects_oversize_before_KDF(t *testing.T) {
	if _, err := Verify(bytes.Repeat([]byte("x"), MaxInput+1), []byte("x")); err == nil {
		t.Fatal("accepted oversize")
	}
}

func TestVerify_rejects_unsafe_KDF_profile(t *testing.T) {
	data := []byte(`{"version":3,"crypto":{"kdf":"scrypt","kdfparams":{"n":2,"r":8,"p":1,"dklen":32}}}`)
	if _, err := Verify(data, []byte("x")); err == nil {
		t.Fatal("accepted weak profile")
	}
}

func TestPersistNoReplace_does_not_overwrite(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "key.json")
	if err := os.WriteFile(destination, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := PersistNoReplace(destination, []byte("new")); err == nil {
		t.Fatal("overwrote destination")
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "existing" {
		t.Fatalf("got %q", got)
	}
}

func TestPersistNoReplace_claims_complete_file(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "key.json")
	data := []byte(`{"version":3}`)
	if err := PersistNoReplace(destination, data); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("got %q", got)
	}
}

func TestPersistNoReplace_rejects_dangling_symlink(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "key.json")
	if err := os.Symlink(filepath.Join(dir, "missing"), destination); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := PersistNoReplace(destination, []byte("secret")); err == nil {
		t.Fatal("claimed dangling symlink")
	}
}

func TestPersistNoReplace_allows_one_concurrent_claim(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "key.json")
	start := make(chan struct{})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for _, data := range [][]byte{[]byte("one"), []byte("two")} {
		wait.Add(1)
		go func() { defer wait.Done(); <-start; results <- PersistNoReplace(destination, data) }()
	}
	close(start)
	wait.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("got %d successful claims", successes)
	}
}

func TestVerify_rejects_wrong_password_and_address_mismatch(t *testing.T) {
	key, err := gethcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := Encrypt(key, []byte("TEST ONLY NEVER FUND"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(encoded, []byte("wrong")); err == nil {
		t.Fatal("accepted wrong password")
	}
	var document map[string]any
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatal(err)
	}
	document["address"] = "0000000000000000000000000000000000000000"
	tampered, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(tampered, []byte("TEST ONLY NEVER FUND")); err == nil {
		t.Fatal("accepted address mismatch")
	}
}

var _ = keystore.StandardScryptN

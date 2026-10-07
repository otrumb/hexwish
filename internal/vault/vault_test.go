package vault

import (
	"bytes"
	"os"
	"path/filepath"
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

var _ = keystore.StandardScryptN

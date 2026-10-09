package vault

import (
	"strings"
	"testing"
)

const validV3 = `{"address":"0000000000000000000000000000000000000000","crypto":{"cipher":"aes-128-ctr","cipherparams":{"iv":"00000000000000000000000000000000"},"ciphertext":"0000000000000000000000000000000000000000000000000000000000000000","kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"0000000000000000000000000000000000000000000000000000000000000000"},"mac":"0000000000000000000000000000000000000000000000000000000000000000"},"id":"00000000-0000-4000-8000-000000000000","version":3}`

func TestParseEnvelope_rejects_noncanonical_documents(t *testing.T) {
	tests := map[string]string{
		"duplicate exact":  strings.Replace(validV3, `"version":3`, `"version":3,"version":3`, 1),
		"weak shadow":      strings.Replace(validV3, `"n":262144`, `"n":2,"N":262144`, 1),
		"dangerous shadow": strings.Replace(validV3, `"n":262144`, `"n":1073741824,"N":262144`, 1),
		"missing salt":     strings.Replace(validV3, `,"salt":"`+strings.Repeat("0", 64)+`"`, "", 1),
		"wrong salt type":  strings.Replace(validV3, `"salt":"`+strings.Repeat("0", 64)+`"`, `"salt":1`, 1),
		"malformed number": strings.Replace(validV3, `"n":262144`, `"n":2.5`, 1),
		"unknown field":    strings.Replace(validV3, `"version":3`, `"extra":0,"version":3`, 1),
		"trailing json":    validV3 + `{}`,
		"nested value":     strings.Replace(validV3, `"version":3`, `"version":{"x":{"x":{"x":3}}}`, 1),
		"long string":      strings.Replace(validV3, `"id":"`, `"id":"`+strings.Repeat("x", 300), 1),
		"long document":    validV3 + strings.Repeat(" ", MaxInput),
	}
	for name, document := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := parseEnvelope([]byte(document)); err == nil {
				t.Fatal("accepted noncanonical document")
			}
		})
	}
}

func TestParseEnvelope_rejects_case_variants_for_every_key(t *testing.T) {
	keys := []string{"address", "crypto", "id", "version", "cipher", "cipherparams", "ciphertext", "kdf", "kdfparams", "mac", "iv", "dklen", "n", "p", "r", "salt"}
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			document := strings.Replace(validV3, `"`+key+`"`, `"`+strings.ToUpper(key[:1])+key[1:]+`"`, 1)
			if _, err := parseEnvelope([]byte(document)); err == nil {
				t.Fatal("accepted case-variant key")
			}
		})
	}
}

func TestParseEnvelope_accepts_exact_v3_shape(t *testing.T) {
	metadata, err := parseEnvelope([]byte(validV3))
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Address != strings.Repeat("0", 40) {
		t.Fatalf("got %q", metadata.Address)
	}
}

func TestParseEnvelope_rejects_invalid_utf8_without_panic(t *testing.T) {
	input := append([]byte(validV3[:20]), 0xff)
	input = append(input, []byte(validV3[20:])...)
	if _, err := parseEnvelope(input); err == nil {
		t.Fatal("accepted invalid UTF-8")
	}
}

func TestVerify_rejects_dangerous_shadow_before_decryption(t *testing.T) {
	document := strings.Replace(validV3, `"n":262144`, `"n":1073741824,"N":262144`, 1)
	if _, err := Verify([]byte(document), []byte("TEST ONLY NEVER FUND")); err == nil {
		t.Fatal("accepted dangerous shadow")
	}
}

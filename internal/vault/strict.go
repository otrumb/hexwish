package vault

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/google/uuid"
)

type envelope struct {
	Address string         `json:"address"`
	Crypto  cryptoEnvelope `json:"crypto"`
	ID      string         `json:"id"`
	Version int            `json:"version"`
}

type cryptoEnvelope struct {
	Cipher       string       `json:"cipher"`
	CipherParams cipherParams `json:"cipherparams"`
	Ciphertext   string       `json:"ciphertext"`
	KDF          string       `json:"kdf"`
	KDFParams    kdfParams    `json:"kdfparams"`
	MAC          string       `json:"mac"`
}

type cipherParams struct {
	IV string `json:"iv"`
}

type kdfParams struct {
	DKLen int    `json:"dklen"`
	N     int    `json:"n"`
	P     int    `json:"p"`
	R     int    `json:"r"`
	Salt  string `json:"salt"`
}

type objectSchema map[string]objectSchema

var envelopeSchema = objectSchema{
	"address": nil,
	"crypto": {
		"cipher": nil,
		"cipherparams": {
			"iv": nil,
		},
		"ciphertext": nil,
		"kdf":        nil,
		"kdfparams": {
			"dklen": nil,
			"n":     nil,
			"p":     nil,
			"r":     nil,
			"salt":  nil,
		},
		"mac": nil,
	},
	"id":      nil,
	"version": nil,
}

func parseEnvelope(encoded []byte) (envelope, error) {
	if len(encoded) > MaxInput {
		return envelope{}, errors.New("keystore exceeds 1 MiB")
	}
	if !utf8.Valid(encoded) {
		return envelope{}, errors.New("keystore is not valid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	if err := scanObject(decoder, envelopeSchema); err != nil {
		return envelope{}, fmt.Errorf("parse keystore shape: %w", err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return envelope{}, errors.New("keystore must contain one JSON document")
	}
	var metadata envelope
	if err := json.Unmarshal(encoded, &metadata); err != nil {
		return envelope{}, fmt.Errorf("parse keystore fields: %w", err)
	}
	if err := validateEnvelope(metadata); err != nil {
		return envelope{}, err
	}
	return metadata, nil
}

func scanObject(decoder *json.Decoder, schema objectSchema) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return errors.New("expected object")
	}
	seen := make(map[string]struct{}, len(schema))
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		name, ok := token.(string)
		if !ok {
			return errors.New("object key must be a string")
		}
		child, ok := schema[name]
		if !ok {
			return fmt.Errorf("unknown field %q", name)
		}
		if _, duplicate := seen[name]; duplicate {
			return fmt.Errorf("duplicate field %q", name)
		}
		seen[name] = struct{}{}
		if child != nil {
			if err := scanObject(decoder, child); err != nil {
				return fmt.Errorf("field %q: %w", name, err)
			}
			continue
		}
		value, err := decoder.Token()
		if err != nil {
			return err
		}
		if value == nil {
			return fmt.Errorf("field %q cannot be null", name)
		}
		if text, ok := value.(string); ok && len(text) > 256 {
			return fmt.Errorf("field %q is too long", name)
		}
	}
	if _, err := decoder.Token(); err != nil {
		return err
	}
	if len(seen) != len(schema) {
		return errors.New("object is missing required fields")
	}
	return nil
}

func validateEnvelope(metadata envelope) error {
	if metadata.Version != 3 {
		return errors.New("keystore version must be 3")
	}
	if metadata.Crypto.Cipher != "aes-128-ctr" || metadata.Crypto.KDF != "scrypt" {
		return errors.New("keystore does not use AES-128-CTR with scrypt")
	}
	params := metadata.Crypto.KDFParams
	if params.N != keystore.StandardScryptN || params.R != 8 || params.P != keystore.StandardScryptP || params.DKLen != 32 {
		return errors.New("keystore does not use fixed standard-scrypt profile")
	}
	if err := requireHex("address", metadata.Address, 20); err != nil {
		return err
	}
	if err := requireHex("iv", metadata.Crypto.CipherParams.IV, 16); err != nil {
		return err
	}
	if err := requireHex("ciphertext", metadata.Crypto.Ciphertext, 32); err != nil {
		return err
	}
	if err := requireHex("salt", params.Salt, 32); err != nil {
		return err
	}
	if err := requireHex("mac", metadata.Crypto.MAC, 32); err != nil {
		return err
	}
	parsedID, err := uuid.Parse(metadata.ID)
	if err != nil || parsedID.String() != metadata.ID {
		return errors.New("keystore id must be a canonical UUID")
	}
	return nil
}

func requireHex(name, value string, byteLength int) error {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != byteLength {
		return fmt.Errorf("keystore %s must be %d hexadecimal bytes", name, byteLength)
	}
	return nil
}

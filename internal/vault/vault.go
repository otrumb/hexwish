package vault

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

const MaxInput = 1 << 20

type envelope struct {
	Version int `json:"version"`
	Crypto  struct {
		Cipher    string `json:"cipher"`
		KDF       string `json:"kdf"`
		KDFParams struct {
			N     int `json:"n"`
			R     int `json:"r"`
			P     int `json:"p"`
			DKLen int `json:"dklen"`
		} `json:"kdfparams"`
	} `json:"crypto"`
}

func Encrypt(privateKey *ecdsa.PrivateKey, password []byte) ([]byte, error) {
	key := &keystore.Key{Id: uuid.New(), Address: gethcrypto.PubkeyToAddress(privateKey.PublicKey), PrivateKey: privateKey}
	encoded, err := keystore.EncryptKey(key, string(password), keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, fmt.Errorf("encrypt key: %w", err)
	}
	return encoded, nil
}

func Verify(encoded, password []byte) (common.Address, error) {
	if len(encoded) > MaxInput {
		return common.Address{}, errors.New("keystore exceeds 1 MiB")
	}
	var metadata envelope
	if err := json.Unmarshal(encoded, &metadata); err != nil {
		return common.Address{}, fmt.Errorf("parse keystore metadata: %w", err)
	}
	p := metadata.Crypto.KDFParams
	if metadata.Version != 3 || metadata.Crypto.KDF != "scrypt" || metadata.Crypto.Cipher != "aes-128-ctr" || p.N != keystore.StandardScryptN || p.R != 8 || p.P != keystore.StandardScryptP || p.DKLen != 32 {
		return common.Address{}, errors.New("keystore does not use fixed standard-scrypt profile")
	}
	key, err := keystore.DecryptKey(encoded, string(password))
	if err != nil {
		return common.Address{}, fmt.Errorf("decrypt keystore: %w", err)
	}
	defer key.PrivateKey.D.SetInt64(0)
	derived := gethcrypto.PubkeyToAddress(key.PrivateKey.PublicKey)
	if derived != key.Address {
		return common.Address{}, errors.New("keystore address does not match private key")
	}
	return derived, nil
}

func PersistNoReplace(destination string, data []byte) (err error) {
	dir := filepath.Dir(destination)
	temp, err := os.CreateTemp(dir, ".hexwish-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary keystore: %w", err)
	}
	tempName := temp.Name()
	closed := false
	defer func() {
		if !closed {
			err = errors.Join(err, temp.Close())
		}
		err = errors.Join(err, os.Remove(tempName))
	}()
	if err := temp.Chmod(0o600); err != nil {
		return fmt.Errorf("restrict temporary keystore: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		return fmt.Errorf("write temporary keystore: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("sync temporary keystore: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary keystore: %w", err)
	}
	closed = true
	if err := os.Link(tempName, destination); err != nil {
		return fmt.Errorf("claim destination without overwrite: %w", err)
	}
	return nil
}

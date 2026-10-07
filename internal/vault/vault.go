package vault

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

const MaxInput = 1 << 20

func Encrypt(privateKey *ecdsa.PrivateKey, password []byte) ([]byte, error) {
	key := &keystore.Key{Id: uuid.New(), Address: gethcrypto.PubkeyToAddress(privateKey.PublicKey), PrivateKey: privateKey}
	encoded, err := keystore.EncryptKey(key, string(password), keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, fmt.Errorf("encrypt key: %w", err)
	}
	return encoded, nil
}

func Verify(encoded, password []byte) (common.Address, error) {
	metadata, err := parseEnvelope(encoded)
	if err != nil {
		return common.Address{}, err
	}
	key, err := decryptKey(encoded, password)
	if err != nil {
		return common.Address{}, fmt.Errorf("decrypt keystore: %w", err)
	}
	defer key.PrivateKey.D.SetInt64(0)
	derived := gethcrypto.PubkeyToAddress(key.PrivateKey.PublicKey)
	declared := common.HexToAddress("0x" + strings.ToLower(metadata.Address))
	if derived != key.Address || derived != declared {
		return common.Address{}, errors.New("keystore address does not match private key")
	}
	return derived, nil
}

func decryptKey(encoded, password []byte) (key *keystore.Key, err error) {
	defer func() {
		if recover() != nil {
			key = nil
			err = errors.New("decrypt keystore failed unexpectedly")
		}
	}()
	return keystore.DecryptKey(encoded, string(password))
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

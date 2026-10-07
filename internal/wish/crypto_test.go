package wish

import (
	"encoding/hex"
	"testing"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func TestEthereum_scalar_one_matches_EIP55_vector(t *testing.T) {
	key, err := gethcrypto.ToECDSA(paddedScalar(1))
	if err != nil {
		t.Fatal(err)
	}
	if got := gethcrypto.PubkeyToAddress(key.PublicKey).Hex(); got != "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf" {
		t.Fatalf("got %s", got)
	}
}

func TestEthereum_uses_legacy_Keccak_not_SHA3(t *testing.T) {
	got := gethcrypto.Keccak256(nil)
	if hex.EncodeToString(got) != "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470" {
		t.Fatalf("got %x", got)
	}
}

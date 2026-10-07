package wish

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"strings"
	"sync/atomic"
	"testing"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func TestSearch_returns_one_matching_key(t *testing.T) {
	var calls atomic.Int64
	source := func() (*ecdsa.PrivateKey, error) {
		calls.Add(1)
		return gethcrypto.ToECDSA(paddedScalar(1))
	}
	address := strings.ToLower(gethcrypto.PubkeyToAddress(mustKey(t, paddedScalar(1)).PublicKey).Hex()[2:])
	matcher, err := NewMatcher([]string{address[:4]}, nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Search(context.Background(), matcher, 4, source)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ToLower(result.Address.Hex()[2:]) != address {
		t.Fatalf("got %s", result.Address.Hex())
	}
	if calls.Load() < 1 {
		t.Fatal("source not called")
	}
}

func TestSearch_propagates_entropy_failure(t *testing.T) {
	want := errors.New("entropy failed")
	matcher, _ := NewMatcher([]string{"0"}, nil)
	_, err := Search(context.Background(), matcher, 2, func() (*ecdsa.PrivateKey, error) { return nil, want })
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestSearch_honors_cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	matcher, _ := NewMatcher([]string{"0"}, nil)
	_, err := Search(ctx, matcher, 2, func() (*ecdsa.PrivateKey, error) { t.Fatal("source called"); return nil, nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func paddedScalar(value int64) []byte {
	return new(big.Int).SetInt64(value).FillBytes(make([]byte, 32))
}
func mustKey(t *testing.T, scalar []byte) *ecdsa.PrivateKey {
	t.Helper()
	key, err := gethcrypto.ToECDSA(scalar)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

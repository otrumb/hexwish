package wish

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

type KeySource func() (*ecdsa.PrivateKey, error)

type Result struct {
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
}

func Search(ctx context.Context, matcher Matcher, workers int, source KeySource) (Result, error) {
	if workers < 1 {
		return Result{}, errors.New("workers must be positive")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make(chan Result, 1)
	errorsOut := make(chan error, workers)
	var wait sync.WaitGroup
	wait.Add(workers)
	for range workers {
		go func() {
			defer wait.Done()
			for {
				if err := ctx.Err(); err != nil {
					return
				}
				key, err := source()
				if err != nil {
					select {
					case errorsOut <- err:
						cancel()
					case <-ctx.Done():
					}
					return
				}
				address := gethcrypto.PubkeyToAddress(key.PublicKey)
				if matcher.Match(fmt.Sprintf("%040x", address.Bytes())) {
					select {
					case results <- Result{key, address}:
						cancel()
					case <-ctx.Done():
						key.D.SetInt64(0)
					}
					return
				}
				key.D.SetInt64(0)
			}
		}()
	}
	done := make(chan struct{})
	go func() { wait.Wait(); close(done) }()
	select {
	case result := <-results:
		<-done
		return result, nil
	case err := <-errorsOut:
		<-done
		return Result{}, fmt.Errorf("generate candidate: %w", err)
	case <-ctx.Done():
		<-done
		select {
		case result := <-results:
			return result, nil
		default:
		}
		select {
		case err := <-errorsOut:
			return Result{}, fmt.Errorf("generate candidate: %w", err)
		default:
			return Result{}, ctx.Err()
		}
	}
}

func GenerateKey() (*ecdsa.PrivateKey, error) { return gethcrypto.GenerateKey() }

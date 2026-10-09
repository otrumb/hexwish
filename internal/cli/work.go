package cli

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/otrumb/hexwish/internal/vault"
	"github.com/otrumb/hexwish/internal/wish"
)

func benchmark(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	set := flag.NewFlagSet("benchmark", flag.ContinueOnError)
	set.SetOutput(stderr)
	duration := set.Duration("duration", time.Second, "bounded duration")
	if err := set.Parse(args); err != nil {
		return err
	}
	if *duration <= 0 || *duration > time.Minute {
		return errors.New("duration must be within (0,1m]")
	}
	matcher, _ := wish.NewMatcher([]string{"f"}, nil)
	deadline := time.Now().Add(*duration)
	var count uint64
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return err
		}
		key, err := wish.GenerateKey()
		if err != nil {
			return err
		}
		matcher.Match(strings.ToLower(fmt.Sprintf("%040x", gethcrypto.PubkeyToAddress(key.PublicKey).Bytes())))
		key.D.SetInt64(0)
		count++
	}
	fmt.Fprintf(stdout, "candidates=%d rate=%.0f/s\n", count, float64(count)/duration.Seconds())
	return nil
}

func generate(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	set := flag.NewFlagSet("generate", flag.ContinueOnError)
	set.SetOutput(stderr)
	var prefixes, suffixes patterns
	set.Var(&prefixes, "prefix", "lowercase hex prefix; repeatable")
	set.Var(&suffixes, "suffix", "lowercase hex suffix; repeatable")
	workers := set.Int("workers", runtime.NumCPU(), "candidate workers")
	output := set.String("output", "", "new keystore destination")
	confirm := set.Bool("confirm", false, "confirm generation")
	if err := set.Parse(args); err != nil {
		return err
	}
	if *output == "" || !*confirm {
		return errors.New("--output and --confirm are required")
	}
	matcher, err := wish.NewMatcher(prefixes, suffixes)
	if err != nil {
		return err
	}
	result, err := wish.Search(ctx, matcher, *workers, wish.GenerateKey)
	if err != nil {
		return err
	}
	defer result.PrivateKey.D.SetInt64(0)
	password, err := readPassword("New password: ", stderr)
	if err != nil {
		return err
	}
	defer clear(password)
	again, err := readPassword("Confirm password: ", stderr)
	if err != nil {
		return err
	}
	defer clear(again)
	if len(password) == 0 || !bytes.Equal(password, again) {
		return errors.New("passwords do not match or are empty")
	}
	encoded, err := vault.Encrypt(result.PrivateKey, password)
	if err != nil {
		return err
	}
	verified, err := vault.Verify(encoded, password)
	if err != nil {
		return err
	}
	if verified != result.Address {
		return errors.New("keystore verification mismatch")
	}
	if err := vault.PersistNoReplace(*output, encoded); err != nil {
		return err
	}
	fmt.Fprintln(stdout, result.Address.Hex())
	return nil
}

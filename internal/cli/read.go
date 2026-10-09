package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/otrumb/hexwish/internal/vault"
	"github.com/otrumb/hexwish/internal/wish"
)

func estimate(args []string, stdout, stderr io.Writer) error {
	set := flag.NewFlagSet("estimate", flag.ContinueOnError)
	set.SetOutput(stderr)
	var prefixes, suffixes patterns
	var rate float64
	set.Var(&prefixes, "prefix", "lowercase hex prefix; repeatable")
	set.Var(&suffixes, "suffix", "lowercase hex suffix; repeatable")
	set.Float64Var(&rate, "rate", 1, "candidates per second")
	if err := set.Parse(args); err != nil {
		return err
	}
	if rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) {
		return errors.New("rate must be finite and positive")
	}
	matcher, err := wish.NewMatcher(prefixes, suffixes)
	if err != nil {
		return err
	}
	p := matcher.Probability(40)
	fmt.Fprintf(stdout, "probability=%s expected_trials=%s\n", p.RatString(), new(big.Rat).Inv(p).FloatString(0))
	for _, q := range []float64{0.5, 0.9, 0.99} {
		trials, err := wish.QuantileTrials(p, q)
		if err != nil {
			fmt.Fprintf(stdout, "p%.0f_trials=unrepresentable eta=unavailable\n", q*100)
			continue
		}
		seconds := float64(trials) / rate
		if seconds > float64(time.Duration(1<<63-1))/float64(time.Second) {
			fmt.Fprintf(stdout, "p%.0f_trials=%d eta=unavailable\n", q*100, trials)
			continue
		}
		fmt.Fprintf(stdout, "p%.0f_trials=%d eta=%s\n", q*100, trials, time.Duration(seconds*float64(time.Second)).Round(time.Second))
	}
	return nil
}

func verify(args []string, stdout, stderr io.Writer) error {
	set := flag.NewFlagSet("verify", flag.ContinueOnError)
	set.SetOutput(stderr)
	file := set.String("file", "", "Web3 V3 keystore path")
	var prefixes, suffixes patterns
	set.Var(&prefixes, "prefix", "lowercase hex prefix; repeatable")
	set.Var(&suffixes, "suffix", "lowercase hex suffix; repeatable")
	if err := set.Parse(args); err != nil {
		return err
	}
	if *file == "" {
		return errors.New("--file is required")
	}
	info, err := os.Stat(*file)
	if err != nil {
		return fmt.Errorf("stat keystore: %w", err)
	}
	if info.Size() > vault.MaxInput {
		return errors.New("keystore exceeds 1 MiB")
	}
	encoded, err := os.ReadFile(*file)
	if err != nil {
		return fmt.Errorf("read keystore: %w", err)
	}
	password, err := readPassword("Password: ", stderr)
	if err != nil {
		return err
	}
	defer clear(password)
	address, err := vault.Verify(encoded, password)
	if err != nil {
		return err
	}
	if len(prefixes)+len(suffixes) > 0 {
		matcher, err := wish.NewMatcher(prefixes, suffixes)
		if err != nil {
			return err
		}
		if !matcher.Match(strings.ToLower(fmt.Sprintf("%040x", address.Bytes()))) {
			return errors.New("address does not match requested pattern")
		}
	}
	fmt.Fprintln(stdout, address.Hex())
	return nil
}

package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var readPassword = terminalPassword

const Version = "0.1.1"

type patterns []string

func (p *patterns) String() string         { return strings.Join(*p, ",") }
func (p *patterns) Set(value string) error { *p = append(*p, value); return nil }

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		help(stdout)
		return nil
	}
	switch args[0] {
	case "version":
		fmt.Fprintf(stdout, "hexwish %s\n", Version)
		return nil
	case "estimate":
		return estimate(args[1:], stdout, stderr)
	case "benchmark":
		return benchmark(ctx, args[1:], stdout, stderr)
	case "verify":
		return verify(args[1:], stdout, stderr)
	case "generate":
		return generate(ctx, args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func help(out io.Writer) {
	fmt.Fprintln(out, "hexwish - local EVM vanity address generator")
	fmt.Fprintln(out, "commands: generate, estimate, verify, benchmark, version, help")
}

func terminalPassword(prompt string, stderr io.Writer) ([]byte, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, errors.New("password requires an interactive terminal")
	}
	fmt.Fprint(stderr, prompt)
	value, err := term.ReadPassword(fd)
	fmt.Fprintln(stderr)
	if err != nil {
		return nil, fmt.Errorf("read hidden password: %w", err)
	}
	return value, nil
}

func clear(value []byte) {
	for i := range value {
		value[i] = 0
	}
}

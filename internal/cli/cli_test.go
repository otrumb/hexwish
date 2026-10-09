package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_help_lists_all_commands(t *testing.T) {
	var out bytes.Buffer
	if err := Run(context.Background(), []string{"help"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"generate", "estimate", "verify", "benchmark", "version", "help"} {
		if !strings.Contains(out.String(), command) {
			t.Fatalf("missing %s", command)
		}
	}
}

func TestRun_estimate_reports_exact_probability(t *testing.T) {
	var out bytes.Buffer
	err := Run(context.Background(), []string{"estimate", "--prefix", "a", "--suffix", "b", "--rate", "1000"}, &out, &out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "probability=1/256") {
		t.Fatalf("got %s", out.String())
	}
}

func TestRun_rejects_unknown_command(t *testing.T) {
	if err := Run(context.Background(), []string{"unknown"}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("accepted unknown command")
	}
}

func TestRun_estimate_never_reports_wrapped_eta(t *testing.T) {
	var out bytes.Buffer
	err := Run(context.Background(), []string{"estimate", "--prefix", "000000000000000", "--rate", "1"}, &out, &out)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "eta=-") {
		t.Fatalf("wrapped ETA: %s", out.String())
	}
}

func TestRun_estimate_rejects_nonfinite_rate(t *testing.T) {
	for _, rate := range []string{"NaN", "+Inf", "-Inf"} {
		err := Run(context.Background(), []string{"estimate", "--prefix", "a", "--rate", rate}, &bytes.Buffer{}, &bytes.Buffer{})
		if err == nil {
			t.Fatalf("accepted rate %s", rate)
		}
	}
}

func TestRun_generate_then_verify_uses_hidden_password_boundary(t *testing.T) {
	password := []byte("TEST ONLY NEVER FUND")
	oldReader := readPassword
	var prompts []string
	readPassword = func(prompt string, _ io.Writer) ([]byte, error) {
		prompts = append(prompts, prompt)
		return append([]byte(nil), password...), nil
	}
	t.Cleanup(func() { readPassword = oldReader })

	output := filepath.Join(t.TempDir(), "key.json")
	var generateOut, verifyOut, stderr bytes.Buffer
	if err := Run(context.Background(), []string{"generate", "--prefix", "0", "--workers", "1", "--output", output, "--confirm"}, &generateOut, &stderr); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatal(err)
	}
	if err := Run(context.Background(), []string{"verify", "--file", output, "--prefix", "0"}, &verifyOut, &stderr); err != nil {
		t.Fatal(err)
	}
	if generateOut.String() != verifyOut.String() {
		t.Fatalf("generated %q verified %q", generateOut.String(), verifyOut.String())
	}
	if got := strings.Join(prompts, "|"); got != "New password: |Confirm password: |Password: " {
		t.Fatalf("prompts %s", got)
	}
}

func TestRun_generate_wrong_confirmation_leaves_no_destination(t *testing.T) {
	oldReader := readPassword
	answers := [][]byte{[]byte("first"), []byte("second")}
	readPassword = func(_ string, _ io.Writer) ([]byte, error) {
		answer := answers[0]
		answers = answers[1:]
		return append([]byte(nil), answer...), nil
	}
	t.Cleanup(func() { readPassword = oldReader })

	output := filepath.Join(t.TempDir(), "key.json")
	err := Run(context.Background(), []string{"generate", "--prefix", "0", "--workers", "1", "--output", output, "--confirm"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("accepted mismatched passwords")
	}
	if _, statErr := os.Stat(output); !os.IsNotExist(statErr) {
		t.Fatalf("destination exists after mismatch: %v", statErr)
	}
}

func TestCLI_exposes_no_secret_flag_or_environment_input(t *testing.T) {
	for _, name := range []string{"root.go", "read.go", "work.go"} {
		content, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		for _, forbidden := range []string{"String(\"password\"", "StringVar(\"password\"", "LookupEnv(", "Getenv("} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s exposes forbidden secret input %q", name, forbidden)
			}
		}
	}
}

package cli

import (
	"bytes"
	"context"
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

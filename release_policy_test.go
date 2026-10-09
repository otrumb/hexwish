package main

import (
	"os"
	"strings"
	"testing"
)

func TestCI_runs_for_main_pull_requests_and_manual_dispatch(t *testing.T) {
	workflow := readPolicyFile(t, ".github/workflows/ci.yml")
	for _, required := range []string{"push:", "branches: [main]", "pull_request:", "workflow_dispatch:", "contents: read"} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("missing CI policy token %q", required)
		}
	}
}

func TestRelease_requires_exact_v011_tag_and_dual_os_validation(t *testing.T) {
	workflow := readPolicyFile(t, ".github/workflows/release.yml")
	for _, required := range []string{"tags: [v0.1.1]", "windows-latest", "ubuntu-latest", "needs: validate", "draft: true", "contents: write"} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("missing release policy token %q", required)
		}
	}
	if strings.Count(workflow, "contents: write") != 1 {
		t.Fatal("write permission must exist only on final release job")
	}
}

func TestRelease_binds_remote_tag_to_validated_commit(t *testing.T) {
	workflow := readPolicyFile(t, ".github/workflows/release.yml")
	for _, required := range []string{"group: release-${{ github.ref }}", "cancel-in-progress: false", "git ls-remote origin", "test \"${remote_target}\" = \"${GITHUB_SHA}\""} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("missing release binding token %q", required)
		}
	}
	if strings.Index(workflow, "git ls-remote origin") > strings.Index(workflow, "softprops/action-gh-release") {
		t.Fatal("release is created before remote tag verification")
	}
}

func TestRelease_uses_full_action_pins_and_local_packager(t *testing.T) {
	workflow := readPolicyFile(t, ".github/workflows/release.yml")
	for line := range strings.SplitSeq(workflow, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- uses:") {
			_, pin, found := strings.Cut(line, "@")
			if !found || len(pin) != 40 {
				t.Fatalf("action lacks full SHA pin: %s", line)
			}
		}
	}
	if !strings.Contains(workflow, "go run ./cmd/package") || !strings.Contains(workflow, "-out dist") {
		t.Fatal("release does not use locally rehearsable packager")
	}
	for _, required := range []string{"-platform source", "hexwish-source-v0.1.1.tar.gz"} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("release does not publish relinking source: %s", required)
		}
	}
}

func readPolicyFile(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

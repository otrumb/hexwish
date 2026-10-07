package wish

import (
	"strings"
	"testing"
)

func TestMatcher_combines_sides_with_and(t *testing.T) {
	m, err := NewMatcher([]string{"ab", "cd"}, []string{"ef"})
	if err != nil {
		t.Fatal(err)
	}
	if !m.Match("ab" + strings.Repeat("0", 36) + "ef") {
		t.Fatal("expected match")
	}
	if m.Match("ab" + strings.Repeat("0", 38)) {
		t.Fatal("suffix must match")
	}
}

func TestMatcher_ors_patterns_on_each_side(t *testing.T) {
	m, err := NewMatcher([]string{"ab", "cd"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !m.Match("cd" + strings.Repeat("0", 38)) {
		t.Fatal("second prefix must match")
	}
}

func TestMatcher_rejects_invalid_or_unbounded_input(t *testing.T) {
	for _, value := range []string{"0x", "AB", "g", "12345678901234567890123456789012345678901"} {
		if _, err := NewMatcher([]string{value}, nil); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
}

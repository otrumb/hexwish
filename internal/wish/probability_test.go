package wish

import (
	"math/big"
	"testing"
)

func TestFavorable_counts_overlap_compatibility(t *testing.T) {
	m, err := NewMatcher([]string{"ab"}, []string{"bc"})
	if err != nil {
		t.Fatal(err)
	}
	got := m.Favorable(3)
	if got.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("got %s", got)
	}
}

func TestProbability_matches_exhaustive_reduced_space(t *testing.T) {
	m, err := NewMatcher([]string{"a", "b"}, []string{"0", "1"})
	if err != nil {
		t.Fatal(err)
	}
	want := new(big.Rat).SetFrac(big.NewInt(4), big.NewInt(256))
	if m.Probability(2).Cmp(want) != 0 {
		t.Fatalf("got %s want %s", m.Probability(2), want)
	}
}

func TestQuantile_is_honest_geometric_trial_count(t *testing.T) {
	p := big.NewRat(1, 16)
	got, err := QuantileTrials(p, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if got != 11 {
		t.Fatalf("got %d", got)
	}
}

func TestQuantile_rejects_unrepresentable_trial_count(t *testing.T) {
	if _, err := QuantileTrials(new(big.Rat).SetFrac(big.NewInt(1), pow16(40)), 0.99); err == nil {
		t.Fatal("accepted overflowing trial count")
	}
}

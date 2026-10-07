package wish

import (
	"errors"
	"math"
	"math/big"
)

func (m Matcher) Favorable(width int) *big.Int {
	if width < 1 {
		return new(big.Int)
	}
	if len(m.prefixes) == 0 {
		return cylinders(m.suffixes, width)
	}
	if len(m.suffixes) == 0 {
		return cylinders(m.prefixes, width)
	}
	total := new(big.Int)
	for _, prefix := range m.prefixes {
		for _, suffix := range m.suffixes {
			if compatible(prefix, suffix, width) {
				fixed := len(prefix) + len(suffix)
				if fixed > width {
					fixed = width
				}
				total.Add(total, pow16(width-fixed))
			}
		}
	}
	return total
}

func cylinders(patterns []string, width int) *big.Int {
	total := new(big.Int)
	for _, pattern := range patterns {
		if len(pattern) <= width {
			total.Add(total, pow16(width-len(pattern)))
		}
	}
	return total
}

func compatible(prefix, suffix string, width int) bool {
	if len(prefix) > width || len(suffix) > width {
		return false
	}
	overlap := len(prefix) + len(suffix) - width
	return overlap <= 0 || prefix[len(prefix)-overlap:] == suffix[:overlap]
}

func pow16(power int) *big.Int {
	return new(big.Int).Exp(big.NewInt(16), big.NewInt(int64(power)), nil)
}

func (m Matcher) Probability(width int) *big.Rat {
	return new(big.Rat).SetFrac(m.Favorable(width), pow16(width))
}

func QuantileTrials(probability *big.Rat, quantile float64) (uint64, error) {
	p, _ := probability.Float64()
	if p <= 0 || quantile <= 0 || quantile >= 1 {
		return 0, errors.New("probability and quantile must be within (0,1)")
	}
	if p >= 1 {
		return 1, nil
	}
	trials := math.Ceil(math.Log1p(-quantile) / math.Log1p(-p))
	if math.IsInf(trials, 0) || trials > math.MaxUint64 {
		return 0, errors.New("quantile exceeds uint64 trial range")
	}
	return uint64(trials), nil
}

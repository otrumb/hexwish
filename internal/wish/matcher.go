package wish

import (
	"errors"
	"strings"
)

var ErrPattern = errors.New("pattern must be 1-40 lowercase hexadecimal characters")

type Matcher struct{ prefixes, suffixes []string }

func NewMatcher(prefixes, suffixes []string) (Matcher, error) {
	p, err := normalize(prefixes, true)
	if err != nil {
		return Matcher{}, err
	}
	s, err := normalize(suffixes, false)
	if err != nil {
		return Matcher{}, err
	}
	if len(p) == 0 && len(s) == 0 {
		return Matcher{}, errors.New("at least one pattern is required")
	}
	if len(p) > 0 && len(s) > 0 {
		possible := false
		for _, prefix := range p {
			for _, suffix := range s {
				possible = possible || compatible(prefix, suffix, 40)
			}
		}
		if !possible {
			return Matcher{}, errors.New("prefix and suffix patterns cannot match one address")
		}
	}
	return Matcher{prefixes: p, suffixes: s}, nil
}

func normalize(values []string, prefix bool) ([]string, error) {
	var out []string
	for _, value := range values {
		if len(value) == 0 || len(value) > 40 {
			return nil, ErrPattern
		}
		for _, c := range value {
			if !strings.ContainsRune("0123456789abcdef", c) {
				return nil, ErrPattern
			}
		}
		subsumed := false
		for i := 0; i < len(out); {
			contains := strings.HasPrefix
			if !prefix {
				contains = strings.HasSuffix
			}
			if contains(value, out[i]) {
				subsumed = true
				break
			}
			if contains(out[i], value) {
				out = append(out[:i], out[i+1:]...)
				continue
			}
			i++
		}
		if !subsumed {
			out = append(out, value)
		}
	}
	return out, nil
}

func (m Matcher) Match(address string) bool {
	if len(address) != 40 {
		return false
	}
	prefix := len(m.prefixes) == 0
	for _, pattern := range m.prefixes {
		prefix = prefix || strings.HasPrefix(address, pattern)
	}
	suffix := len(m.suffixes) == 0
	for _, pattern := range m.suffixes {
		suffix = suffix || strings.HasSuffix(address, pattern)
	}
	return prefix && suffix
}

func (m Matcher) Prefixes() []string { return append([]string(nil), m.prefixes...) }
func (m Matcher) Suffixes() []string { return append([]string(nil), m.suffixes...) }

package search

import (
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/sahilm/fuzzy"
)

type MatchMode int

const (
	ModeFuzzy MatchMode = iota
	ModeExact
	ModeGlob
	ModeRegex
)

type Matcher struct {
	mode          MatchMode
	pattern       string
	caseSensitive bool
	re            *regexp.Regexp
}

func NewMatcher(pattern string, mode MatchMode, caseSensitive bool) (*Matcher, error) {
	m := &Matcher{mode: mode, pattern: pattern, caseSensitive: caseSensitive}
	if !caseSensitive && mode != ModeGlob {
		m.pattern = strings.ToLower(pattern)
	}
	if mode == ModeRegex {
		flags := "(?i)"
		if caseSensitive {
			flags = ""
		}
		re, err := regexp.Compile(flags + pattern)
		if err != nil {
			return nil, err
		}
		m.re = re
	}
	return m, nil
}

func (m *Matcher) Match(name string) (bool, int) {
	target := name
	if !m.caseSensitive && m.mode != ModeGlob {
		target = strings.ToLower(name)
	}
	switch m.mode {
	case ModeExact:
		return strings.Contains(target, m.pattern), 50
	case ModeGlob:
		pat := m.pattern
		if !m.caseSensitive {
			pat = strings.ToLower(pat)
			target = strings.ToLower(target)
		}
		ok, _ := doublestar.Match(pat, target)
		return ok, 100
	case ModeRegex:
		return m.re.MatchString(name), 80
	default: // ModeFuzzy
		matches := fuzzy.Find(m.pattern, []string{target})
		if len(matches) == 0 {
			return false, 0
		}
		return true, matches[0].Score
	}
}

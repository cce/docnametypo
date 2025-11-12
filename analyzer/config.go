package analyzer

import (
	"slices"
	"strings"
)

type matchConfig struct {
	allowedLeadingWords map[string]struct{}
	allowedPrefixes     []string
}

// newMatchConfig builds the configuration used for doc/token comparisons.
func newMatchConfig() matchConfig {
	return matchConfig{
		allowedLeadingWords: buildAllowedLeadingWords(allowedLeadingWordsFlag),
		allowedPrefixes:     splitCSV(allowedPrefixesFlag),
	}
}

// isAllowedLeadingWord reports whether the token is in the narrative word list.
func (c matchConfig) isAllowedLeadingWord(word string) bool {
	if word == "" || len(c.allowedLeadingWords) == 0 {
		return false
	}

	_, ok := c.allowedLeadingWords[strings.ToLower(word)]

	return ok
}

// matchesAllowedPrefixVariant checks if removing a configured prefix yields a match.
func (c matchConfig) matchesAllowedPrefixVariant(docToken, symbol string) bool {
	if len(c.allowedPrefixes) == 0 {
		return false
	}

	symbolLower := strings.ToLower(symbol)

	return slices.ContainsFunc(c.allowedPrefixes, func(rawPrefix string) bool {
		prefix := strings.TrimSpace(rawPrefix)
		if prefix == "" || len(symbol) <= len(prefix) {
			return false
		}

		if !strings.HasPrefix(symbolLower, strings.ToLower(prefix)) {
			return false
		}

		trimmed := symbol[len(prefix):]

		return trimmed != "" && strings.EqualFold(docToken, trimmed)
	})
}

// buildAllowedLeadingWords normalizes the CSV list of narrative words.
func buildAllowedLeadingWords(raw string) map[string]struct{} {
	words := make(map[string]struct{})

	for _, w := range splitCSV(raw) {
		if w == "" {
			continue
		}

		words[strings.ToLower(w)] = struct{}{}
	}

	return words
}

// splitCSV splits a comma/whitespace separated list and trims empties.
func splitCSV(raw string) []string {
	if raw == "" {
		return nil
	}

	fields := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', ';', '/', '\n', '\t', ' ':
			return true
		}

		return false
	})

	return fields
}

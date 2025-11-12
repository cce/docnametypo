package analyzer

import (
	"strings"
	"unicode"
)

var sectionHeaderSecondWords = map[string]struct{}{
	"helper":   {},
	"helpers":  {},
	"section":  {},
	"sections": {},
	"overview": {},
	"summary":  {},
}

var narrativeSecondWords = map[string]struct{}{
	"that":    {},
	"the":     {},
	"a":       {},
	"an":      {},
	"this":    {},
	"these":   {},
	"those":   {},
	"whether": {},
	"if":      {},
}

// isSectionHeader reports whether the doc line looks like a heading.
func isSectionHeader(firstTok, line string) bool {
	if firstTok == "" || line == "" {
		return false
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return false
	}

	first := stripWordToken(fields[0])
	if !strings.EqualFold(firstTok, first) {
		return false
	}

	second := strings.ToLower(stripWordToken(fields[1]))
	if second == "" {
		return false
	}

	_, ok := sectionHeaderSecondWords[second]
	return ok
}

// isNarrativeSentenceIntro detects natural-language sentences.
func isNarrativeSentenceIntro(firstTok, line string) bool {
	if !looksLikeSimpleWord(firstTok) || line == "" {
		return false
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return false
	}

	first := stripWordToken(fields[0])
	if !strings.EqualFold(firstTok, first) {
		return false
	}

	second := strings.ToLower(stripWordToken(fields[1]))
	if second == "" {
		return false
	}

	_, ok := narrativeSecondWords[second]
	return ok
}

// containsWildcardToken returns true if the token is clearly generic.
func containsWildcardToken(token, line string) bool {
	if strings.ContainsAny(token, "*?[]") {
		return true
	}
	if token == "" || line == "" {
		return false
	}

	lowerLine := strings.ToLower(line)
	lowerToken := strings.ToLower(token)
	if rest, ok := strings.CutPrefix(lowerLine, lowerToken); ok && rest != "" {
		return strings.HasPrefix(rest, "*")
	}
	return false
}

// looksLikeSimpleWord reports whether the token is a single plain word.
func looksLikeSimpleWord(word string) bool {
	if word == "" {
		return false
	}
	runes := []rune(word)
	for _, r := range runes {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	if len(runes) == 1 {
		return true
	}
	rest := strings.ToLower(string(runes[1:]))
	if rest != string(runes[1:]) {
		return false
	}
	return unicode.IsLower(runes[0]) || unicode.IsUpper(runes[0])
}

// hasCamelCaseInterior reports whether a name contains inner capitals.
func hasCamelCaseInterior(name string) bool {
	for i, r := range name {
		if unicode.IsUpper(r) && i > 0 {
			return true
		}
	}
	return false
}

// stripWordToken removes punctuation from both ends of a token.
func stripWordToken(word string) string {
	return strings.Trim(word, " \t:.,;\r\n-*")
}

// docFirstWordHasDot detects package-qualified references like json.Marshal.
func docFirstWordHasDot(line string) bool {
	if line == "" {
		return false
	}
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}
	return strings.Contains(fields[0], ".")
}

// isNarrativeVerbForm detects verbs like "Creates" when the symbol starts similarly.
func isNarrativeVerbForm(word, funcName string) bool {
	if len(word) < 2 {
		return false
	}
	stem, ok := strings.CutSuffix(strings.ToLower(word), "s")
	if !ok || stem == "" {
		return false
	}
	return strings.HasPrefix(strings.ToLower(funcName), stem)
}

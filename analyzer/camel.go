package analyzer

import (
	"strings"
	"unicode"
)

// hasCamelChunkReplacement allows a limited number of camel chunks to differ.
func hasCamelChunkReplacement(docToken, symbol string, maxMismatch int) bool {
	if maxMismatch <= 0 {
		return false
	}

	docWords := splitCamelWords(docToken)
	symWords := splitCamelWords(symbol)
	if len(docWords) == 0 || len(docWords) != len(symWords) {
		return false
	}
	if len(docWords) < 2 {
		return false
	}

	mismatches := 0
	matches := 0
	for i := range docWords {
		if docWords[i] == symWords[i] {
			matches++
			continue
		}
		mismatches++
		if mismatches > maxMismatch {
			return false
		}
	}

	if mismatches == 0 {
		return false
	}
	return matches >= len(docWords)-maxMismatch && matches > 0
}

// hasCamelChunkInsertionOrRemoval tolerates inserted or removed camel chunks.
func hasCamelChunkInsertionOrRemoval(docToken, symbol string, maxChunkDiff int) bool {
	if maxChunkDiff <= 0 {
		return false
	}

	docWords := splitCamelWords(docToken)
	symWords := splitCamelWords(symbol)
	if len(docWords) == 0 || len(symWords) == 0 {
		return false
	}

	diff := abs(len(docWords) - len(symWords))
	if diff == 0 || diff > maxChunkDiff {
		return false
	}

	if len(docWords) > len(symWords) {
		return camelSubsequence(symWords, docWords, maxChunkDiff)
	}
	return camelSubsequence(docWords, symWords, maxChunkDiff)
}

// camelSubsequence checks if the shorter chunk list appears in order.
func camelSubsequence(shorter, longer []string, maxSkips int) bool {
	if len(shorter) == 0 || len(longer) == 0 {
		return false
	}

	i, j := 0, 0
	skips := 0
	for i < len(shorter) && j < len(longer) {
		if shorter[i] == longer[j] {
			i++
			j++
			continue
		}
		j++
		skips++
		if skips > maxSkips {
			return false
		}
	}
	return i == len(shorter) && len(shorter) > 0
}

// isCamelSwapVariant detects swapped adjacent camel chunks.
func isCamelSwapVariant(docToken, symbol string) bool {
	docWords := splitCamelWords(docToken)
	symWords := splitCamelWords(symbol)
	if len(docWords) != len(symWords) || len(docWords) < 2 {
		return false
	}

	var diffs [2]int
	diffCount := 0
	for i := range docWords {
		if docWords[i] == symWords[i] {
			continue
		}
		if diffCount == len(diffs) {
			return false
		}
		diffs[diffCount] = i
		diffCount++
	}

	if diffCount != 2 {
		return false
	}
	i, j := diffs[0], diffs[1]
	return docWords[i] == symWords[j] && docWords[j] == symWords[i]
}

// hasSimilarCamelWord allows a single camel chunk to be a close typo.
func hasSimilarCamelWord(docToken, symbol string) bool {
	docWords := splitCamelWords(docToken)
	symWords := splitCamelWords(symbol)
	if len(docWords) == 0 || len(docWords) != len(symWords) {
		return false
	}

	mismatches := 0
	for i := range docWords {
		if docWords[i] == symWords[i] {
			continue
		}
		if !wordClose(docWords[i], symWords[i]) {
			return false
		}
		mismatches++
		if mismatches > 1 {
			return false
		}
	}
	return mismatches > 0
}

// wordClose reports whether two words are similar under distance heuristics.
func wordClose(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	al := strings.ToLower(a)
	bl := strings.ToLower(b)
	if al == bl {
		return true
	}

	dist := damerauLevenshtein(al, bl)
	if dist > maxDistFlag+1 {
		return false
	}

	minLen := min(len(al), len(bl))
	if minLen <= 1 {
		return false
	}

	threshold := minLen - 2
	if threshold < 2 {
		threshold = minLen
	}
	prefix := commonPrefixLength(al, bl)
	suffix := commonSuffixLength(al, bl)
	return prefix >= threshold || suffix >= threshold
}

// hasSmallChunkDifference allows a small suffix/prefix chunk variance.
func hasSmallChunkDifference(a, b string, maxChunk int) bool {
	if maxChunk <= 0 {
		return false
	}
	if len(a) == len(b) {
		return false
	}
	if len(a) < len(b) {
		return hasSmallChunkDifference(b, a, maxChunk)
	}

	diff := len(a) - len(b)
	if diff > maxChunk {
		return false
	}
	for i := 0; i <= len(a)-diff; i++ {
		if strings.HasPrefix(b, a[:i]) && strings.HasSuffix(b, a[i+diff:]) {
			return true
		}
	}
	return false
}

// splitCamelWords tokenizes a camelCase or snake_case identifier.
func splitCamelWords(s string) []string {
	s = strings.ReplaceAll(s, "_", "")
	if s == "" {
		return nil
	}
	runes := []rune(s)

	var words []string
	start := 0
	for i := 1; i < len(runes); i++ {
		if camelWordBoundary(runes, i) {
			if w := strings.ToLower(string(runes[start:i])); w != "" {
				words = append(words, w)
			}
			start = i
		}
	}
	if start < len(runes) {
		if w := strings.ToLower(string(runes[start:])); w != "" {
			words = append(words, w)
		}
	}
	return words
}

// camelWordBoundary reports whether a boundary occurs before index idx.
func camelWordBoundary(runes []rune, idx int) bool {
	prev := runes[idx-1]
	curr := runes[idx]

	if unicode.IsDigit(prev) != unicode.IsDigit(curr) {
		return true
	}
	if unicode.IsLetter(prev) != unicode.IsLetter(curr) {
		return true
	}
	if unicode.IsLower(prev) && unicode.IsUpper(curr) {
		return true
	}
	if unicode.IsUpper(prev) && unicode.IsUpper(curr) {
		return idx+1 < len(runes) && unicode.IsLower(runes[idx+1])
	}
	return false
}

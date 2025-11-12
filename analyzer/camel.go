package analyzer

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/cce/docnametypo/analyzer/internal/camelcase"
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
	closeEnough := slices.EqualFunc(docWords, symWords, func(a, b string) bool {
		if a == b {
			matches++
			return true
		}
		mismatches++
		return mismatches <= maxMismatch
	})
	if !closeEnough || mismatches == 0 {
		return false
	}
	return matches > 0 && matches >= len(docWords)-maxMismatch
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

	skips := 0
	pos := 0
	for _, chunk := range shorter {
		idx := slices.Index(longer[pos:], chunk)
		if idx == -1 {
			return false
		}
		skips += idx
		if skips > maxSkips {
			return false
		}
		pos += idx + 1
	}
	return len(shorter) > 0
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
	closeEnough := slices.EqualFunc(docWords, symWords, func(a, b string) bool {
		if a == b {
			return true
		}
		if mismatches == 1 || !wordClose(a, b) {
			return false
		}
		mismatches++
		return true
	})
	return closeEnough && mismatches == 1
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
	if !utf8.ValidString(s) {
		return []string{strings.ToLower(s)}
	}

	rawParts := camelcase.Split(s)
	if len(rawParts) == 0 {
		return []string{strings.ToLower(s)}
	}
	rawParts = slices.DeleteFunc(rawParts, func(part string) bool { return part == "" })
	if len(rawParts) == 0 {
		return []string{strings.ToLower(s)}
	}

	words := make([]string, 0, len(rawParts))
	for i := 0; i < len(rawParts); i++ {
		part := rawParts[i]
		if i+1 < len(rawParts) && shouldMergeCamelParts(part, rawParts[i+1]) {
			part += rawParts[i+1]
			i++
		}
		words = append(words, strings.ToLower(part))
	}
	return words
}

func shouldMergeCamelParts(a, b string) bool {
	if len(a) != 1 {
		return false
	}
	rA, _ := utf8.DecodeRuneInString(a)
	if !unicode.IsUpper(rA) {
		return false
	}
	if len(b) < 2 {
		return false
	}
	rB, size := utf8.DecodeRuneInString(b)
	if !unicode.IsUpper(rB) {
		return false
	}
	rest := b[size:]
	if rest == "" {
		return false
	}
	for len(rest) > 0 {
		r, s := utf8.DecodeRuneInString(rest)
		if unicode.IsUpper(r) {
			return false
		}
		rest = rest[s:]
	}
	return true
}

package analyzer

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode/utf8"
)

// firstIdentifierLike extracts the first identifier-looking token from the first
// non-empty line of a comment group. It also returns the token range so a
// SuggestedFix can rewrite it in-place, plus the trimmed first line for
// downstream heuristics.
func firstIdentifierLike(cg *ast.CommentGroup) (string, token.Pos, token.Pos, string) {
	if cg == nil || len(cg.List) == 0 {
		return "", token.NoPos, token.NoPos, ""
	}

	comment := cg.List[0]

	line, lineOffset := firstDocLine(comment.Text)
	if line == "" {
		return "", token.NoPos, token.NoPos, ""
	}

	id, rel := identifierFromLine(line)
	if id == "" {
		return "", token.NoPos, token.NoPos, line
	}

	start := comment.Slash + token.Pos(lineOffset+rel)
	end := start + token.Pos(len(id))

	return id, start, end, line
}

// firstDocLine returns the first non-empty line of the raw comment text.
func firstDocLine(raw string) (string, int) {
	if raw == "" {
		return "", 0
	}

	text := raw

	var consumed int

	if trimmed, ok := strings.CutPrefix(text, "//"); ok {
		text = trimmed
		consumed += 2
	} else if trimmed, ok := strings.CutPrefix(text, "/*"); ok {
		text = trimmed
		consumed += 2

		if withoutSuffix, ok := strings.CutSuffix(text, "*/"); ok {
			text = withoutSuffix
		}
	}

	currentOffset := consumed

	for text != "" {
		line := text
		advance := len(text)

		if before, after, found := strings.Cut(line, "\n"); found {
			line = before
			advance = len(before) + 1
			text = after
		} else {
			text = ""
		}

		lineOffset := currentOffset
		currentOffset += advance
		trimmed, leftTrim := trimDocLine(line)
		lineOffset += leftTrim

		if trimmed == "" {
			continue
		}

		return trimmed, lineOffset
	}

	return "", 0
}

// trimDocLine removes leading comment markers and trailing whitespace.
func trimDocLine(line string) (string, int) {
	if line == "" {
		return "", 0
	}

	var consumed int

	trimLeft := func(cutset string) {
		trimmed := strings.TrimLeft(line, cutset)

		consumed += len(line) - len(trimmed)

		line = trimmed
	}

	trimLeft(" \t\r")
	trimLeft("* \t")
	trimLeft(" \t")

	line = strings.TrimRight(line, " \t\r")

	return line, consumed
}

// identifierFromLine finds the first identifier token within a line.
func identifierFromLine(line string) (string, int) {
	if line == "" {
		return "", 0
	}

	var i int

	isDocSpace := func(r rune) bool { return r == ' ' || r == '\t' }

	for i < len(line) {
		rest := line[i:]

		skip := strings.IndexFunc(rest, func(r rune) bool { return !isDocSpace(r) })
		if skip == -1 {
			break
		}

		i += skip
		tokenStart := i

		wordLen := strings.IndexFunc(line[i:], isDocSpace)
		if wordLen == -1 {
			wordLen = len(line) - i
		}

		word := line[tokenStart : tokenStart+wordLen]
		i += wordLen

		trimmed, leftTrim := trimWord(word)
		if trimmed == "" {
			continue
		}

		label := trimmed
		if withoutColon, ok := strings.CutSuffix(label, ":"); ok {
			label = withoutColon
		}

		lw := strings.ToLower(label)
		if isSkippableLabel(lw) {
			continue
		}

		if id, rel := extractIdentifierToken(trimmed); id != "" {
			return id, tokenStart + leftTrim + rel
		}

		break
	}

	return "", 0
}

// trimWord strips punctuation around a token and returns the offset.
func trimWord(word string) (string, int) {
	trimmed := strings.TrimLeftFunc(word, isWordBoundaryRune)

	left := len(word) - len(trimmed)

	trimmed = strings.TrimRightFunc(trimmed, isWordBoundaryRune)

	return trimmed, left
}

// isWordBoundaryRune reports whether the rune terminates identifier scanning.
func isWordBoundaryRune(r rune) bool {
	switch r {
	case ',', '.', ';', ':', '(', ')', '[', ']', '{', '}', '\t', ' ', '\r':
		return true
	}

	return false
}

// trimPointerPrefixes removes leading pointer markers before scanning.
func trimPointerPrefixes(s string) (string, int) {
	trimmed := strings.TrimLeft(s, "*&")

	return trimmed, len(s) - len(trimmed)
}

// leadingIdentRun reads the initial identifier characters from a string.
func leadingIdentRun(s string) (string, int) {
	var b strings.Builder

	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == '-' || r == '.' || r == '"' || r == '\'' {
			break
		}

		if r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			break
		}

		if r == ':' || r == ';' || r == ',' || r == ')' || r == '(' || r == ']' || r == '[' || r == '{' || r == '}' {
			break
		}

		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r == '_' {
			b.WriteRune(r)

			i += size

			continue
		}

		break
	}

	if b.Len() == 0 {
		return "", 0
	}

	return b.String(), i
}

// extractIdentifierToken pulls the last identifier component from a token.
func extractIdentifierToken(word string) (string, int) {
	if word == "" {
		return "", 0
	}

	trimmed, removed := trimPointerPrefixes(word)
	if id, _ := leadingIdentRun(trimmed); id != "" {
		return id, removed
	}

	return "", 0
}

// isSkippableLabel returns true for doc labels like Deprecated or TODO.
func isSkippableLabel(word string) bool {
	switch word {
	case "deprecated", "todo", "note", "fixme", "nolint", "lint", "warning":
		return true
	}

	return false
}

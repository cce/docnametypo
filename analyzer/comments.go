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
	consumed := 0
	switch {
	case strings.HasPrefix(text, "//"):
		text = text[2:]
		consumed += 2
	case strings.HasPrefix(text, "/*"):
		text = text[2:]
		consumed += 2
		text = strings.TrimSuffix(text, "*/")
	}

	currentOffset := consumed
	for len(text) > 0 {
		newline := strings.IndexByte(text, '\n')
		var line string
		var advance int
		if newline == -1 {
			line = text
			advance = len(text)
			text = ""
		} else {
			line = text[:newline]
			advance = newline + 1
			text = text[advance:]
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
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t' || line[i] == '\r') {
		i++
	}
	consumed := i
	line = line[i:]

	i = 0
	for i < len(line) && (line[i] == '*' || line[i] == ' ' || line[i] == '\t') {
		i++
	}
	consumed += i
	line = line[i:]

	i = 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	consumed += i
	line = line[i:]

	line = strings.TrimRight(line, " \t\r")
	return line, consumed
}

// identifierFromLine finds the first identifier token within a line.
func identifierFromLine(line string) (string, int) {
	if line == "" {
		return "", 0
	}
	i := 0
	for i < len(line) {
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		if i >= len(line) {
			break
		}
		tokenStart := i
		for i < len(line) && line[i] != ' ' && line[i] != '\t' {
			i++
		}
		word := line[tokenStart:i]
		trimmed, leftTrim := trimWord(word)
		if trimmed == "" {
			continue
		}
		lw := strings.ToLower(strings.TrimSuffix(trimmed, ":"))
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
	left := 0
	right := len(word)
	for left < right && isWordBoundary(word[left]) {
		left++
	}
	for right > left && isWordBoundary(word[right-1]) {
		right--
	}
	return word[left:right], left
}

// isWordBoundary reports whether the rune terminates identifier scanning.
func isWordBoundary(b byte) bool {
	switch b {
	case ',', '.', ';', ':', '(', ')', '[', ']', '{', '}', '\t', ' ', '\r':
		return true
	}
	return false
}

// trimPointerPrefixes removes leading pointer markers before scanning.
func trimPointerPrefixes(s string) (string, int) {
	i := 0
	for i < len(s) {
		if s[i] == '*' || s[i] == '&' {
			i++
			continue
		}
		break
	}
	return s[i:], i
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
	if strings.Contains(word, ".") {
		parts := strings.Split(word, ".")
		offsets := make([]int, len(parts))
		off := 0
		for i, part := range parts {
			offsets[i] = off
			off += len(part)
			if i < len(parts)-1 {
				off++
			}
		}
		for i := len(parts) - 1; i >= 0; i-- {
			trimmed, removed := trimPointerPrefixes(parts[i])
			if id, _ := leadingIdentRun(trimmed); id != "" {
				return id, offsets[i] + removed
			}
		}
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

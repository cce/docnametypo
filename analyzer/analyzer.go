// Package analyzer provides a go/analysis Analyzer that flags doc comments
// which appear to intend starting with the function/method name, but contain a
// likely typo or stale name (e.g., after refactors).
package analyzer

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	maxDistFlag                 = 1
	includeUnexportedFlag       = true
	includeExportedFlag         = false
	includeTypesFlag            = false
	includeGeneratedFlag        = false
	includeInterfaceMethodsFlag = false
)

const (
	minDocTokenLen   = 3
	maxChunkDiffSize = 6
)

// Analyzer implements the check.
var Analyzer = &analysis.Analyzer{
	Name:     "docnamecheck",
	Doc:      "flag doc comments that start with an identifier very similar to the symbol's name (probable typo/stale)",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func init() {
	Analyzer.Flags.IntVar(&maxDistFlag, "maxdist", 1, "maximum Damerau-Levenshtein distance to consider a likely typo")
	Analyzer.Flags.BoolVar(&includeUnexportedFlag, "include-unexported", true, "check unexported declarations")
	Analyzer.Flags.BoolVar(&includeExportedFlag, "include-exported", false, "check exported declarations (disabled by default)")
	Analyzer.Flags.BoolVar(&includeTypesFlag, "include-types", false, "also check type declarations")
	Analyzer.Flags.BoolVar(&includeGeneratedFlag, "include-generated", false, "check files marked as generated")
	Analyzer.Flags.BoolVar(&includeInterfaceMethodsFlag, "include-interface-methods", false, "check interface method declarations")
}

func run(pass *analysis.Pass) (any, error) {
	tokenToAST := make(map[*token.File]*ast.File, len(pass.Files))
	for _, f := range pass.Files {
		if f == nil {
			continue
		}
		if tf := pass.Fset.File(f.Pos()); tf != nil {
			tokenToAST[tf] = f
		}
	}
	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil), (*ast.GenDecl)(nil)}
	ins.Preorder(nodeFilter, func(n ast.Node) {
		if !includeGeneratedFlag {
			if tf := pass.Fset.File(n.Pos()); tf != nil {
				if af, ok := tokenToAST[tf]; ok && ast.IsGenerated(af) {
					return
				}
			}
		}
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Doc == nil || node.Name == nil {
				return
			}
			checkSymbol(pass, node.Doc, node.Name.Name, ast.IsExported(node.Name.Name), kindFunc, node.Name.Pos())
		case *ast.GenDecl:
			if node.Tok != token.TYPE {
				return
			}
			for _, spec := range node.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name == nil {
					continue
				}
				if includeTypesFlag {
					doc := ts.Doc
					if doc == nil {
						doc = node.Doc
					}
					if doc != nil {
						checkSymbol(pass, doc, ts.Name.Name, ast.IsExported(ts.Name.Name), kindType, ts.Name.Pos())
					}
				}
				if includeInterfaceMethodsFlag {
					if iface, ok := ts.Type.(*ast.InterfaceType); ok {
						checkInterfaceMethods(pass, iface)
					}
				}
			}
		}
	})
	return nil, nil
}

type symbolKind int

const (
	kindFunc symbolKind = iota
	kindType
)

func checkSymbol(pass *analysis.Pass, doc *ast.CommentGroup, name string, exported bool, kind symbolKind, declPos token.Pos) {
	if name == "" || doc == nil {
		return
	}

	if exported {
		if !includeExportedFlag {
			return
		}
	} else if !includeUnexportedFlag {
		return
	}

	firstTok, tokStart, tokEnd := firstIdentifierLike(doc)
	if firstTok == "" || len(firstTok) < minDocTokenLen {
		return
	}

	if kind == kindFunc && isNarrativeVerbForm(firstTok, name) {
		return
	}

	lenDiff := abs(len(firstTok) - len(name))
	if lenDiff > maxDistFlag+1 && lenDiff > maxChunkDiffSize {
		return
	}

	docLower := strings.ToLower(firstTok)
	nameLower := strings.ToLower(name)
	d := damerauLevenshtein(docLower, nameLower)
	match := d > 0 && d <= maxDistFlag
	if !match && isCamelSwapVariant(firstTok, name) {
		match = true
	}
	if !match && strings.EqualFold(firstTok, name) && firstTok != name {
		match = true
	}
	if !match && hasSimilarCamelWord(firstTok, name) {
		match = true
	}
	if !match && hasSmallChunkDifference(docLower, nameLower, maxChunkDiffSize) {
		match = true
	}
	if match {
		msg := "doc comment starts with '" + firstTok + "' but symbol is '" + name + "' (possible typo or old name)"
		var fixes []analysis.SuggestedFix
		if tokStart.IsValid() && tokEnd.IsValid() && tokStart < tokEnd {
			fixes = []analysis.SuggestedFix{{
				Message:   "replace doc token with symbol name",
				TextEdits: []analysis.TextEdit{{Pos: tokStart, End: tokEnd, NewText: []byte(name)}},
			}}
		}
		pass.Report(analysis.Diagnostic{
			Pos:            declPos,
			Message:        msg,
			SuggestedFixes: fixes,
		})
	}
}

// firstIdentifierLike extracts the first identifier-looking token from the first non-empty
// line of a comment group (skipping common labels like Deprecated:). It also returns the
// exact token.Pos range so a SuggestedFix can rewrite the token in-place.
func firstIdentifierLike(cg *ast.CommentGroup) (string, token.Pos, token.Pos) {
	if cg == nil || len(cg.List) == 0 {
		return "", token.NoPos, token.NoPos
	}
	comment := cg.List[0]
	line, lineOffset := firstDocLine(comment.Text)
	if line == "" {
		return "", token.NoPos, token.NoPos
	}
	id, rel := identifierFromLine(line)
	if id == "" {
		return "", token.NoPos, token.NoPos
	}
	start := comment.Slash + token.Pos(lineOffset+rel)
	end := start + token.Pos(len(id))
	return id, start, end
}

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

func isWordBoundary(b byte) bool {
	switch b {
	case ',', '.', ';', ':', '(', ')', '[', ']', '{', '}', '\t', ' ', '\r':
		return true
	}
	return false
}

func checkInterfaceMethods(pass *analysis.Pass, iface *ast.InterfaceType) {
	if iface == nil || iface.Methods == nil {
		return
	}
	for _, field := range iface.Methods.List {
		if field == nil || len(field.Names) == 0 {
			continue
		}
		doc := field.Doc
		if doc == nil {
			doc = field.Comment
		}
		if doc == nil {
			continue
		}
		for _, name := range field.Names {
			if name == nil {
				continue
			}
			checkSymbol(pass, doc, name.Name, ast.IsExported(name.Name), kindFunc, name.Pos())
		}
	}
}

func isSkippableLabel(w string) bool {
	switch w {
	case "deprecated", "todo", "note", "fixme", "nolint", "lint", "warning":
		return true
	}
	return false
}

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

func isCamelSwapVariant(docToken, symbol string) bool {
	docWords := splitCamelWords(docToken)
	symbolWords := splitCamelWords(symbol)
	if len(docWords) != len(symbolWords) || len(docWords) < 2 {
		return false
	}
	var diffs [2]int
	diffCount := 0
	for i := range docWords {
		if docWords[i] == symbolWords[i] {
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
	return docWords[i] == symbolWords[j] && docWords[j] == symbolWords[i]
}

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

func hasSimilarCamelWord(docToken, symbol string) bool {
	docWords := splitCamelWords(docToken)
	symbolWords := splitCamelWords(symbol)
	if len(docWords) == 0 || len(docWords) != len(symbolWords) {
		return false
	}
	mismatches := 0
	for i := range docWords {
		if docWords[i] == symbolWords[i] {
			continue
		}
		if !wordClose(docWords[i], symbolWords[i]) {
			return false
		}
		mismatches++
		if mismatches > 1 {
			return false
		}
	}
	return mismatches > 0
}

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

func isNarrativeVerbForm(word, funcName string) bool {
	if len(word) < 2 {
		return false
	}
	lowerWord := strings.ToLower(word)
	if !strings.HasSuffix(lowerWord, "s") {
		return false
	}
	stem := lowerWord[:len(lowerWord)-1]
	if stem == "" {
		return false
	}
	return strings.HasPrefix(strings.ToLower(funcName), stem)
}

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
		if idx+1 < len(runes) && unicode.IsLower(runes[idx+1]) {
			return true
		}
	}
	return false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func commonPrefixLength(a, b string) int {
	minLen := min(len(a), len(b))
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return minLen
}

func commonSuffixLength(a, b string) int {
	ia := len(a) - 1
	ib := len(b) - 1
	count := 0
	for ia >= 0 && ib >= 0 {
		if a[ia] != b[ib] {
			break
		}
		count++
		ia--
		ib--
	}
	return count
}

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

// damerauLevenshtein computes the optimal string edit distance with transpositions.
// Simple O(len(a)*len(b)) DP; fine for short identifiers.
func damerauLevenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	na := len(ra)
	nb := len(rb)
	if na == 0 {
		return nb
	}
	if nb == 0 {
		return na
	}
	d := make([][]int, na+1)
	for i := 0; i <= na; i++ {
		d[i] = make([]int, nb+1)
		d[i][0] = i
	}
	for j := 0; j <= nb; j++ {
		d[0][j] = j
	}

	for i := 1; i <= na; i++ {
		for j := 1; j <= nb; j++ {
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			del := d[i-1][j] + 1
			ins := d[i][j-1] + 1
			sub := d[i-1][j-1] + cost
			v := min3(del, ins, sub)
			// transposition
			if i > 1 && j > 1 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				v = min(v, d[i-2][j-2]+1)
			}
			d[i][j] = v
		}
	}
	return d[na][nb]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func min3(a, b, c int) int { return min(min(a, b), c) }

// Package analyzer provides a go/analysis Analyzer that flags doc comments
// which appear to intend starting with the function/method name, but contain a
// likely typo or stale name (e.g., after refactors).
package analyzer

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	maxDistFlag           = 1
	includeUnexportedFlag = true
	includeExportedFlag   = false
	includeTypesFlag      = false
	includeGeneratedFlag  = false
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
			if node.Tok != token.TYPE || !includeTypesFlag {
				return
			}
			for _, spec := range node.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name == nil {
					continue
				}
				doc := ts.Doc
				if doc == nil {
					doc = node.Doc
				}
				if doc == nil {
					continue
				}
				checkSymbol(pass, doc, ts.Name.Name, ast.IsExported(ts.Name.Name), kindType, ts.Name.Pos())
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

	firstTok, _ := firstIdentifierLike(doc)
	if firstTok == "" || len(firstTok) < minDocTokenLen {
		return
	}

	if kind == kindFunc && isNarrativeVerbForm(firstTok, name) {
		return
	}

	if !exported && isUpperFirstRune(firstTok) {
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
		pass.Report(analysis.Diagnostic{
			Pos:     declPos,
			Message: msg,
		})
	}
}

// firstIdentifierLike extracts the first identifier-looking token from the first non-empty
// line of a comment group (skipping common labels like Deprecated:).
func firstIdentifierLike(cg *ast.CommentGroup) (string, token.Pos) {
	if cg == nil || len(cg.List) == 0 {
		return "", token.NoPos
	}
	// Use the text of the first comment in the group; godoc rules expect the lead sentence there.
	text := cg.List[0].Text
	// Trim comment markers.
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimPrefix(text, "/*")
	text = strings.TrimSuffix(text, "*/")
	line := strings.TrimSpace(text)
	if line == "" {
		return "", token.NoPos
	}

	// Split by spaces; scan tokens, skipping well-known prefixes.
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", token.NoPos
	}
	// Skip labels like Deprecated:, TODO:, NOTE: etc.
	start := 0
	for start < len(fields) {
		w := strings.Trim(fields[start], ",.;:()[]{}\t ")
		lw := strings.ToLower(strings.TrimSuffix(w, ":"))
		if isSkippableLabel(lw) {
			start++
			continue
		}
		// First non-skippable word: strip punctuation and keep identifier run.
		id := extractIdentifierToken(w)
		if id != "" {
			return id, cg.List[0].Slash
		}
		break
	}
	return "", token.NoPos
}

func isSkippableLabel(w string) bool {
	switch w {
	case "deprecated", "todo", "note", "fixme", "nolint", "lint", "warning":
		return true
	}
	return false
}

func leadingIdentRun(s string) string {
	// Accept letters, digits, and underscores until first non-identifier rune.
	var b strings.Builder
	for _, r := range s {
		if r == '-' { // treat hyphen as a breaker (avoid kebab/narrative)
			break
		}
		if r == '.' || r == '"' || r == '\'' {
			break
		}
		if r == '\n' || r == '\r' || r == '\t' {
			break
		}
		if r == ' ' {
			break
		}
		if r == ':' || r == ';' || r == ',' || r == ')' || r == '(' || r == ']' || r == '[' || r == '{' || r == '}' {
			break
		}
		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r == '_' {
			b.WriteRune(r)
			continue
		}
		// any other symbol ends the ident-like run
		break
	}
	return b.String()
}

func extractIdentifierToken(word string) string {
	if word == "" {
		return ""
	}
	if strings.Contains(word, ".") {
		parts := strings.Split(word, ".")
		for i := len(parts) - 1; i >= 0; i-- {
			part := strings.TrimLeft(parts[i], "*&")
			if id := leadingIdentRun(part); id != "" {
				return id
			}
		}
	}
	return leadingIdentRun(word)
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

func isUpperFirstRune(s string) bool {
	for _, r := range s {
		return 'A' <= r && r <= 'Z'
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

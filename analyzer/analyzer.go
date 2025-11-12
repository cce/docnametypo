// Package analyzer provides a go/analysis Analyzer that flags doc comments
// which appear to intend starting with the function/method name, but contain a
// likely typo or stale name (e.g., after refactors).
package analyzer

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer implements the check.
var Analyzer = newAnalyzer()

func newAnalyzer() *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name:     "docnametypo",
		Doc:      "flag doc comments that start with an identifier very similar to the symbol's name (probable typo/stale)",
		Run:      run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	a.Flags.IntVar(&maxDistFlag, "maxdist", maxDistFlag, "maximum Damerau-Levenshtein distance to consider a likely typo")
	a.Flags.BoolVar(&includeUnexportedFlag, "include-unexported", includeUnexportedFlag, "check unexported declarations")
	a.Flags.BoolVar(&includeExportedFlag, "include-exported", includeExportedFlag, "check exported declarations (disabled by default)")
	a.Flags.BoolVar(&includeTypesFlag, "include-types", includeTypesFlag, "also check type declarations")
	a.Flags.BoolVar(&includeGeneratedFlag, "include-generated", includeGeneratedFlag, "check files marked as generated")
	a.Flags.BoolVar(&includeInterfaceMethodsFlag, "include-interface-methods", includeInterfaceMethodsFlag, "check interface method declarations")
	a.Flags.StringVar(&allowedLeadingWordsFlag, "allowed-leading-words", allowedLeadingWordsFlag, "comma-separated list of leading words to ignore (treated as narrative)")
	a.Flags.StringVar(&allowedPrefixesFlag, "allowed-prefixes", allowedPrefixesFlag, "comma-separated list of symbol prefixes to ignore when matching doc tokens")
	a.Flags.BoolVar(&skipPlainWordCamelFlag, "skip-plain-word-camel", skipPlainWordCamelFlag, "skip plain leading words when the symbol looks camelCase (reduces narrative false positives)")
	a.Flags.IntVar(&maxCamelChunkInsertFlag, "max-camel-chunk-insert", maxCamelChunkInsertFlag, "maximum number of camelCase chunks that may be inserted or removed (detects missing words)")
	a.Flags.IntVar(&maxCamelChunkReplaceFlag, "max-camel-chunk-replace", maxCamelChunkReplaceFlag, "maximum number of camelCase chunks that may be replaced (detects word changes)")

	return a
}

func run(pass *analysis.Pass) (any, error) {
	cfg := newMatchConfig()

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
			checkSymbol(pass, cfg, node.Doc, node.Name.Name, ast.IsExported(node.Name.Name), kindFunc, node.Name.Pos())

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
						checkSymbol(pass, cfg, doc, ts.Name.Name, ast.IsExported(ts.Name.Name), kindType, ts.Name.Pos())
					}
				}

				if includeInterfaceMethodsFlag {
					if iface, ok := ts.Type.(*ast.InterfaceType); ok {
						checkInterfaceMethods(pass, cfg, iface)
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

// checkSymbol compares the comment token against the provided symbol.
func checkSymbol(pass *analysis.Pass, cfg matchConfig, doc *ast.CommentGroup, name string, exported bool, kind symbolKind, declPos token.Pos) {
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

	firstTok, tokStart, tokEnd, docLine := firstIdentifierLike(doc)
	if firstTok == "" || len(firstTok) < minDocTokenLen {
		return
	}

	if docFirstWordHasDot(docLine) {
		return
	}
	if cfg.isAllowedLeadingWord(firstTok) {
		return
	}
	if cfg.matchesAllowedPrefixVariant(firstTok, name) {
		return
	}
	if isSectionHeader(firstTok, docLine) {
		return
	}
	if isNarrativeSentenceIntro(firstTok, docLine) {
		return
	}
	if containsWildcardToken(firstTok, docLine) {
		return
	}
	if kind == kindFunc && isNarrativeVerbForm(firstTok, name) {
		return
	}
	if skipPlainWordCamelFlag && looksLikeSimpleWord(firstTok) && hasCamelCaseInterior(name) {
		return
	}

	lenDiff := abs(len(firstTok) - len(name))
	var docLower, nameLower string
	match := false
	if lenDiff <= maxDistFlag+1 || lenDiff <= maxChunkDiffSize {
		docLower = strings.ToLower(firstTok)
		nameLower = strings.ToLower(name)
		d := damerauLevenshtein(docLower, nameLower)
		match = d > 0 && d <= maxDistFlag
		if match && !passesDistanceGate(docLower, nameLower, d) {
			match = false
		}
	}

	if !match && isCamelSwapVariant(firstTok, name) {
		match = true
	}
	if !match && strings.EqualFold(firstTok, name) && firstTok != name {
		match = true
	}
	if !match && hasSimilarCamelWord(firstTok, name) {
		match = true
	}
	if !match && hasCamelChunkReplacement(firstTok, name, maxCamelChunkReplaceFlag) {
		match = true
	}
	if !match && hasCamelChunkInsertionOrRemoval(firstTok, name, maxCamelChunkInsertFlag) {
		match = true
	}
	if !match && nameLower != "" && docLower != "" && hasSmallChunkDifference(docLower, nameLower, maxChunkDiffSize) {
		match = true
	}

	if !match {
		return
	}

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

// checkInterfaceMethods inspects each interface method doc comment.
func checkInterfaceMethods(pass *analysis.Pass, cfg matchConfig, iface *ast.InterfaceType) {
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
			checkSymbol(pass, cfg, doc, name.Name, ast.IsExported(name.Name), kindFunc, name.Pos())
		}
	}
}

package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testCases := []struct {
		desc  string
		setup func()
		dir   string
	}{
		{
			desc: "defaults",
			dir:  "unexported",
		},
		{
			desc: "exportedAndTypes",
			setup: func() {
				includeExportedFlag = true
				includeTypesFlag = true
			},
			dir: "exported",
		},
		{
			desc: "generatedOptIn",
			setup: func() {
				includeExportedFlag = true
				includeGeneratedFlag = true
			},
			dir: "generatedcode",
		},
		{
			desc: "interfaceMethodsOptIn",
			setup: func() {
				includeInterfaceMethodsFlag = true
			},
			dir: "interfaces",
		},
		{
			desc: "narrativeLeadingWords",
			dir:  "narrative",
		},
		{
			desc: "allowedPrefixes",
			setup: func() {
				allowedPrefixesFlag = "asm,op"
			},
			dir: "prefixaliases",
		},
		{
			desc: "plainWordCamelFlag",
			dir:  "plainwordcamel",
		},
		{
			desc: "plainWordCamelFlagDisabled",
			setup: func() {
				skipPlainWordCamelFlag = false
			},
			dir: "plainwordcamelexpect",
		},
		{
			desc: "maxDistanceGate",
			setup: func() {
				maxDistFlag = 5
			},
			dir: "maxdistance",
		},
		{
			desc: "camelChunkHeuristics",
			dir:  "camelchunks",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			resetFlags()

			if test.setup != nil {
				test.setup()
			}

			analysistest.Run(t, analysistest.TestData(), Analyzer, test.dir)
		})
	}

	t.Run("fixSuggested", func(t *testing.T) {
		resetFlags()
		analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), Analyzer, "fixes")
	})
}

func resetFlags() {
	maxDistFlag = 5
	includeUnexportedFlag = true
	includeExportedFlag = false
	includeTypesFlag = false
	includeGeneratedFlag = false
	includeInterfaceMethodsFlag = false
	allowedLeadingWordsFlag = defaultAllowedLeadingWords
	allowedPrefixesFlag = ""
	skipPlainWordCamelFlag = true
	maxCamelChunkInsertFlag = 2
	maxCamelChunkReplaceFlag = 2
}

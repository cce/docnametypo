package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		resetFlags()
		analysistest.Run(t, analysistest.TestData(), Analyzer, "unexported")
	})

	t.Run("exportedAndTypes", func(t *testing.T) {
		resetFlags()
		includeExportedFlag = true
		includeTypesFlag = true
		analysistest.Run(t, analysistest.TestData(), Analyzer, "exported")
	})

	t.Run("generatedOptIn", func(t *testing.T) {
		resetFlags()
		includeExportedFlag = true
		includeGeneratedFlag = true
		analysistest.Run(t, analysistest.TestData(), Analyzer, "generatedcode")
	})

	t.Run("interfaceMethodsOptIn", func(t *testing.T) {
		resetFlags()
		includeInterfaceMethodsFlag = true
		analysistest.Run(t, analysistest.TestData(), Analyzer, "interfaces")
	})

	t.Run("fixSuggested", func(t *testing.T) {
		resetFlags()
		analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), Analyzer, "fixes")
	})
}

func resetFlags() {
	maxDistFlag = 1
	includeUnexportedFlag = true
	includeExportedFlag = false
	includeTypesFlag = false
	includeGeneratedFlag = false
	includeInterfaceMethodsFlag = false
}

package gclplugin

import (
	"fmt"
	"strconv"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/cce/docnamecheck/analyzer"
)

func init() {
	register.Plugin("docnamecheck", New)
}

// Plugin implements register.LinterPlugin for docnamecheck.
type Plugin struct {
	settings Settings
}

// New constructs a Plugin instance from raw settings.
func New(raw any) (register.LinterPlugin, error) {
	settings, err := register.DecodeSettings[Settings](raw)
	if err != nil {
		return nil, err
	}
	return Plugin{settings: settings}, nil
}

// GetLoadMode declares the loader requirements.
func (Plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}

// BuildAnalyzers wires the configured analyzer.
func (p Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	if err := applySettings(p.settings); err != nil {
		return nil, err
	}
	return []*analysis.Analyzer{analyzer.Analyzer}, nil
}

func applySettings(s Settings) error {
	if s.MaxDist != nil {
		if err := analyzer.Analyzer.Flags.Set("maxdist", strconv.Itoa(*s.MaxDist)); err != nil {
			return fmt.Errorf("set maxdist: %w", err)
		}
	}
	if s.IncludeExported != nil {
		if err := analyzer.Analyzer.Flags.Set("include-exported", strconv.FormatBool(*s.IncludeExported)); err != nil {
			return fmt.Errorf("set include-exported: %w", err)
		}
	}
	if s.IncludeUnexported != nil {
		if err := analyzer.Analyzer.Flags.Set("include-unexported", strconv.FormatBool(*s.IncludeUnexported)); err != nil {
			return fmt.Errorf("set include-unexported: %w", err)
		}
	}
	if s.IncludeTypes != nil {
		if err := analyzer.Analyzer.Flags.Set("include-types", strconv.FormatBool(*s.IncludeTypes)); err != nil {
			return fmt.Errorf("set include-types: %w", err)
		}
	}
	if s.IncludeGenerated != nil {
		if err := analyzer.Analyzer.Flags.Set("include-generated", strconv.FormatBool(*s.IncludeGenerated)); err != nil {
			return fmt.Errorf("set include-generated: %w", err)
		}
	}
	return nil
}

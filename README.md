# docnamecheck

[![build and test](https://github.com/cce/docnamecheck/actions/workflows/test.yml/badge.svg)](https://github.com/cce/docnamecheck/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cce/docnamecheck.svg)](https://pkg.go.dev/github.com/cce/docnamecheck)
[![Go Report Card](https://goreportcard.com/badge/github.com/cce/docnamecheck)](https://goreportcard.com/report/github.com/cce/docnamecheck)
[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)

_Detect when Go doc comments probably meant to reference the identifier but got it wrong._

`docnamecheck` is a linter that analyzes the first word in doc comments to understand whether it's attempting to reference the identifier name. It flags cases where the comment's first word seems to intend to reference the identifier but doesn't match, due to typos, refactoring, or copy-paste errors, while allowing you to write narrative comments freely.

## Table of Contents

- [Why docnamecheck?](#why-docnamecheck)
- [Quick Start](#quick-start)
- [Examples](#examples)
- [Installation](#installation)
- [Usage](#usage)
- [golangci-lint Integration](#golangci-lint-module-plugin)
- [How It Works](#how-it-works)
- [Configuration Tips](#configuration-tips)
- [Related Tools](#related-tools)
- [Troubleshooting](#troubleshooting)
- [License](#license)

## Why docnamecheck?

Does your codebase sometimes start doc comments with the function or type name for unexported symbols, but not always? Many codebases use a relaxed documentation style that allows both:

```go
// parseConfig reads and validates the configuration file
func parseConfig(path string) error { ... }

// Reads the manifest and returns structured data
func parseManifest(path string) (*Manifest, error) { ... }
```

If you're using linters like `revive` or `go vet`, they'll enforce strict `// Name does X` formatting for exported functions. But for unexported code, your codebase might not follow this rule consistently, and that's fine.

**The problem:** When you refactor code, it's easy to miss updating doc comments that referenced the old name:

```go
// parseConfig reads and validates configuration  <- Stale comment!
func parseManifest(path string) error { ... }

// ServerHTTP handles incoming requests  <- Typo!
func ServeHTTP(w http.ResponseWriter, r *http.Request) { ... }
```

**The solution:** `docnamecheck` uses heuristics to understand your intent. It analyzes whether a comment appears to be trying to use the symbol name (and got it wrong) versus writing narrative prose (which is fine). The tool catches actual mistakes while staying out of your way for legitimate narrative comments.

### How it understands intent

`docnamecheck` uses multiple strategies to distinguish typos from narrative:

- **Damerau-Levenshtein distance**: Catches typos and single-character transpositions (`confgure` vs `configure`)
- **CamelCase analysis**: Detects reordered words (`ServerHTTP` vs `ServeHTTP`) or missing chunks
- **Capitalization patterns**: Flags `newHandler` when the function is `NewHandler`
- **Narrative detection**: Skips comments starting with verbs like `Creates`, `Initializes`, `Generates`, etc.
- **Prefix handling**: Allows configured prefixes like `op` to be stripped before matching

This lets your codebase maintain a loose practice for documenting code—sometimes using the function name as the first word, sometimes not—while still catching cases where the comment clearly intended to reference the name but got it wrong.

## Quick Start

```bash
# Install
go install github.com/cce/docnamecheck/cmd/docnamecheck@latest

# Run on your codebase
docnamecheck ./...

# Auto-fix issues
docnamecheck -fix ./...
```

## Examples

### What docnamecheck Reports

**Before:**
```go
// ServerHTTP handles incoming HTTP requests and routes them
// to the appropriate handler based on the request path.
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

**Output:**
```
example.go:1:1: doc comment starts with "ServerHTTP" but function name is "ServeHTTP" (2 transposed camelCase words)
```

**After fix:**
```go
// ServeHTTP handles incoming HTTP requests and routes them
// to the appropriate handler based on the request path.
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

### More Examples

**Levenshtein distance detection:**
```go
// confgure sets up the application  <- Typo: "confgure" vs "configure"
func configure(app *App) error { ... }
```

**Prefix mismatch:**
```go
// newTelemetryHook creates a hook  <- Should be "NewTelemetryHook"
func NewTelemetryHook() *Hook { ... }
```

**Narrative comments are allowed:**
```go
// Creates a new HTTP client  <- This is fine (starts with "Creates")
func newHTTPClient() *Client { ... }

// This helper generates test fixtures  <- This is fine (narrative)
func makeTestData() []byte { ... }

// Generates encryption keys for the session
func generateKeys() []byte { ... }  <- This is fine (narrative verb)
```

## Installation

```bash
go install github.com/cce/docnamecheck/cmd/docnamecheck@latest
```

## Usage

```bash
docnamecheck ./...
```

The analyzer understands several flags:

| Flag | Default | Description |
| --- | --- | --- |
| `-maxdist` | `1` | Maximum Damerau-Levenshtein distance before a pair of words stops being considered a typo. |
| `-include-unexported` | `true` | Check unexported functions/methods/types. This is the primary use case. |
| `-include-exported` | `false` | Also check exported declarations. Enable this if you do not already enforce `// Name ...` elsewhere. |
| `-include-types` | `false` | Extend the check to `type` declarations (honoring the exported/unexported switches above). |
| `-include-generated` | `false` | Include files that carry the `// Code generated ... DO NOT EDIT.` header; off by default to avoid noisy generated code. |
| `-include-interface-methods` | `false` | Check interface method declarations. Useful when interface docs must track implementation names. |
| `-allowed-leading-words` | *(see note)* | Comma-separated verbs treated as narrative intros (e.g. `Create`, `Configure`, `Tests`); matching comments are skipped. |
| `-allowed-prefixes` | `` | Comma-separated list of symbol prefixes (such as `op`) that may be stripped before comparing to the doc token. |

> **Note:** the default `-allowed-leading-words` list is `create,creates,creating,initialize,initializes,init,configure,configures,setup,setups,start,starts,read,reads,write,writes,send,sends,generate,generates,decode,decodes,encode,encodes,marshal,marshals,unmarshal,unmarshals,apply,applies,process,processes,make,makes,build,builds,test,tests`.

### Applying Fixes

`docnamecheck` emits suggested fixes that rewrite the incorrect identifier token in the doc comment. Run:

```bash
docnamecheck -fix ./...
```

to automatically apply those edits. The golangci-lint module plugin also respects `golangci-lint run --fix`, so the same replacement happens when the linter runs inside larger pipelines.

## golangci-lint Module Plugin

`docnamecheck` ships a golangci-lint module plugin. To integrate it:

1. Create `.custom-gcl.yml` next to your `go.mod`:

   ```yaml
   ---
   version: v2.5.0
   name: custom-golangci-lint
   plugins:
     - module: github.com/cce/docnamecheck
       import: github.com/cce/docnamecheck/gclplugin
       version: main # or a tagged release
   ```

2. Run `golangci-lint custom` to build a local binary that bundles the plugin (the command uses `.custom-gcl.yml`).

3. Reference the linter from your `.golangci.yml`, for example:

   ```yaml
   ---
   version: "2"
   linters:
     enable:
       - docnamecheck
     settings:
       custom:
         docnamecheck:
           type: module
           description: "docnamecheck catches doc comments whose first token drifted from the actual name"
           original-url: "https://github.com/cce/docnamecheck"
           settings:
             include-exported: true
             include-interface-methods: true
             include-types: true
             include-generated: false
             allowed-prefixes: asm,op
             allowed-leading-words: create,creates,setup,read
             maxdist: 2
   ```

## How It Works

`docnamecheck` uses multiple string matching algorithms to detect likely typos while avoiding false positives on legitimate narrative comments:

1. **Extracts the first identifier-like token** from doc comments, skipping labels such as `Deprecated:`, `TODO:`, `NOTE:`, etc.

2. **Compares using multiple algorithms:**
   - **Damerau-Levenshtein distance**: Catches typos and single-character transpositions
     - Example: `confgure` vs `configure` (distance = 1)
   - **CamelCase transposition detection**: Catches reordered words in camelCase
     - Example: `ServerHTTP` vs `ServeHTTP` (chunks swapped)
     - Example: `TelemetryHistoryState` vs `TelemetryHistory` (suffix difference)
   - **Prefix/suffix heuristics**: Catches capitalization mismatches
     - Example: `newTelemetryFilteredHook` vs `NewTelemetryFilteredHook`

3. **Filters false positives** by ignoring:
   - Narrative comments starting with common verbs (`generates`, `creates`, `initializes`, etc.)
   - Comments that clearly don't reference the symbol (diverge too much)
   - Configured prefix variants (e.g., `opThing` vs `Thing` when `op` is in `-allowed-prefixes`)

4. **Works across all symbol types**: functions, methods, types, and interface methods (based on configuration flags)

Because the analyzer is heuristic, the defaults stay conservative: only unexported symbols are checked out of the box so that it can complement, rather than duplicate, tools such as `godoclint`. Turn on `-include-exported`, `-include-interface-methods`, and `-include-types` when you want broader coverage.

## Configuration Tips

### For codebases with prefixed helpers

If your codebase uses consistent prefixes (e.g., `opThing`, `asmRegister`):

```bash
docnamecheck -allowed-prefixes=op,asm ./...
```

This allows doc comments to reference `Thing` in the first word when the function is `opThing`, without flagging this as a typo.

### For narrative documentation styles

If you prefer starting comments with imperative verbs:

```bash
docnamecheck -allowed-leading-words=create,configure,setup,validate,process ./...
```

Comments like "Configure sets up the application" won't be flagged for a function named `configure`.

### For strict enforcement

Check everything including exported symbols and types:

```bash
docnamecheck -include-exported -include-types -include-interface-methods ./...
```

### Adjusting sensitivity

Increase the distance threshold for looser matching:

```bash
docnamecheck -maxdist=2 ./...
```

Default is `1`, which catches single typos or transpositions.

## Related Tools

`docnamecheck` is designed to complement existing Go documentation linters:

- **[revive](https://github.com/mgechev/revive)**: General-purpose linter with exported comment rules. Use `docnamecheck` for unexported symbols or when you want typo detection instead of strict format enforcement.
- **[godoclint](https://github.com/tommy-muehle/go-mnd)**: Enforces strict "comment must start with name" rules. `docnamecheck` offers a middle ground that allows narrative comments while catching clear errors.
- **[golangci-lint](https://golangci-lint.run/)**: Meta-linter that can run multiple linters. `docnamecheck` integrates as a module plugin.

## Troubleshooting

### "Why isn't it flagging exported functions?"

By default, `docnamecheck` only checks unexported symbols to avoid duplicating other linters. Enable exported checking with:

```bash
docnamecheck -include-exported ./...
```

### "Too many false positives on narrative comments"

Add common starting verbs to the allowed list:

```bash
docnamecheck -allowed-leading-words=create,configure,setup,handle,process ./...
```

Or extend the default list:

```bash
# Check current defaults
docnamecheck -h | grep allowed-leading-words

# Add your own words to the list
docnamecheck -allowed-leading-words=create,...,yourword ./...
```

### "It's flagging comments on generated code"

Generated code is excluded by default. If you're still seeing issues, ensure your generated files have the standard header:

```go
// Code generated ... DO NOT EDIT.
```

### "False positives on prefixed helpers"

If your codebase uses consistent prefixes (e.g., all assembly helpers start with `asm`), use:

```bash
docnamecheck -allowed-prefixes=asm ./...
```

### "How do I run this in CI?"

Add to your GitHub Actions workflow:

```yaml
- name: Install docnamecheck
  run: go install github.com/cce/docnamecheck/cmd/docnamecheck@latest

- name: Run docnamecheck
  run: docnamecheck ./...
```

Or integrate via golangci-lint (see [integration section](#golangci-lint-module-plugin)).

## License

This project is licensed under the [BSD 3-Clause License](LICENSE).

---

**Found a bug or have a feature request?** [Open an issue](https://github.com/cce/docnamecheck/issues) on GitHub.

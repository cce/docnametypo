# docnamecheck

_A doc comment typo detector for Go identifiers that keeps unexported symbols honest and exported ones optionally double‑checked._

`docnamecheck` looks at the first identifier-like token in every doc comment and compares it to the symbol's actual
name using Damerau-Levenshtein distance, camel-case chunk comparison, and a handful of heuristics. It is designed to
cover the gap left after disabling "comments must start with the function name" rules for unexported code: refactors,
misspellings, or stale comments are still flagged when the comment clearly meant to mention the identifier.

## Installation

```shell
go install github.com/cce/docnamecheck/cmd/docnamecheck@latest
```

## Usage

```shell
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

The heuristics intentionally skip obviously narrative comments (`generates keys ...`), `NOTE:`/`TODO:` labels, and cases
where the prefix/suffix diverges too much, so you only hear about comments that almost certainly meant to reference the
symbol name.

## golangci-lint module plugin

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
             include-types: true
             include-generated: false
             maxdist: 2
   ```

## How it works

- Parses doc comments and extracts the first identifier-looking token, skipping labels such as `Deprecated:` or `NOTE:`.
- Compares that token to the actual symbol name using Damerau-Levenshtein distance, camel-case transposition detection,
  and suffix/prefix heuristics that catch `ServerHTTP` vs `ServeHTTP`, `newTelemetryFilteredHook` vs
  `NewTelemetryFilteredHook`, or `TelemetryHistoryState` vs `TelemetryHistory`.
- Ignores sentences that obviously read narratively (`generates`, `creates`, etc.) so you can keep writing regular prose
  when you intend to.
- Works across methods, top-level functions, and (optionally) types, whether or not they are exported.

Because the analyzer is heuristic, the defaults stay conservative: only unexported symbols are checked out of the box so
that it can complement, rather than duplicate, tools such as `godoclint`. Turn on `-include-exported` and `-include-types`
when you want broader coverage.

package analyzer

import "testing"

func Test_buildAllowedLeadingWords(t *testing.T) {
	got := buildAllowedLeadingWords("foo,bar /baz")
	if _, ok := got["foo"]; !ok {
		t.Fatalf("expected foo in map")
	}

	if _, ok := got["baz"]; !ok {
		t.Fatalf("expected baz in map")
	}
}

func Test_matchConfigAllowedPrefix(t *testing.T) {
	cfg := matchConfig{allowedPrefixes: []string{"op"}}
	if !cfg.matchesAllowedPrefixVariant("Thing", "opThing") {
		t.Fatalf("expected prefix match")
	}

	if cfg.matchesAllowedPrefixVariant("Other", "Thing") {
		t.Fatalf("did not expect match")
	}
}

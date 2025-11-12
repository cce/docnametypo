package analyzer

import "testing"

func Test_hasCamelChunkReplacement(t *testing.T) {
	tests := []struct {
		doc, sym string
		max      int
		want     bool
	}{
		{"processCIDRs", "validateCIDRs", 2, true},
		{"processCIDRs", "validateCIDRs", 1, true},
		{"processCIDRs", "validateFooCIDRs", 1, false},
		{"oneWord", "oneWord", 1, false},
		{"getPodIPs", "getIPsPod", 2, true},
	}
	for _, tt := range tests {
		if got := hasCamelChunkReplacement(tt.doc, tt.sym, tt.max); got != tt.want {
			t.Errorf("hasCamelChunkReplacement(%q,%q,%d)=%v, want %v", tt.doc, tt.sym, tt.max, got, tt.want)
		}
	}
}

func Test_hasCamelChunkInsertionOrRemoval(t *testing.T) {
	tests := []struct {
		doc, sym string
		max      int
		want     bool
	}{
		{"handleVolume", "handleEphemeralVolume", 2, true},
		{"syncHandler", "sync", 2, true},
		{"syncHandler", "sync", 0, false},
		{"fooBar", "fooBar", 2, false},
		{"UIDTracker", "UIDEventTracker", 2, true},
	}
	for _, tt := range tests {
		if got := hasCamelChunkInsertionOrRemoval(tt.doc, tt.sym, tt.max); got != tt.want {
			t.Errorf("hasCamelChunkInsertionOrRemoval(%q,%q,%d)=%v, want %v", tt.doc, tt.sym, tt.max, got, tt.want)
		}
	}
}

func Test_isCamelSwapVariant(t *testing.T) {
	tests := []struct {
		doc, sym string
		want     bool
	}{
		{"getPodsReady", "getReadyPods", true},
		{"getPodIPs", "getPodIPs", false},
		{"HTTPServerReady", "HTTPReadyServer", true},
	}
	for _, tt := range tests {
		if got := isCamelSwapVariant(tt.doc, tt.sym); got != tt.want {
			t.Errorf("isCamelSwapVariant(%q,%q)=%v, want %v", tt.doc, tt.sym, got, tt.want)
		}
	}
}

func Test_splitCamelWords(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"getPodIPs", []string{"get", "pod", "ips"}},
		{"HTTPServerReady", []string{"http", "server", "ready"}},
	}
	for _, tt := range tests {
		got := splitCamelWords(tt.input)
		if len(got) != len(tt.want) {
			t.Fatalf("splitCamelWords(%q)=%v want %v", tt.input, got, tt.want)
		}

		for i := range got {
			if got[i] != tt.want[i] {
				t.Fatalf("splitCamelWords(%q)=%v want %v", tt.input, got, tt.want)
			}
		}
	}
}

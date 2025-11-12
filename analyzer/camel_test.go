package analyzer

import "testing"

func Test_hasCamelChunkReplacement(t *testing.T) {
	tests := []struct {
		doc, sym string
		max      int
		want     bool
	}{
		{
			doc:  "processCIDRs",
			sym:  "validateCIDRs",
			max:  2,
			want: true,
		},
		{
			doc:  "processCIDRs",
			sym:  "validateCIDRs",
			max:  1,
			want: true,
		},
		{
			doc: "processCIDRs",
			sym: "validateFooCIDRs",
			max: 1,
		},
		{
			doc: "oneWord",
			sym: "oneWord",
			max: 1,
		},
		{
			doc:  "getPodIPs",
			sym:  "getIPsPod",
			max:  2,
			want: true,
		},
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
		{
			doc:  "handleVolume",
			sym:  "handleEphemeralVolume",
			max:  2,
			want: true,
		},
		{
			doc:  "syncHandler",
			sym:  "sync",
			max:  2,
			want: true,
		},
		{
			doc: "syncHandler",
			sym: "sync",
		},
		{
			doc: "fooBar",
			sym: "fooBar",
			max: 2,
		},
		{
			doc:  "UIDTracker",
			sym:  "UIDEventTracker",
			max:  2,
			want: true,
		},
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
		{
			doc:  "getPodsReady",
			sym:  "getReadyPods",
			want: true,
		},
		{
			doc: "getPodIPs",
			sym: "getPodIPs",
		},
		{
			doc:  "HTTPServerReady",
			sym:  "HTTPReadyServer",
			want: true,
		},
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
		{
			input: "getPodIPs",
			want:  []string{"get", "pod", "ips"},
		},
		{
			input: "HTTPServerReady",
			want:  []string{"http", "server", "ready"},
		},
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

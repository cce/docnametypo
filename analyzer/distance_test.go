package analyzer

import "testing"

func Test_passesDistanceGate(t *testing.T) {
	tests := []struct {
		doc, sym string
		dist     int
		want     bool
	}{
		{
			doc:  "validateAllowedTopology",
			sym:  "validateAllowedTopologies",
			dist: 2,
			want: true,
		},
		{
			doc:  "validateAllowedTopology",
			sym:  "Foo",
			dist: 2,
		},
	}
	for _, tt := range tests {
		if got := passesDistanceGate(tt.doc, tt.sym, tt.dist); got != tt.want {
			t.Errorf("passesDistanceGate(%q,%q,%d)=%v, want %v", tt.doc, tt.sym, tt.dist, got, tt.want)
		}
	}
}

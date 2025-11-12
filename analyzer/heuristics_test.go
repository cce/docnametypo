package analyzer

import "testing"

func Test_docFirstWordHasDot(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"reflect.DeepEqual doesn't work", true},
		{"foo.Bar is weird", true},
		{"UID.Event happens", true},
		{".Hello starts with dot", true},
		{"This. is a dot after", true},
		{"ServeHTTP handles", false},
	}
	for _, tt := range tests {
		if got := docFirstWordHasDot(tt.line); got != tt.want {
			t.Errorf("docFirstWordHasDot(%q)=%v, want %v", tt.line, got, tt.want)
		}
	}
}

func Test_containsWildcardToken(t *testing.T) {
	tests := []struct {
		token, line string
		want        bool
	}{
		{"commonPrefixLen*", "commonPrefixLen* returns", true},
		{"Token", "Token returns", false},
	}
	for _, tt := range tests {
		if got := containsWildcardToken(tt.token, tt.line); got != tt.want {
			t.Errorf("containsWildcardToken(%q,%q)=%v, want %v", tt.token, tt.line, got, tt.want)
		}
	}
}

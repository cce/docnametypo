package analyzer

import "testing"

func Test_docFirstWordHasDot(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{
			line: "reflect.DeepEqual doesn't work",
			want: true,
		},
		{
			line: "foo.Bar is weird",
			want: true,
		},
		{
			line: "UID.Event happens",
			want: true,
		},
		{
			line: ".Hello starts with dot",
			want: true,
		},
		{
			line: "This. is a dot after",
			want: true,
		},
		{
			line: "ServeHTTP handles",
		},
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
		{
			token: "commonPrefixLen*",
			line:  "commonPrefixLen* returns",
			want:  true,
		},
		{
			token: "Token",
			line:  "Token returns",
		},
	}
	for _, tt := range tests {
		if got := containsWildcardToken(tt.token, tt.line); got != tt.want {
			t.Errorf("containsWildcardToken(%q,%q)=%v, want %v", tt.token, tt.line, got, tt.want)
		}
	}
}

package oidc

import "testing"

func TestSanitizeReturnToAllowsOnlyLocalAbsolutePaths(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/", "/"},
		{"/mcp?x=1", "/mcp?x=1"},
		{"", "/"},
		{"relative", "/"},
		{"https://attacker.example", "/"},
		{"//attacker.example/path", "/"},
		{`/\\attacker.example`, "/"},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if got := sanitizeReturnTo(test.input); got != test.want {
				t.Fatalf("sanitizeReturnTo(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

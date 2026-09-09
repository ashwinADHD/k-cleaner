package match

import "testing"

func TestPearFormat(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"com.apple.Safari", "comapplesafari"},
		{"Google Chrome", "googlechrome"},
		{"", ""},
		{"Test-App_v2", "testappv2"},
	}
	for _, tc := range tests {
		got := PearFormat(tc.in)
		if got != tc.want {
			t.Errorf("PearFormat(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestMatchesAppStrict(t *testing.T) {
	if !MatchesApp("com.google.Chrome", "/Users/x/Library/Caches/com.google.Chrome",
		"com.google.Chrome", "Google Chrome", nil, 0) {
		t.Error("expected strict bundle ID match")
	}
	if MatchesApp("random", "/tmp/random", "com.google.Chrome", "Google Chrome", nil, 0) {
		t.Error("expected no match for unrelated name")
	}
}

package cli

import "testing"

func TestTrustKeyMatch(t *testing.T) {
	for _, c := range []struct {
		goos, key, path string
		want            bool
	}{
		{"windows", `C:\Users\Me\Repo`, "C:/Users/Me/Repo", true},
		{"windows", `c:\users\me\repo`, "C:/Users/Me/Repo", true},
		{"windows", `C:\Users\Other`, "C:/Users/Me/Repo", false},
		{"darwin", "/users/me/repo", "/Users/Me/Repo", true},
		{"darwin", "/Users/Me/Other", "/Users/Me/Repo", false},
		{"linux", "/home/me/repo", "/home/me/repo", true},
		{"linux", "/home/me/Repo", "/home/me/repo", false},
	} {
		if got := trustKeyMatch(c.goos, c.key, c.path); got != c.want {
			t.Errorf("%s %q vs %q = %v, want %v", c.goos, c.key, c.path, got, c.want)
		}
	}
}

func TestHookCounts(t *testing.T) {
	for _, c := range []struct {
		goos string
		mode uint32
		want bool
	}{
		{"windows", 0o644, true},
		{"windows", 0o755, true},
		{"darwin", 0o644, false},
		{"darwin", 0o755, true},
		{"linux", 0o644, false},
		{"linux", 0o700, true},
	} {
		if got := hookCounts(c.goos, c.mode); got != c.want {
			t.Errorf("%s %o = %v, want %v", c.goos, c.mode, got, c.want)
		}
	}
}

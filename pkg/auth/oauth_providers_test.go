package auth

import "testing"

// TestPickGitHubEmail covers the login e-mail chosen from GitHub and whether
// it counts as verified (only a verified e-mail links to another account).
func TestPickGitHubEmail(t *testing.T) {
	emails := []githubEmail{
		{Email: "other@example.org", Verified: true},
		{Email: "main@example.org", Primary: true, Verified: true},
		{Email: "new@example.org"},
	}
	cases := []struct {
		name     string
		public   string
		emails   []githubEmail
		want     string
		verified bool
	}{
		{"primary verified", "", emails, "main@example.org", true},
		{"public verified", "other@example.org", emails, "other@example.org", true},
		{"public unverified", "new@example.org", emails, "new@example.org", false},
		{"public unknown to the emails API", "x@example.org", emails, "x@example.org", false},
		{"public without the emails API", "x@example.org", nil, "x@example.org", false},
		{"primary unverified", "", []githubEmail{{Email: "a@example.org", Primary: true}, {Email: "b@example.org", Verified: true}}, "a@example.org", false},
		{"no primary", "", []githubEmail{{Email: "a@example.org"}, {Email: "b@example.org", Verified: true}}, "b@example.org", true},
		{"nothing verified", "", []githubEmail{{Email: "a@example.org"}}, "a@example.org", false},
		{"none", "", nil, "", false},
	}
	for _, c := range cases {
		got, verified := pickGitHubEmail(c.public, c.emails)
		if got != c.want || verified != c.verified {
			t.Errorf("%s: got %q verified=%v, want %q verified=%v", c.name, got, verified, c.want, c.verified)
		}
	}
}

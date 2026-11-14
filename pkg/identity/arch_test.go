package identity

import "testing"

// A Linux arm64 host must still claim x64, because the Linux user agent says
// x86_64. Any other pairing is a contradiction inside one request.
func TestArchForLinuxClaimsX64(t *testing.T) {
	cases := map[string]string{
		"amd64": "x64",
		"arm64": "x64",
	}
	for goarch, want := range cases {
		if got := archFor("linux", goarch); got != want {
			t.Fatalf("archFor(linux, %s) = %q, want %q", goarch, got, want)
		}
	}
}

func TestArchForOtherPlatforms(t *testing.T) {
	if got := archFor("windows", "amd64"); got != "x64" {
		t.Fatalf("windows amd64 = %q", got)
	}
	if got := archFor("windows", "arm64"); got != "arm64" {
		t.Fatalf("windows arm64 = %q", got)
	}
	if got := archFor("darwin", "arm64"); got != "arm64" {
		t.Fatalf("darwin arm64 = %q", got)
	}
}

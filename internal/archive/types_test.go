package archive

import (
	"strings"
	"testing"
	"time"
)

func TestMakeID(t *testing.T) {
	ts := time.Date(2026, 4, 10, 14, 30, 0, 0, time.UTC)
	id := MakeID("feat-login", ts)
	want := "feat-login--2026-04-10T14-30-00"
	if id != want {
		t.Errorf("MakeID = %q, want %q", id, want)
	}
}

func TestMakeIDSpecialChars(t *testing.T) {
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	id := MakeID("my-feature/test", ts)
	want := "my-feature/test--2026-01-01T00-00-00"
	if id != want {
		t.Errorf("MakeID = %q, want %q", id, want)
	}
}

// Invariant: IDs are embedded in git refs (refs/grove-archive/<ID>/<repo>)
// and appear as ls-able strings in CLI output. Colons would break both.
func TestMakeID_NoColons(t *testing.T) {
	ts := time.Date(2026, 4, 21, 14, 30, 45, 0, time.UTC)
	id := MakeID("ws", ts)
	if strings.Contains(id, ":") {
		t.Errorf("MakeID must not contain colons (unsafe for git refs): %q", id)
	}
}

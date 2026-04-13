package archive

import (
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

package cmd

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name, in string
		want     time.Duration
		wantErr  bool
	}{
		{"30 days", "30d", 30 * 24 * time.Hour, false},
		{"1 day", "1d", 24 * time.Hour, false},
		{"2 weeks", "2w", 2 * 7 * 24 * time.Hour, false},
		{"6 hours", "6h", 6 * time.Hour, false},
		{"zero days", "0d", 0, false},

		{"empty", "", 0, true},
		{"single char", "d", 0, true},
		{"unsupported minutes", "30m", 0, true},
		{"unsupported seconds", "30s", 0, true},
		{"missing unit", "30", 0, true},
		{"non-numeric", "xyzd", 0, true},
		{"unit only", "-d", 0, true}, // "%d" parse fails
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseDuration(%q) err = %v, wantErr = %v", tt.in, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

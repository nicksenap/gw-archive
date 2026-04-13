package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/nicksenap/gw-archive/internal/archive"
	"github.com/spf13/cobra"
)

var pruneOlderThan string

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove old archives",
	Run:   runPrune,
}

func init() {
	pruneCmd.Flags().StringVar(&pruneOlderThan, "older-than", "", "Remove archives older than duration (e.g., 30d, 90d)")
	pruneCmd.MarkFlagRequired("older-than")
}

func runPrune(cmd *cobra.Command, args []string) {
	dur, err := parseDuration(pruneOlderThan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid duration %q: %s\n", pruneOlderThan, err)
		os.Exit(1)
	}

	cutoff := time.Now().Add(-dur)

	archives, err := archive.LoadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	pruned := 0
	for _, a := range archives {
		t, err := time.Parse("2006-01-02T15:04:05", a.ArchivedAt)
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			// Clean up git refs
			for _, repo := range a.Repos {
				if repo.StashRef != "" {
					archive.DeleteRef(repo.SourceRepo, a.Name, repo.RepoName)
				}
			}
			if err := archive.Remove(a.ID); err != nil {
				fmt.Fprintf(os.Stderr, "warn: failed to remove %s: %s\n", a.ID, err)
				continue
			}
			pruned++
		}
	}

	fmt.Printf("Pruned %d archive(s)\n", pruned)
}

// parseDuration parses a human-friendly duration like "30d", "90d", "2w".
func parseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("too short")
	}
	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	var n int
	if _, err := fmt.Sscanf(numStr, "%d", &n); err != nil {
		return 0, err
	}

	switch unit {
	case 'd':
		return time.Duration(n) * 24 * time.Hour, nil
	case 'w':
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	case 'h':
		return time.Duration(n) * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown unit %q (use d, w, or h)", string(unit))
	}
}

package cmd

import (
	"fmt"
	"os"

	"github.com/nicksenap/gw-archive/internal/archive"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Delete an archive and clean up git refs",
	Args:  cobra.ExactArgs(1),
	Run:   runRemove,
}

func runRemove(cmd *cobra.Command, args []string) {
	a, err := archive.Find(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	if a == nil {
		fmt.Fprintf(os.Stderr, "error: archive %q not found\n", args[0])
		os.Exit(1)
	}

	// Clean up git refs
	for _, repo := range a.Repos {
		if repo.StashRef != "" {
			archive.DeleteRef(repo.SourceRepo, a.Name, repo.RepoName)
		}
	}

	// Remove from JSONL
	if err := archive.Remove(a.ID); err != nil {
		fmt.Fprintf(os.Stderr, "error: removing archive: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Removed archive %s\n", a.ID)
}

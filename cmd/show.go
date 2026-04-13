package cmd

import (
	"fmt"
	"os"

	"github.com/nicksenap/gw-archive/internal/archive"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show archive details",
	Args:  cobra.ExactArgs(1),
	Run:   runShow,
}

func runShow(cmd *cobra.Command, args []string) {
	a, err := archive.Find(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	if a == nil {
		fmt.Fprintf(os.Stderr, "error: archive %q not found\n", args[0])
		os.Exit(1)
	}

	fmt.Printf("ID:         %s\n", a.ID)
	fmt.Printf("Name:       %s\n", a.Name)
	fmt.Printf("Branch:     %s\n", a.Branch)
	fmt.Printf("Created:    %s\n", a.CreatedAt)
	fmt.Printf("Archived:   %s\n", a.ArchivedAt)
	fmt.Printf("Repos:      %d\n", len(a.Repos))
	fmt.Println()

	for _, r := range a.Repos {
		status := "clean"
		if r.HasChanges {
			status = fmt.Sprintf("changes (ref: %s)", r.StashRef[:min(12, len(r.StashRef))])
		}
		fmt.Printf("  %-20s %s\n", r.RepoName, status)
		fmt.Printf("  %-20s source: %s\n", "", r.SourceRepo)
	}
}

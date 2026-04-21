package cmd

import (
	"github.com/nicksenap/gw-archive/internal/archive"
	"github.com/spf13/cobra"
)

// completeArchiveIDs provides shell completion for archive IDs.
// Used by show, revive, and remove — all of which take a single <id> arg.
func completeArchiveIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	archives, err := archive.LoadAll()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	ids := make([]string, 0, len(archives))
	for _, a := range archives {
		ids = append(ids, a.ID)
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}

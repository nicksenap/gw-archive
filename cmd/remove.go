package cmd

import (
	"fmt"

	"github.com/nicksenap/gw-archive/internal/archive"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:               "remove <id>",
	Short:             "Delete an archive and clean up git refs",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeArchiveIDs,
	Run:               runRemove,
}

func runRemove(cmd *cobra.Command, args []string) {
	a, err := archive.Find(args[0])
	if err != nil {
		exitError("%s", err)
	}
	if a == nil {
		exitError("archive %q not found", args[0])
	}

	if err := archive.DeleteArchive(a); err != nil {
		exitError("removing archive: %s", err)
	}

	fmt.Printf("Removed archive %s\n", a.ID)
}

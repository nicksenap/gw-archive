package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "gw-archive",
	Short: "Archive and revive Grove workspaces",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(saveCmd, listCmd, showCmd, reviveCmd, removeCmd, pruneCmd)
	rootCmd.Version = Version
}

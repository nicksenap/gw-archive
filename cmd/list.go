package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/nicksenap/gw-archive/internal/archive"
	"github.com/spf13/cobra"
)

var listJSON bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List archived workspaces",
	Run:   runList,
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
}

func runList(cmd *cobra.Command, args []string) {
	archives, err := archive.LoadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	if len(archives) == 0 {
		fmt.Println("No archives")
		return
	}

	if listJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(archives)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tBRANCH\tREPOS\tCHANGES\tARCHIVED")
	// Show newest first
	for i := len(archives) - 1; i >= 0; i-- {
		a := archives[i]
		repoNames := make([]string, len(a.Repos))
		changes := 0
		for j, r := range a.Repos {
			repoNames[j] = r.RepoName
			if r.HasChanges {
				changes++
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d/%d\t%s\n",
			a.ID, a.Name, a.Branch,
			strings.Join(repoNames, ","),
			changes, len(a.Repos),
			a.ArchivedAt,
		)
	}
	w.Flush()
}

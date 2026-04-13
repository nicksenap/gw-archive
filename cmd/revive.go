package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nicksenap/gw-archive/internal/archive"
	"github.com/nicksenap/gw-archive/internal/grove"
	"github.com/spf13/cobra"
)

var reviveCmd = &cobra.Command{
	Use:   "revive <id>",
	Short: "Recreate a workspace from an archive",
	Args:  cobra.ExactArgs(1),
	Run:   runRevive,
}

func runRevive(cmd *cobra.Command, args []string) {
	a, err := archive.Find(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	if a == nil {
		fmt.Fprintf(os.Stderr, "error: archive %q not found\n", args[0])
		os.Exit(1)
	}

	// Build repo name list
	repoNames := make([]string, len(a.Repos))
	for i, r := range a.Repos {
		repoNames[i] = r.RepoName
	}

	// Recreate the workspace via gw create
	fmt.Printf("Reviving workspace %s (branch: %s, repos: %s)\n", a.Name, a.Branch, strings.Join(repoNames, ", "))

	gwArgs := []string{"create", a.Name, "--branch", a.Branch, "--repos", strings.Join(repoNames, ",")}
	gwCmd := exec.Command("gw", gwArgs...)
	gwCmd.Stdout = os.Stdout
	gwCmd.Stderr = os.Stderr
	if err := gwCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: gw create failed: %s\n", err)
		os.Exit(1)
	}

	// Read the newly created workspace to get worktree paths
	ws, err := grove.FindWorkspace(a.Name)
	if err != nil || ws == nil {
		fmt.Fprintf(os.Stderr, "warn: could not find revived workspace in state\n")
		return
	}

	// Build worktree path lookup
	wtPaths := make(map[string]string)
	for _, r := range ws.Repos {
		wtPaths[r.RepoName] = r.WorktreePath
	}

	// Apply stashed changes
	appliedCount := 0
	for _, repo := range a.Repos {
		if !repo.HasChanges || repo.StashRef == "" {
			continue
		}

		wtPath, ok := wtPaths[repo.RepoName]
		if !ok {
			fmt.Fprintf(os.Stderr, "warn: %s: worktree path not found\n", repo.RepoName)
			continue
		}

		if err := archive.ApplyStash(wtPath, repo.StashRef); err != nil {
			fmt.Fprintf(os.Stderr, "warn: %s: stash apply failed (may have conflicts): %s\n", repo.RepoName, err)
			continue
		}
		appliedCount++
		fmt.Printf("  %s: changes restored\n", repo.RepoName)
	}

	fmt.Printf("Revived %s (%d repos, %d with changes applied)\n", a.Name, len(a.Repos), appliedCount)
}

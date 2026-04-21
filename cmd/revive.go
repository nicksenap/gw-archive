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

var reviveAndRemove bool

var reviveCmd = &cobra.Command{
	Use:               "revive <id>",
	Short:             "Recreate a workspace from an archive",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeArchiveIDs,
	Run:               runRevive,
}

func init() {
	reviveCmd.Flags().BoolVar(&reviveAndRemove, "and-remove", false, "Delete the archive after a fully successful revive")
}

func runRevive(cmd *cobra.Command, args []string) {
	a, err := archive.Find(args[0])
	if err != nil {
		exitError("%s", err)
	}
	if a == nil {
		exitError("archive %q not found", args[0])
	}

	// Pre-flight: workspace name must not already exist.
	if existing, _ := grove.FindWorkspace(a.Name); existing != nil {
		exitError("workspace %q already exists; delete or rename it before reviving", a.Name)
	}

	// Pre-flight: every recorded stash ref must still resolve. If one was GC'd
	// or manually deleted we'd silently skip apply later — louder is better.
	var missing []string
	for _, repo := range a.Repos {
		if repo.HasChanges && repo.StashRef != "" && !archive.CommitExists(repo.SourceRepo, repo.StashRef) {
			missing = append(missing, repo.RepoName)
		}
	}
	if len(missing) > 0 {
		exitError("stash commits missing for: %s. Archive is unrecoverable — remove with: gw archive remove %s",
			strings.Join(missing, ", "), a.ID)
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
		exitError("gw create failed: %s (archive preserved; retry with: gw archive revive %s)", err, a.ID)
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
	expectedApplies := 0
	hadFailures := false
	for _, repo := range a.Repos {
		if !repo.HasChanges {
			continue
		}
		// HasChanges=true with no ref means the save succeeded partially — stash
		// was created but SaveRef failed. Data is lost; flag it so --and-remove
		// doesn't compound the loss by deleting the JSONL record.
		if repo.StashRef == "" {
			fmt.Fprintf(os.Stderr, "warn: %s: HasChanges=true but no stash ref (archive corrupt, data was lost at save time)\n", repo.RepoName)
			hadFailures = true
			continue
		}
		expectedApplies++

		wtPath, ok := wtPaths[repo.RepoName]
		if !ok {
			fmt.Fprintf(os.Stderr, "warn: %s: worktree path not found\n", repo.RepoName)
			hadFailures = true
			continue
		}

		if err := archive.ApplyStash(wtPath, repo.StashRef); err != nil {
			fmt.Fprintf(os.Stderr, "warn: %s: stash apply failed (may have conflicts): %s\n", repo.RepoName, err)
			hadFailures = true
			continue
		}
		appliedCount++
		fmt.Printf("  %s: changes restored\n", repo.RepoName)
	}

	fmt.Printf("Revived %s (%d repos, %d/%d changes applied)\n", a.Name, len(a.Repos), appliedCount, expectedApplies)

	if reviveAndRemove {
		if hadFailures {
			fmt.Fprintln(os.Stderr, "warn: --and-remove skipped: revive had failures, archive preserved for retry")
			return
		}
		if err := archive.DeleteArchive(a); err != nil {
			fmt.Fprintf(os.Stderr, "warn: --and-remove failed: %s\n", err)
			return
		}
		fmt.Printf("Archive %s removed\n", a.ID)
	}
}

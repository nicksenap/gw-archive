package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/nicksenap/gw-archive/internal/archive"
	"github.com/nicksenap/gw-archive/internal/grove"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save <name> <path> <branch>",
	Short: "Archive a workspace before deletion",
	Args:  cobra.ExactArgs(3),
	Run:   runSave,
}

func runSave(cmd *cobra.Command, args []string) {
	wsName := args[0]
	// args[1] is path (available from hook but we read state for full repo info)
	branch := args[2]

	ws, err := grove.FindWorkspace(wsName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: reading state: %s\n", err)
		os.Exit(1)
	}
	if ws == nil {
		fmt.Fprintf(os.Stderr, "error: workspace %q not found in state\n", wsName)
		os.Exit(1)
	}

	now := time.Now()
	a := archive.Archive{
		ID:         archive.MakeID(wsName, now),
		Name:       wsName,
		Branch:     branch,
		CreatedAt:  ws.CreatedAt,
		ArchivedAt: now.Format("2006-01-02T15:04:05"),
	}

	changedCount := 0
	for _, repo := range ws.Repos {
		ar := archive.ArchivedRepo{
			RepoName:   repo.RepoName,
			SourceRepo: repo.SourceRepo,
		}

		// Create stash commit (captures staged + unstaged + untracked)
		sha, err := archive.StashCreate(repo.WorktreePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: %s: stash create failed: %s\n", repo.RepoName, err)
			a.Repos = append(a.Repos, ar)
			continue
		}

		if sha != "" {
			// Save the stash commit as a custom ref in the source repo
			if err := archive.SaveRef(repo.SourceRepo, wsName, repo.RepoName, sha); err != nil {
				fmt.Fprintf(os.Stderr, "warn: %s: save ref failed: %s\n", repo.RepoName, err)
			} else {
				ar.StashRef = sha
				ar.HasChanges = true
				changedCount++
			}
		}
		a.Repos = append(a.Repos, ar)
	}

	if err := archive.Append(a); err != nil {
		fmt.Fprintf(os.Stderr, "error: saving archive: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Archived %s (%d repos, %d with changes)\n", wsName, len(a.Repos), changedCount)
}

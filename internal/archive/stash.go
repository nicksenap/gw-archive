package archive

import (
	"fmt"
	"os/exec"
	"strings"
)

// refPrefix is the namespace for grove archive refs.
const refPrefix = "refs/grove-archive"

// RefName returns the full ref path for an archived repo.
func RefName(wsName, repoName string) string {
	return fmt.Sprintf("%s/%s/%s", refPrefix, wsName, repoName)
}

// StashCreate creates a stash commit in the given worktree directory
// capturing staged, unstaged, and untracked changes.
// Returns the SHA (empty string if working tree is clean).
func StashCreate(worktreePath string) (string, error) {
	// Try with --include-untracked first (captures everything)
	out, err := runGit(worktreePath, "stash", "create", "--include-untracked")
	if err != nil {
		return "", fmt.Errorf("stash create: %w", err)
	}
	sha := strings.TrimSpace(out)
	if sha != "" {
		return sha, nil
	}

	// git stash create --include-untracked returns empty when only untracked
	// files exist. Check if there are untracked files and stage them temporarily.
	untracked, _ := runGit(worktreePath, "ls-files", "--others", "--exclude-standard")
	if strings.TrimSpace(untracked) == "" {
		return "", nil // truly clean
	}

	// Stage everything, create stash, then unstage
	if _, err := runGit(worktreePath, "add", "-A"); err != nil {
		return "", fmt.Errorf("staging for stash: %w", err)
	}
	out, err = runGit(worktreePath, "stash", "create")
	if err != nil {
		runGit(worktreePath, "reset") // best-effort unstage
		return "", fmt.Errorf("stash create: %w", err)
	}
	// Unstage so the worktree looks the same as before
	runGit(worktreePath, "reset")

	return strings.TrimSpace(out), nil
}

// SaveRef stores a stash SHA as a custom ref in the source repo.
func SaveRef(sourceRepo, wsName, repoName, sha string) error {
	ref := RefName(wsName, repoName)
	_, err := runGit(sourceRepo, "update-ref", ref, sha)
	if err != nil {
		return fmt.Errorf("saving ref %s: %w", ref, err)
	}
	return nil
}

// DeleteRef removes a custom archive ref from the source repo.
func DeleteRef(sourceRepo, wsName, repoName string) error {
	ref := RefName(wsName, repoName)
	_, err := runGit(sourceRepo, "update-ref", "-d", ref)
	if err != nil {
		// Ref may already be gone — not fatal
		return nil
	}
	return nil
}

// ApplyStash applies a stash ref in the given worktree.
func ApplyStash(worktreePath, sha string) error {
	_, err := runGit(worktreePath, "stash", "apply", sha)
	if err != nil {
		return fmt.Errorf("stash apply: %w", err)
	}
	return nil
}

// ResolveRef reads the SHA stored in a custom archive ref.
func ResolveRef(sourceRepo, wsName, repoName string) (string, error) {
	ref := RefName(wsName, repoName)
	out, err := runGit(sourceRepo, "rev-parse", ref)
	if err != nil {
		return "", fmt.Errorf("resolving ref %s: %w", ref, err)
	}
	return strings.TrimSpace(out), nil
}

// CommitExists returns true if the given SHA is reachable in the source repo.
// Use this before revive to detect stash commits that were GC'd or manually deleted.
func CommitExists(sourceRepo, sha string) bool {
	if sha == "" {
		return false
	}
	_, err := runGit(sourceRepo, "cat-file", "-e", sha)
	return err == nil
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Do not return captured output as the first value on error — it's stderr,
		// and callers (e.g. ls-files) would treat stderr as a valid result.
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

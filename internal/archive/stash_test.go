package archive

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTestRepo creates a git repo with an initial commit.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bare := filepath.Join(dir, "origin.git")
	clone := filepath.Join(dir, "repo")

	run(t, dir, "git", "init", "--bare", bare)
	run(t, dir, "git", "clone", bare, clone)
	run(t, clone, "git", "config", "user.email", "test@test.com")
	run(t, clone, "git", "config", "user.name", "Test")
	os.WriteFile(filepath.Join(clone, "README.md"), []byte("# test"), 0o644)
	run(t, clone, "git", "add", ".")
	run(t, clone, "git", "commit", "-m", "initial")
	run(t, clone, "git", "push", "origin", "HEAD")

	return clone
}

func run(t *testing.T, dir, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "GIT_DIR=") ||
			strings.HasPrefix(kv, "GIT_WORK_TREE=") ||
			strings.HasPrefix(kv, "GIT_INDEX_FILE=") {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %s\n%s", name, strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestRefName(t *testing.T) {
	got := RefName("my-ws", "backend")
	want := "refs/grove-archive/my-ws/backend"
	if got != want {
		t.Errorf("RefName = %q, want %q", got, want)
	}
}

func TestStashCreateCleanRepo(t *testing.T) {
	repo := initTestRepo(t)

	sha, err := StashCreate(repo)
	if err != nil {
		t.Fatalf("StashCreate: %v", err)
	}
	if sha != "" {
		t.Errorf("expected empty SHA for clean repo, got %q", sha)
	}
}

func TestStashCreateWithTrackedChanges(t *testing.T) {
	repo := initTestRepo(t)

	// Modify a tracked file
	os.WriteFile(filepath.Join(repo, "README.md"), []byte("# modified"), 0o644)

	sha, err := StashCreate(repo)
	if err != nil {
		t.Fatalf("StashCreate: %v", err)
	}
	if sha == "" {
		t.Fatal("expected non-empty SHA for dirty repo")
	}
}

func TestStashCreateWithUntrackedOnly(t *testing.T) {
	repo := initTestRepo(t)

	// Add only untracked files
	os.WriteFile(filepath.Join(repo, "new-file.txt"), []byte("untracked"), 0o644)

	sha, err := StashCreate(repo)
	if err != nil {
		t.Fatalf("StashCreate: %v", err)
	}
	if sha == "" {
		t.Fatal("expected non-empty SHA for repo with untracked files")
	}

	// Verify repo is still clean (we unstaged after stash create)
	out := run(t, repo, "git", "diff", "--cached", "--stat")
	if out != "" {
		t.Errorf("expected clean index after StashCreate, got: %s", out)
	}
}

func TestSaveRefAndResolveRef(t *testing.T) {
	repo := initTestRepo(t)

	// Create a change and stash it
	os.WriteFile(filepath.Join(repo, "README.md"), []byte("# changed"), 0o644)
	sha, err := StashCreate(repo)
	if err != nil || sha == "" {
		t.Fatalf("StashCreate: sha=%q err=%v", sha, err)
	}

	// Save as custom ref
	if err := SaveRef(repo, "test-ws", "myrepo", sha); err != nil {
		t.Fatalf("SaveRef: %v", err)
	}

	// Resolve it back
	resolved, err := ResolveRef(repo, "test-ws", "myrepo")
	if err != nil {
		t.Fatalf("ResolveRef: %v", err)
	}
	if resolved != sha {
		t.Errorf("ResolveRef = %q, want %q", resolved, sha)
	}

	// Verify it's under the custom namespace
	out := run(t, repo, "git", "for-each-ref", "refs/grove-archive/")
	if !strings.Contains(out, "refs/grove-archive/test-ws/myrepo") {
		t.Errorf("ref not found in for-each-ref output: %s", out)
	}

	// Verify it's NOT visible in regular commands
	branches := run(t, repo, "git", "branch", "-a")
	if strings.Contains(branches, "grove-archive") {
		t.Error("archive ref should not appear in git branch output")
	}
	tags := run(t, repo, "git", "tag")
	if strings.Contains(tags, "grove-archive") {
		t.Error("archive ref should not appear in git tag output")
	}
}

func TestDeleteRef(t *testing.T) {
	repo := initTestRepo(t)

	os.WriteFile(filepath.Join(repo, "README.md"), []byte("# changed"), 0o644)
	sha, _ := StashCreate(repo)
	SaveRef(repo, "ws", "repo", sha)

	// Delete it
	if err := DeleteRef(repo, "ws", "repo"); err != nil {
		t.Fatalf("DeleteRef: %v", err)
	}

	// Should be gone
	out := run(t, repo, "git", "for-each-ref", "refs/grove-archive/")
	if strings.Contains(out, "grove-archive") {
		t.Error("ref should be deleted")
	}
}

func TestDeleteRefIdempotent(t *testing.T) {
	repo := initTestRepo(t)

	// Deleting a non-existent ref should not error
	if err := DeleteRef(repo, "nonexistent", "repo"); err != nil {
		t.Fatalf("DeleteRef on missing ref should not error: %v", err)
	}
}

func TestApplyStash(t *testing.T) {
	repo := initTestRepo(t)

	// Make a change and stash it
	os.WriteFile(filepath.Join(repo, "README.md"), []byte("# stashed change"), 0o644)
	sha, _ := StashCreate(repo)
	if sha == "" {
		t.Fatal("expected stash SHA")
	}

	// Reset the file to original
	run(t, repo, "git", "checkout", "README.md")

	// Verify file is back to original
	content, _ := os.ReadFile(filepath.Join(repo, "README.md"))
	if string(content) == "# stashed change" {
		t.Fatal("file should be reverted before apply")
	}

	// Apply the stash
	if err := ApplyStash(repo, sha); err != nil {
		t.Fatalf("ApplyStash: %v", err)
	}

	// Verify the change is back
	content, _ = os.ReadFile(filepath.Join(repo, "README.md"))
	if string(content) != "# stashed change" {
		t.Errorf("after ApplyStash, content = %q", string(content))
	}
}

func TestFullRoundTrip(t *testing.T) {
	repo := initTestRepo(t)

	// Make changes (tracked + untracked)
	os.WriteFile(filepath.Join(repo, "README.md"), []byte("# modified for archive"), 0o644)
	os.WriteFile(filepath.Join(repo, "new-file.txt"), []byte("new content"), 0o644)

	// Stash
	sha, err := StashCreate(repo)
	if err != nil || sha == "" {
		t.Fatalf("StashCreate: sha=%q err=%v", sha, err)
	}

	// Save ref
	SaveRef(repo, "roundtrip-ws", "myrepo", sha)

	// Simulate deletion: reset all changes
	run(t, repo, "git", "checkout", "README.md")
	os.Remove(filepath.Join(repo, "new-file.txt"))

	// Resolve and apply
	resolved, _ := ResolveRef(repo, "roundtrip-ws", "myrepo")
	if resolved != sha {
		t.Fatalf("resolved SHA mismatch")
	}

	ApplyStash(repo, resolved)

	// Verify tracked change restored
	content, _ := os.ReadFile(filepath.Join(repo, "README.md"))
	if string(content) != "# modified for archive" {
		t.Errorf("tracked change not restored: %q", string(content))
	}

	// Clean up ref
	DeleteRef(repo, "roundtrip-ws", "myrepo")
	out := run(t, repo, "git", "for-each-ref", "refs/grove-archive/")
	if strings.Contains(out, "roundtrip-ws") {
		t.Error("ref should be cleaned up")
	}
}

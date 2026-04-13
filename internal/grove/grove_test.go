package grove

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	t.Setenv("GROVE_DIR", "/custom/grove")
	if got := Dir(); got != "/custom/grove" {
		t.Errorf("Dir() = %q, want /custom/grove", got)
	}
}

func TestDirDefault(t *testing.T) {
	t.Setenv("GROVE_DIR", "")
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".grove")
	if got := Dir(); got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}

func TestStatePath(t *testing.T) {
	t.Setenv("GROVE_STATE", "/custom/state.json")
	if got := StatePath(); got != "/custom/state.json" {
		t.Errorf("StatePath() = %q", got)
	}
}

func TestStatePathDefault(t *testing.T) {
	t.Setenv("GROVE_STATE", "")
	t.Setenv("GROVE_DIR", "/mygrove")
	want := "/mygrove/state.json"
	if got := StatePath(); got != want {
		t.Errorf("StatePath() = %q, want %q", got, want)
	}
}

func TestLoadStateEmpty(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GROVE_STATE", filepath.Join(dir, "state.json"))

	ws, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if len(ws) != 0 {
		t.Errorf("expected 0 workspaces, got %d", len(ws))
	}
}

func TestLoadState(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "state.json")
	t.Setenv("GROVE_STATE", stateFile)

	data := `[{"name":"ws1","path":"/ws/1","branch":"main","created_at":"2026-01-01","repos":[{"repo_name":"api","source_repo":"/repos/api","worktree_path":"/ws/1/api","branch":"main"}]}]`
	os.WriteFile(stateFile, []byte(data), 0o644)

	ws, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if len(ws) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(ws))
	}
	if ws[0].Name != "ws1" {
		t.Errorf("Name = %q", ws[0].Name)
	}
	if len(ws[0].Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(ws[0].Repos))
	}
	if ws[0].Repos[0].RepoName != "api" {
		t.Errorf("RepoName = %q", ws[0].Repos[0].RepoName)
	}
}

func TestFindWorkspace(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "state.json")
	t.Setenv("GROVE_STATE", stateFile)

	data := `[{"name":"alpha","path":"/ws/alpha","branch":"main","repos":[]},{"name":"beta","path":"/ws/beta","branch":"dev","repos":[]}]`
	os.WriteFile(stateFile, []byte(data), 0o644)

	ws, err := FindWorkspace("beta")
	if err != nil {
		t.Fatalf("FindWorkspace: %v", err)
	}
	if ws == nil {
		t.Fatal("expected to find beta")
	}
	if ws.Branch != "dev" {
		t.Errorf("Branch = %q", ws.Branch)
	}

	ws, err = FindWorkspace("nonexistent")
	if err != nil {
		t.Fatalf("FindWorkspace: %v", err)
	}
	if ws != nil {
		t.Error("expected nil for nonexistent workspace")
	}
}

// Package grove reads Grove state and config files.
// This package has zero dependencies on the grove CLI — it reads the files directly.
package grove

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Dir returns the Grove directory, preferring the GROVE_DIR env var.
func Dir() string {
	if d := os.Getenv("GROVE_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".grove")
}

// StatePath returns the path to state.json.
func StatePath() string {
	if p := os.Getenv("GROVE_STATE"); p != "" {
		return p
	}
	return filepath.Join(Dir(), "state.json")
}

// Workspace is a workspace from state.json.
type Workspace struct {
	Name      string         `json:"name"`
	Path      string         `json:"path"`
	Branch    string         `json:"branch"`
	CreatedAt string         `json:"created_at"`
	Repos     []RepoWorktree `json:"repos"`
}

// RepoWorktree is a single repo's worktree within a workspace.
type RepoWorktree struct {
	RepoName     string `json:"repo_name"`
	SourceRepo   string `json:"source_repo"`
	WorktreePath string `json:"worktree_path"`
	Branch       string `json:"branch"`
}

// LoadState reads state.json.
func LoadState() ([]Workspace, error) {
	data, err := os.ReadFile(StatePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	var workspaces []Workspace
	if err := json.Unmarshal(data, &workspaces); err != nil {
		return nil, err
	}
	return workspaces, nil
}

// FindWorkspace finds a workspace by name.
func FindWorkspace(name string) (*Workspace, error) {
	workspaces, err := LoadState()
	if err != nil {
		return nil, err
	}
	for i := range workspaces {
		if workspaces[i].Name == name {
			return &workspaces[i], nil
		}
	}
	return nil, nil
}

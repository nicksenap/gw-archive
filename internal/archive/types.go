package archive

import "time"

// Archive represents a single archived workspace.
type Archive struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Branch     string       `json:"branch"`
	CreatedAt  string       `json:"created_at"`
	ArchivedAt string       `json:"archived_at"`
	Repos      []ArchivedRepo `json:"repos"`
}

// ArchivedRepo holds the stash state for one repo in the archive.
type ArchivedRepo struct {
	RepoName   string `json:"repo_name"`
	SourceRepo string `json:"source_repo"`
	StashRef   string `json:"stash_ref"`
	HasChanges bool   `json:"has_changes"`
}

// MakeID generates an archive ID from a workspace name and timestamp.
func MakeID(name string, t time.Time) string {
	return name + "--" + t.Format("2006-01-02T15-04-05")
}

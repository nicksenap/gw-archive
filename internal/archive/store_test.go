package archive

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicksenap/gw-archive/internal/grove"
)

// setupTestStore creates a temporary GROVE_DIR and returns a cleanup function.
func setupTestStore(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("GROVE_DIR", dir)
}

func TestAppendAndLoadAll(t *testing.T) {
	setupTestStore(t)

	a1 := Archive{
		ID: "ws1--2026-04-10T10-00-00", Name: "ws1", Branch: "feat/a",
		ArchivedAt: "2026-04-10T10:00:00",
		Repos: []ArchivedRepo{
			{RepoName: "backend", SourceRepo: "/repos/backend", StashRef: "abc123", HasChanges: true},
		},
	}
	a2 := Archive{
		ID: "ws2--2026-04-10T11-00-00", Name: "ws2", Branch: "feat/b",
		ArchivedAt: "2026-04-10T11:00:00",
		Repos: []ArchivedRepo{
			{RepoName: "frontend", SourceRepo: "/repos/frontend", HasChanges: false},
		},
	}

	if err := Append(a1); err != nil {
		t.Fatalf("Append a1: %v", err)
	}
	if err := Append(a2); err != nil {
		t.Fatalf("Append a2: %v", err)
	}

	all, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 archives, got %d", len(all))
	}
	if all[0].ID != a1.ID {
		t.Errorf("first archive ID = %q, want %q", all[0].ID, a1.ID)
	}
	if all[1].ID != a2.ID {
		t.Errorf("second archive ID = %q, want %q", all[1].ID, a2.ID)
	}
	if !all[0].Repos[0].HasChanges {
		t.Error("first repo should have changes")
	}
}

func TestFind(t *testing.T) {
	setupTestStore(t)

	a := Archive{ID: "find-me--2026-01-01T00-00-00", Name: "find-me", Branch: "main"}
	Append(a)

	found, err := Find("find-me--2026-01-01T00-00-00")
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find archive")
	}
	if found.Name != "find-me" {
		t.Errorf("Name = %q, want %q", found.Name, "find-me")
	}

	notFound, err := Find("nonexistent")
	if err != nil {
		t.Fatalf("Find nonexistent: %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for nonexistent archive")
	}
}

func TestRemove(t *testing.T) {
	setupTestStore(t)

	Append(Archive{ID: "keep", Name: "keep", Branch: "main"})
	Append(Archive{ID: "remove-me", Name: "remove-me", Branch: "feat/x"})
	Append(Archive{ID: "also-keep", Name: "also-keep", Branch: "main"})

	if err := Remove("remove-me"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	all, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 archives after remove, got %d", len(all))
	}
	for _, a := range all {
		if a.ID == "remove-me" {
			t.Error("removed archive should not be present")
		}
	}
}

func TestDeleteArchive(t *testing.T) {
	setupTestStore(t)

	// Populate store with two archives.
	a := Archive{ID: "del-me", Name: "del-me", Branch: "main",
		Repos: []ArchivedRepo{{RepoName: "backend", SourceRepo: "/repos/none", StashRef: "abc123"}},
	}
	Append(a)
	Append(Archive{ID: "keeper", Name: "keeper", Branch: "main"})

	// SourceRepo doesn't exist; DeleteRef is best-effort, so DeleteArchive must still remove the JSONL entry.
	if err := DeleteArchive(&a); err != nil {
		t.Fatalf("DeleteArchive: %v", err)
	}

	all, _ := LoadAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 archive after DeleteArchive, got %d", len(all))
	}
	if all[0].ID != "keeper" {
		t.Errorf("remaining archive ID = %q, want %q", all[0].ID, "keeper")
	}
}

func TestLoadAllEmpty(t *testing.T) {
	setupTestStore(t)

	all, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 archives, got %d", len(all))
	}
}

func TestLoadAllCorruptLine(t *testing.T) {
	setupTestStore(t)

	// Write a valid line then a corrupt line
	path := filepath.Join(grove.Dir(), "archives.jsonl")
	os.MkdirAll(filepath.Dir(path), 0o755)
	data := `{"id":"good","name":"good","branch":"main","repos":[]}` + "\n" +
		`{corrupt json` + "\n" +
		`{"id":"also-good","name":"also-good","branch":"main","repos":[]}` + "\n"
	os.WriteFile(path, []byte(data), 0o644)

	all, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	// Should skip corrupt line and load the valid ones
	if len(all) != 2 {
		t.Errorf("expected 2 archives (skipping corrupt), got %d", len(all))
	}
}

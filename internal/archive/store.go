package archive

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/nicksenap/gw-archive/internal/grove"
)

// StorePath returns the path to archives.jsonl.
func StorePath() string {
	return filepath.Join(grove.Dir(), "archives.jsonl")
}

// Append adds an archive entry to the JSONL store.
// Uses an advisory file lock (flock) so concurrent `gw archive save` invocations
// cannot interleave writes — large entries (many repos, long paths) can exceed
// PIPE_BUF and would otherwise be split across write(2) syscalls without mutual exclusion.
func Append(a Archive) error {
	path := StorePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close() // releases the flock on macOS/Linux

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("locking %s: %w", path, err)
	}

	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = f.Write(data)
	return err
}

// LoadAll reads all archive entries from the JSONL store.
func LoadAll() ([]Archive, error) {
	f, err := os.Open(StorePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var archives []Archive
	scanner := bufio.NewScanner(f)
	// Increase buffer for large lines
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var a Archive
		if err := json.Unmarshal(line, &a); err != nil {
			fmt.Fprintf(os.Stderr, "warn: archives.jsonl line %d is corrupt, skipping: %s\n", lineNo, err)
			continue
		}
		archives = append(archives, a)
	}
	return archives, scanner.Err()
}

// Find returns the archive with the given ID, or nil.
func Find(id string) (*Archive, error) {
	archives, err := LoadAll()
	if err != nil {
		return nil, err
	}
	for i := range archives {
		if archives[i].ID == id {
			return &archives[i], nil
		}
	}
	return nil, nil
}

// DeleteArchive removes all of an archive's git refs and its entry from the JSONL store.
// Ref cleanup is best-effort (matches prune / remove semantics); store removal is fatal on error.
func DeleteArchive(a *Archive) error {
	for _, repo := range a.Repos {
		if repo.StashRef != "" {
			DeleteRef(repo.SourceRepo, a.Name, repo.RepoName)
		}
	}
	return Remove(a.ID)
}

// Remove deletes the archive entry with the given ID by rewriting the file.
func Remove(id string) error {
	archives, err := LoadAll()
	if err != nil {
		return err
	}

	path := StorePath()
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, a := range archives {
		if a.ID == id {
			continue
		}
		data, err := json.Marshal(a)
		if err != nil {
			continue
		}
		f.Write(data)
		f.Write([]byte{'\n'})
	}
	f.Close()
	return os.Rename(tmp, path)
}

package archive

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/nicksenap/gw-archive/internal/grove"
)

// StorePath returns the path to archives.jsonl.
func StorePath() string {
	return filepath.Join(grove.Dir(), "archives.jsonl")
}

// Append adds an archive entry to the JSONL store.
func Append(a Archive) error {
	path := StorePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

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
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var a Archive
		if err := json.Unmarshal(line, &a); err != nil {
			continue // skip corrupt lines
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

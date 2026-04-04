package store

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
)

// ObjectStore implements port.ObjectStorage using the filesystem.
// Objects are stored at: baseDir/<hash[:2]>/<hash[2:]>
type ObjectStore struct {
	baseDir string
}

// NewObjectStore creates a new ObjectStore rooted at baseDir.
func NewObjectStore(baseDir string) *ObjectStore {
	os.MkdirAll(baseDir, 0755)
	return &ObjectStore{baseDir: baseDir}
}

func (s *ObjectStore) objectPath(hash string) string {
	return filepath.Join(s.baseDir, hash[:2], hash[2:])
}

// Exists checks whether an object with the given hash exists.
func (s *ObjectStore) Exists(hash string) (bool, error) {
	_, err := os.Stat(s.objectPath(hash))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// Read returns the content of the object identified by hash.
func (s *ObjectStore) Read(hash string) ([]byte, error) {
	return os.ReadFile(s.objectPath(hash))
}

// Write stores the data at the content-addressed path.
// Writing the same hash twice is idempotent (overwrites).
func (s *ObjectStore) Write(hash string, data []byte) error {
	path := s.objectPath(hash)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Delete removes the object file from disk.
func (s *ObjectStore) Delete(hash string) error {
	return os.Remove(s.objectPath(hash))
}

// Walk iterates over all stored objects and calls fn for each hash.
// Hash is returned in the format "ab/cdef..." (slash-separated).
func (s *ObjectStore) Walk(fn func(hash string) error) error {
	return filepath.WalkDir(s.baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(s.baseDir, path)
		if err != nil {
			return err
		}
		hash := filepath.ToSlash(rel) // normalize to forward slash
		return fn(hash)
	})
}

// ComputeHash returns the SHA-256 hex digest of data.
func ComputeHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

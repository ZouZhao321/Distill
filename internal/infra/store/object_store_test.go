package store

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestObjectStore_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	s := NewObjectStore(dir)

	data := []byte("hello distill")
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	// Write
	err := s.Write(hashStr, data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify file exists at correct path
	expectedPath := filepath.Join(dir, hashStr[:2], hashStr[2:])
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("object file not found at %s", expectedPath)
	}

	// Read
	got, err := s.Read(hashStr)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(got) != "hello distill" {
		t.Errorf("Read got %q, want %q", string(got), "hello distill")
	}
}

func TestObjectStore_Exists(t *testing.T) {
	dir := t.TempDir()
	s := NewObjectStore(dir)

	data := []byte("exists test")
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	exists, _ := s.Exists(hashStr)
	if exists {
		t.Error("should not exist before write")
	}

	s.Write(hashStr, data)

	exists, _ = s.Exists(hashStr)
	if !exists {
		t.Error("should exist after write")
	}
}

func TestObjectStore_Delete(t *testing.T) {
	dir := t.TempDir()
	s := NewObjectStore(dir)

	data := []byte("delete me")
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	s.Write(hashStr, data)
	err := s.Delete(hashStr)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	exists, _ := s.Exists(hashStr)
	if exists {
		t.Error("should not exist after delete")
	}
}

func TestObjectStore_Walk(t *testing.T) {
	dir := t.TempDir()
	s := NewObjectStore(dir)

	// Write 3 objects
	for _, content := range []string{"aaa", "bbb", "ccc"} {
		data := []byte(content)
		hash := sha256.Sum256(data)
		h := hex.EncodeToString(hash[:])
		s.Write(h, data)
	}

	// Walk
	var walked []string
	s.Walk(func(hash string) error {
		walked = append(walked, hash)
		return nil
	})

	if len(walked) != 3 {
		t.Errorf("Walk returned %d items, want 3", len(walked))
	}
}

func TestObjectStore_WriteDuplicate_Idempotent(t *testing.T) {
	dir := t.TempDir()
	s := NewObjectStore(dir)

	data := []byte("duplicate")
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	s.Write(hashStr, data)
	err := s.Write(hashStr, data)
	if err != nil {
		t.Errorf("duplicate write should be idempotent, got: %v", err)
	}
}

func TestObjectStore_Read_NotFound(t *testing.T) {
	dir := t.TempDir()
	s := NewObjectStore(dir)

	_, err := s.Read("nonexistent00000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Error("Read should return error for nonexistent object")
	}
}

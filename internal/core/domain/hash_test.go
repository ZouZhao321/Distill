package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// expectedHash 辅助函数：用标准库计算 SHA-256 hex 摘要作为期望值。
func expectedHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func TestComputeHash_EmptyInput(t *testing.T) {
	got := ComputeHash([]byte{})
	want := expectedHash([]byte{})
	if got != want {
		t.Errorf("ComputeHash([]byte{}) = %q, want %q", got, want)
	}
}

func TestComputeHash_KnownInput(t *testing.T) {
	data := []byte("hello world")
	got := ComputeHash(data)
	want := expectedHash(data)
	if got != want {
		t.Errorf("ComputeHash(%q) = %q, want %q", string(data), got, want)
	}
}

func TestComputeHash_BinaryData(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0xff, 0xfe}
	got := ComputeHash(data)
	want := expectedHash(data)
	if got != want {
		t.Errorf("ComputeHash(binary) = %q, want %q", got, want)
	}
}

func TestComputeHash_ConsistentLength(t *testing.T) {
	got := ComputeHash([]byte("anything"))
	if len(got) != 64 {
		t.Errorf("ComputeHash result length = %d, want 64", len(got))
	}
}

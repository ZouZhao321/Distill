package domain

import (
	"crypto/sha256"
	"encoding/hex"
)

// ComputeHash 返回数据的 SHA-256 十六进制摘要。
func ComputeHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

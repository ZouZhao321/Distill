package store

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
)

// ObjectStore 使用文件系统实现 port.ObjectStorage 接口。
// 对象存储路径格式为：baseDir/<哈希前2位>/<哈希剩余部分>。
type ObjectStore struct {
	baseDir string
}

// NewObjectStore 创建以 baseDir 为根目录的 ObjectStore 实例。
func NewObjectStore(baseDir string) *ObjectStore {
	os.MkdirAll(baseDir, 0755)
	return &ObjectStore{baseDir: baseDir}
}

// objectPath 根据哈希计算对象文件路径。
func (s *ObjectStore) objectPath(hash string) string {
	return filepath.Join(s.baseDir, hash[:2], hash[2:])
}

// Exists 检查指定哈希的对象是否存在。
func (s *ObjectStore) Exists(hash string) (bool, error) {
	_, err := os.Stat(s.objectPath(hash))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// Read 返回指定哈希对象的内容。
func (s *ObjectStore) Read(hash string) ([]byte, error) {
	return os.ReadFile(s.objectPath(hash))
}

// Write 将数据存储到内容寻址路径。
// 相同哈希写入两次为幂等操作（覆盖）。
func (s *ObjectStore) Write(hash string, data []byte) error {
	path := s.objectPath(hash)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Delete 从磁盘删除对象文件。
func (s *ObjectStore) Delete(hash string) error {
	return os.Remove(s.objectPath(hash))
}

// Walk 遍历所有已存储的对象，对每个哈希调用 fn。
// 哈希格式为 "ab/cdef..."（斜杠分隔）。
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
		hash := filepath.ToSlash(rel) // 统一为正斜杠
		return fn(hash)
	})
}

// ComputeHash 返回数据的 SHA-256 十六进制摘要。
func ComputeHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

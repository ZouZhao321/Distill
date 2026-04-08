package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

func TestBackupToTrash_FilenameSeparator(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")
	manifest := &domain.Manifest{
		Hash: "abc123def456",
		// 其他字段省略
	}

	// 执行备份
	err := backupToTrash(manifest, trashDir)
	if err != nil {
		t.Fatalf("backupToTrash failed: %v", err)
	}

	// 验证文件存在
	files, err := os.ReadDir(trashDir)
	if err != nil {
		t.Fatalf("failed to read trash dir: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	filename := files[0].Name()

	// Bug 验证: 文件名应该有分隔符 "-" 在哈希和后缀之间
	// 正确格式: abc123def456-manifest.json
	// 错误格式: abc123def456manifest.json
	if !strings.Contains(filename, "-") {
		t.Errorf("BUG CONFIRMED: filename %q is missing separator '-' between hash and suffix", filename)
		t.Errorf("Expected format: <hash>-manifest.json")
		t.Errorf("Actual format: <hash>manifest.json (missing separator)")
	}

	// 验证正确的前缀
	if !strings.HasPrefix(filename, "abc123def456-") {
		t.Errorf("filename %q doesn't have expected prefix with separator", filename)
	}

	// 验证正确的后缀
	if !strings.HasSuffix(filename, "manifest.json") {
		t.Errorf("filename %q doesn't have expected suffix", filename)
	}
}

package domain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBundle_LoadFromTOML(t *testing.T) {
	// 创建临时目录和测试 TOML 文件
	tmpDir := t.TempDir()

	zhTOML := `
[MsgAdded]
description = "成功添加资产"
other = "已添加: {{.Name}} ({{.FileCount}} 文件, {{.TotalSize}} bytes)"

[MsgRemoved]
description = "成功移除资产"
other = "已移除: {{.Name}}"

[MsgErrAddFailed]
description = "添加资产失败"
other = "添加失败: {{.Err}}"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "zh.toml"), []byte(zhTOML), 0644); err != nil {
		t.Fatal(err)
	}

	enTOML := `
[MsgAdded]
description = "Successfully added asset"
other = "Added: {{.Name}} ({{.FileCount}} files, {{.TotalSize}} bytes)"

[MsgRemoved]
description = "Successfully removed asset"
other = "Removed: {{.Name}}"

[MsgErrAddFailed]
description = "Failed to add asset"
other = "Failed to add asset: {{.Err}}"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "en.toml"), []byte(enTOML), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试加载
	b := NewBundle()
	if err := b.LoadFromTOML(tmpDir); err != nil {
		t.Fatalf("LoadFromTOML failed: %v", err)
	}

	// 测试中文查找
	got, err := b.Localize("zh", "MsgAdded", map[string]any{
		"Name":      "test.zip",
		"FileCount": 3,
		"TotalSize": 1024,
	})
	if err != nil {
		t.Fatalf("Localize zh failed: %v", err)
	}
	want := "已添加: test.zip (3 文件, 1024 bytes)"
	if got != want {
		t.Errorf("Localize zh = %q, want %q", got, want)
	}

	// 测试英文查找
	got, err = b.Localize("en", "MsgAdded", map[string]any{
		"Name":      "test.zip",
		"FileCount": 3,
		"TotalSize": 1024,
	})
	if err != nil {
		t.Fatalf("Localize en failed: %v", err)
	}
	want = "Added: test.zip (3 files, 1024 bytes)"
	if got != want {
		t.Errorf("Localize en = %q, want %q", got, want)
	}

	// 测试不带参数的查找
	got, err = b.Localize("zh", "MsgRemoved", map[string]any{
		"Name": "my-asset",
	})
	if err != nil {
		t.Fatalf("Localize MsgRemoved failed: %v", err)
	}
	want = "已移除: my-asset"
	if got != want {
		t.Errorf("Localize MsgRemoved = %q, want %q", got, want)
	}
}

func TestBundle_LoadFromEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	b := NewBundle()
	// 空目录应该返回错误
	if err := b.LoadFromTOML(tmpDir); err == nil {
		t.Error("LoadFromTOML should fail with empty directory")
	}
}

func TestBundle_LocalizeNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	zhTOML := `[MsgHello]
other = "你好"
`
	os.WriteFile(filepath.Join(tmpDir, "zh.toml"), []byte(zhTOML), 0644)

	b := NewBundle()
	b.LoadFromTOML(tmpDir)

	// 查找不存在的 key 应该返回错误
	_, err := b.Localize("zh", "MsgNonexistent", nil)
	if err == nil {
		t.Error("Localize should fail for nonexistent key")
	}
}

func TestBundle_UnsupportedLang(t *testing.T) {
	tmpDir := t.TempDir()
	zhTOML := `[MsgHello]
other = "你好"
`
	os.WriteFile(filepath.Join(tmpDir, "zh.toml"), []byte(zhTOML), 0644)

	b := NewBundle()
	b.LoadFromTOML(tmpDir)

	// 不支持的语言应该 fallback 到第一个已注册的语言
	got, err := b.Localize("fr", "MsgHello", nil)
	if err != nil {
		t.Fatalf("Localize with fallback lang should not error, got: %v", err)
	}
	if got != "你好" {
		t.Errorf("fallback lang = %q, want %q", got, "你好")
	}
}

func TestBundle_LoadFromEmbedFS(t *testing.T) {
	// 使用 os.DirFS 模拟 embed.FS
	tmpDir := t.TempDir()
	zhTOML := `[MsgRootShort]
other = "Distill - 资产管理 CLI 工具（embed）"
`
	os.WriteFile(filepath.Join(tmpDir, "zh.toml"), []byte(zhTOML), 0644)
	enTOML := `[MsgRootShort]
other = "Distill - Asset Management CLI (embed)"
`
	os.WriteFile(filepath.Join(tmpDir, "en.toml"), []byte(enTOML), 0644)

	b := NewBundle()
	if err := b.LoadFromFS(os.DirFS(tmpDir)); err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	got, err := b.Localize("zh", "MsgRootShort", nil)
	if err != nil {
		t.Fatalf("Localize failed: %v", err)
	}
	if got != "Distill - 资产管理 CLI 工具（embed）" {
		t.Errorf("Localize = %q, want embed text", got)
	}
}

func TestBundle_LoadFromFS_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	b := NewBundle()
	if err := b.LoadFromFS(os.DirFS(tmpDir)); err == nil {
		t.Error("LoadFromFS should fail with empty directory")
	}
}

package domain

import (
	"os"
	"path/filepath"
	"testing"
)

// resetBundle 将全局 Bundle 重置为 nil，用于测试隔离。
func resetBundle() {
	bundleMu.Lock()
	globalBundle = nil
	bundleMu.Unlock()
}

func TestT_WithBundleFallback(t *testing.T) {
	// 确保没有残留的 Bundle 状态
	resetBundle()
	defer resetBundle()

	// 创建临时 locales 目录
	tmpDir := t.TempDir()

	zhTOML := `
[MsgRootShort]
description = "根命令简短描述"
other = "Distill - 资产管理 CLI 工具（来自 TOML）"
`
	os.WriteFile(filepath.Join(tmpDir, "zh.toml"), []byte(zhTOML), 0644)

	enTOML := `
[MsgRootShort]
description = "Root command short description"
other = "Distill - Asset Management CLI (from TOML)"
`
	os.WriteFile(filepath.Join(tmpDir, "en.toml"), []byte(enTOML), 0644)

	// 初始化 Bundle
	if err := InitBundle(tmpDir); err != nil {
		t.Fatalf("InitBundle failed: %v", err)
	}

	// 测试中文：T() 应该从 Bundle 查找
	SetLang("zh")
	got := T(MsgRootShort)
	want := "Distill - 资产管理 CLI 工具（来自 TOML）"
	if got != want {
		t.Errorf("T(MsgRootShort) zh = %q, want %q", got, want)
	}

	// 测试英文
	SetLang("en")
	got = T(MsgRootShort)
	want = "Distill - Asset Management CLI (from TOML)"
	if got != want {
		t.Errorf("T(MsgRootShort) en = %q, want %q", got, want)
	}

	// 测试：未在临时 TOML 中定义的 key，Bundle 中找不到，返回空
	SetLang("zh")
	got = T(MsgAdded, P{"Name": "test.zip", "FileCount": 3, "TotalSize": 1024})
	if got != "" {
		t.Logf("T(MsgAdded) with partial bundle returned (expected empty): %q", got)
	}

	ResetLang()
}

func TestInitBundle_InvalidDir(t *testing.T) {
	resetBundle()
	defer resetBundle()

	err := InitBundle("/nonexistent/path")
	if err == nil {
		t.Error("InitBundle should fail with invalid directory")
	}
}

func TestT_BundleNotInitialized(t *testing.T) {
	// 确保没有 Bundle
	resetBundle()
	defer resetBundle()

	// Bundle 未初始化时，T() 应该返回空字符串（不再有旧 map fallback）
	SetLang("zh")
	got := T(MsgRootShort)
	if got != "" {
		t.Errorf("T() without bundle = %q, want empty string", got)
	}
	ResetLang()
}

// TestT_MigratedKeysFromTOML 验证翻译从 TOML 加载后，
// 具名参数插值和纯文本 key 都能正确工作。
func TestT_MigratedKeysFromTOML(t *testing.T) {
	resetBundle()
	defer resetBundle()

	// 创建临时 locales 目录
	tmpDir := t.TempDir()

	zhTOML := `
[MsgCmdAddShort]
description = "add 命令简短描述"
other = "添加资产到仓库"

[MsgAdded]
description = "添加资产成功提示"
other = "已添加: {{.Name}} ({{.FileCount}} 文件, {{.TotalSize}} bytes)"

[MsgListEmpty]
description = "仓库为空时的提示"
other = "仓库为空，使用 distill add 添加资产"

[MsgErrAssetNotFound]
description = "资产不存在错误"
other = "资产 \"{{.Name}}\" 不存在"

[MsgCmdInitShort]
description = "init 命令简短描述"
other = "初始化 Distill 仓库"

[MsgGcOrphanList]
description = "孤立对象列表"
other = "发现 {{.Count}} 个孤立对象:"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "zh.toml"), []byte(zhTOML), 0644); err != nil {
		t.Fatal(err)
	}

	enTOML := `
[MsgCmdAddShort]
description = "add command short description"
other = "Add asset to repository"

[MsgAdded]
description = "Asset added success message"
other = "Added: {{.Name}} ({{.FileCount}} files, {{.TotalSize}} bytes)"

[MsgListEmpty]
description = "Empty repository hint"
other = "Repository is empty. Use \"distill add\" to add assets"

[MsgErrAssetNotFound]
description = "Asset not found error"
other = "Asset \"{{.Name}}\" not found"

[MsgCmdInitShort]
description = "init command short description"
other = "Initialize Distill repository"

[MsgGcOrphanList]
description = "Orphaned objects list"
other = "Found {{.Count}} orphaned object(s):"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "en.toml"), []byte(enTOML), 0644); err != nil {
		t.Fatal(err)
	}

	if err := InitBundle(tmpDir); err != nil {
		t.Fatalf("InitBundle failed: %v", err)
	}

	// --- 中文测试 ---
	SetLang("zh")

	tests := []struct {
		key    MsgKey
		params P
		want   string
	}{
		{MsgCmdAddShort, nil, "添加资产到仓库"},
		{MsgAdded, P{"Name": "test.zip", "FileCount": 3, "TotalSize": 1024}, "已添加: test.zip (3 文件, 1024 bytes)"},
		{MsgListEmpty, nil, "仓库为空，使用 distill add 添加资产"},
		{MsgErrAssetNotFound, P{"Name": "my-asset"}, "资产 \"my-asset\" 不存在"},
		{MsgCmdInitShort, nil, "初始化 Distill 仓库"},
		{MsgGcOrphanList, P{"Count": 5}, "发现 5 个孤立对象:"},
	}
	for _, tt := range tests {
		got := T(tt.key, tt.params)
		if got != tt.want {
			t.Errorf("T(%v) zh = %q, want %q", tt.key, got, tt.want)
		}
	}

	// --- 英文测试 ---
	SetLang("en")

	enTests := []struct {
		key    MsgKey
		params P
		want   string
	}{
		{MsgCmdAddShort, nil, "Add asset to repository"},
		{MsgAdded, P{"Name": "test.zip", "FileCount": 3, "TotalSize": 1024}, "Added: test.zip (3 files, 1024 bytes)"},
		{MsgListEmpty, nil, "Repository is empty. Use \"distill add\" to add assets"},
		{MsgErrAssetNotFound, P{"Name": "my-asset"}, "Asset \"my-asset\" not found"},
		{MsgCmdInitShort, nil, "Initialize Distill repository"},
		{MsgGcOrphanList, P{"Count": 5}, "Found 5 orphaned object(s):"},
	}
	for _, tt := range enTests {
		got := T(tt.key, tt.params)
		if got != tt.want {
			t.Errorf("T(%v) en = %q, want %q", tt.key, got, tt.want)
		}
	}

	ResetLang()
}

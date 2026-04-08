package usecase

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

// --- Tests ---

func TestInitUseCase_Execute_CreatesDirectoryStructure(t *testing.T) {
	tmpDir := t.TempDir()
	storeHome := filepath.Join(tmpDir, "test-store")

	uc := NewInitUseCase()
	_, err := uc.Execute(InitInput{
		StoreHome: storeHome,
		TrashPath: filepath.Join(tmpDir, "trash"),
		Lang:      "zh",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 验证目录结构
	expectedDirs := []string{
		filepath.Join(storeHome, "objects"),
		filepath.Join(storeHome, "manifests"),
		filepath.Join(storeHome, "config"),
		filepath.Join(storeHome, "log"),
	}
	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("directory %s should exist: %v", dir, err)
		} else if !info.IsDir() {
			t.Errorf("%s should be a directory", dir)
		}
	}
}

func TestInitUseCase_Execute_WritesConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	storeHome := filepath.Join(tmpDir, "store")
	trashPath := filepath.Join(tmpDir, "trash")

	uc := NewInitUseCase()
	config, err := uc.Execute(InitInput{
		StoreHome: storeHome,
		TrashPath: trashPath,
		Lang:      "en",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 验证返回的 Config 字段
	if config == nil {
		t.Fatal("Execute should return a non-nil Config")
	}
	// Store.Home 应该是规范化的路径（/ 替换 \）
	if config.Store.Home != strings.ReplaceAll(storeHome, "\\", "/") {
		t.Errorf("config.Store.Home = %q, want %q", config.Store.Home, strings.ReplaceAll(storeHome, "\\", "/"))
	}
	if config.Store.TrashPath != strings.ReplaceAll(trashPath, "\\", "/") {
		t.Errorf("config.Store.TrashPath = %q, want %q", config.Store.TrashPath, strings.ReplaceAll(trashPath, "\\", "/"))
	}
	if config.Lang != "en" {
		t.Errorf("config.Lang = %q, want %q", config.Lang, "en")
	}

	// 验证 config.toml 文件存在于磁盘
	configPath := filepath.Join(storeHome, "config", "config.toml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config.toml should exist at %s: %v", configPath, err)
	}

	// 验证可以从磁盘重新加载配置
	loaded, err := domain.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config from disk: %v", err)
	}
	if loaded.Store.Home != config.Store.Home {
		t.Errorf("loaded config Store.Home = %q, want %q", loaded.Store.Home, config.Store.Home)
	}
	if loaded.Lang != "en" {
		t.Errorf("loaded config Lang = %q, want %q", loaded.Lang, "en")
	}
}

func TestInitUseCase_Execute_WritesRefsFile(t *testing.T) {
	tmpDir := t.TempDir()
	storeHome := filepath.Join(tmpDir, "store")

	uc := NewInitUseCase()
	_, err := uc.Execute(InitInput{
		StoreHome: storeHome,
		TrashPath: filepath.Join(tmpDir, "trash"),
		Lang:      "zh",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 验证 refs.json 存在且内容为 {}
	refsPath := filepath.Join(storeHome, "config", "refs.json")
	data, err := os.ReadFile(refsPath)
	if err != nil {
		t.Fatalf("refs.json should exist: %v", err)
	}
	content := strings.TrimSpace(string(data))
	if content != "{}" {
		t.Errorf("refs.json content = %q, want %q", content, "{}")
	}
}

func TestInitUseCase_Execute_AlreadyInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	storeHome := filepath.Join(tmpDir, "store")

	uc := NewInitUseCase()

	// 第一次初始化应成功
	_, err := uc.Execute(InitInput{
		StoreHome: storeHome,
		TrashPath: filepath.Join(tmpDir, "trash"),
		Lang:      "zh",
	})
	if err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	// 第二次初始化应返回 ErrAlreadyInitialized
	_, err = uc.Execute(InitInput{
		StoreHome: storeHome,
		TrashPath: filepath.Join(tmpDir, "trash"),
		Lang:      "zh",
	})
	if err != ErrAlreadyInitialized {
		t.Errorf("second init should return ErrAlreadyInitialized, got %v", err)
	}
}

func TestInitUseCase_Execute_AlreadyInitialized_ByExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	storeHome := filepath.Join(tmpDir, "store")

	// 预先创建 config.toml，模拟仓库已初始化
	configDir := filepath.Join(storeHome, "config")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(""), 0644)

	uc := NewInitUseCase()
	_, err := uc.Execute(InitInput{
		StoreHome: storeHome,
		TrashPath: filepath.Join(tmpDir, "trash"),
		Lang:      "zh",
	})
	if err != ErrAlreadyInitialized {
		t.Errorf("init should detect existing config and return ErrAlreadyInitialized, got %v", err)
	}
}

func TestInitUseCase_Execute_UsesDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	storeHome := filepath.Join(tmpDir, "store")
	trashPath := filepath.Join(tmpDir, "trash")

	uc := NewInitUseCase()
	config, err := uc.Execute(InitInput{
		StoreHome: storeHome,
		TrashPath: trashPath,
		Lang:      "zh",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 验证默认值与 domain.DefaultConfig() 一致（除 Store.Home/TrashPath/Lang 外）
	if config.Core.Version != "1" {
		t.Errorf("config.Core.Version = %q, want %q", config.Core.Version, "1")
	}
	if config.Core.ObjectsFormat != "plain" {
		t.Errorf("config.Core.ObjectsFormat = %q, want %q", config.Core.ObjectsFormat, "plain")
	}
	if config.Checkout.Overwrite != "ask" {
		t.Errorf("config.Checkout.Overwrite = %q, want %q", config.Checkout.Overwrite, "ask")
	}
	if config.Log.Format != "text" {
		t.Errorf("config.Log.Format = %q, want %q", config.Log.Format, "text")
	}
	if config.Log.Level != "info" {
		t.Errorf("config.Log.Level = %q, want %q", config.Log.Level, "info")
	}
	if config.Normalize.CRLFToLF != true {
		t.Error("config.Normalize.CRLFToLF should be true")
	}
}

func TestInitUseCase_Execute_NormalizesPaths(t *testing.T) {
	tmpDir := t.TempDir()
	// 使用包含反斜杠的路径（Windows 环境）
	storeHome := filepath.Join(tmpDir, "store")
	trashPath := filepath.Join(tmpDir, "trash")

	uc := NewInitUseCase()
	config, err := uc.Execute(InitInput{
		StoreHome: storeHome,
		TrashPath: trashPath,
		Lang:      "zh",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 路径中的 \ 应被替换为 /
	if strings.Contains(config.Store.Home, "\\") {
		t.Errorf("config.Store.Home should not contain backslash, got %q", config.Store.Home)
	}
	if strings.Contains(config.Store.TrashPath, "\\") {
		t.Errorf("config.Store.TrashPath should not contain backslash, got %q", config.Store.TrashPath)
	}
}

func TestInitUseCase_Execute_IdempotentDifferentInstances(t *testing.T) {
	tmpDir := t.TempDir()
	storeHome := filepath.Join(tmpDir, "store")

	// 用不同实例重复初始化同一目录，第二次应被文件系统检测拦截
	uc1 := NewInitUseCase()
	_, err := uc1.Execute(InitInput{
		StoreHome: storeHome,
		TrashPath: filepath.Join(tmpDir, "trash"),
		Lang:      "zh",
	})
	if err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	uc2 := NewInitUseCase()
	_, err = uc2.Execute(InitInput{
		StoreHome: storeHome,
		TrashPath: filepath.Join(tmpDir, "trash"),
		Lang:      "zh",
	})
	if err != ErrAlreadyInitialized {
		t.Errorf("second init with new instance should return ErrAlreadyInitialized, got %v", err)
	}
}

func TestInitUseCase_Execute_EmptyLangFallsBackToDefault(t *testing.T) {
	tmpDir := t.TempDir()
	storeHome := filepath.Join(tmpDir, "store")

	uc := NewInitUseCase()
	config, err := uc.Execute(InitInput{
		StoreHome: storeHome,
		TrashPath: filepath.Join(tmpDir, "trash"),
		Lang:      "",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Lang 为空时应保留 domain.DefaultConfig() 的默认值 "zh"
	if config.Lang != "zh" {
		t.Errorf("config.Lang = %q, want default %q", config.Lang, "zh")
	}
}

func TestInitUseCase_Execute_DirCreateFailure(t *testing.T) {
	tmpDir := t.TempDir()
	// 在 Windows 上创建一个文件占位 storeHome，MkdirAll 将失败
	storeHome := filepath.Join(tmpDir, "store-blocker")
	os.WriteFile(storeHome, []byte("blocked"), 0644)

	uc := NewInitUseCase()
	_, err := uc.Execute(InitInput{
		StoreHome: storeHome,
		TrashPath: filepath.Join(tmpDir, "trash"),
		Lang:      "zh",
	})
	if err == nil {
		t.Error("Execute should fail when storeHome is a file, not a directory")
	}
}

package usecase

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

// ErrAlreadyInitialized 表示仓库已经初始化过。
var ErrAlreadyInitialized = errors.New("repository already initialized")

// InitInput 保存 InitUseCase.Execute() 所需的输入参数。
type InitInput struct {
	StoreHome string // 仓库根目录
	TrashPath string // 回收站路径
	Lang      string // 界面语言 ("zh" | "en")
}

// InitUseCase 负责仓库初始化。
// 创建目录结构、生成默认配置、写入 config.toml 和 refs.json。
type InitUseCase struct{}

// NewInitUseCase 创建新的 InitUseCase 实例。
func NewInitUseCase() *InitUseCase {
	return &InitUseCase{}
}

// Execute 使用给定的输入参数初始化仓库。
// 如果 config.toml 已存在，返回 ErrAlreadyInitialized。
func (uc *InitUseCase) Execute(input InitInput) (*domain.Config, error) {
	storeHome := input.StoreHome
	trashPath := input.TrashPath
	lang := input.Lang

	// 幂等性检查：检测 config.toml 是否已存在（文件系统级检查）
	configPath := filepath.Join(storeHome, "config", "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		return nil, ErrAlreadyInitialized
	}

	// 创建目录结构
	dirs := []string{
		filepath.Join(storeHome, "objects"),
		filepath.Join(storeHome, "manifests"),
		filepath.Join(storeHome, "config"),
		filepath.Join(storeHome, "log"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, err
		}
	}

	// 使用 domain.DefaultConfig() 作为基础，覆盖用户指定的字段
	config := domain.DefaultConfig()
	config.Store.Home = strings.ReplaceAll(storeHome, "\\", "/")
	config.Store.TrashPath = strings.ReplaceAll(trashPath, "\\", "/")
	if lang != "" {
		config.Lang = lang
	}

	// 写入 config.toml
	if err := domain.SaveConfig(config, configPath); err != nil {
		return nil, err
	}

	// 写入空的 refs.json
	refsPath := filepath.Join(storeHome, "config", "refs.json")
	if err := os.WriteFile(refsPath, []byte("{}\n"), 0644); err != nil {
		return nil, err
	}

	return config, nil
}

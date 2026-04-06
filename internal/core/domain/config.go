// Package domain 定义 Distill 的核心领域实体、错误类型和配置结构。
package domain

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config 表示仓库级别的配置。
type Config struct {
	Core struct {
		Version       string `toml:"version"`
		ObjectsFormat string `toml:"objects_format"`
	} `toml:"core"`

	Store struct {
		Home      string `toml:"home"`
		TrashPath string `toml:"trash_path"`
	} `toml:"store"`

	Checkout struct {
		Overwrite string `toml:"overwrite"` // "ask" | "skip" | "force"
	} `toml:"checkout"`

	Log struct {
		Format string `toml:"format"` // "text" | "json"
		Level  string `toml:"level"`  // "debug" | "info" | "warn" | "error"
	} `toml:"log"`

	Normalize struct {
		CRLFToLF bool `toml:"crlf_to_lf"`
	} `toml:"normalize"`

	Lang string `toml:"lang"` // "zh" | "en"
}

// DefaultConfig 返回带有默认值的 Config 实例。
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	config := &Config{}
	config.Core.Version = "1"
	config.Core.ObjectsFormat = "plain"
	config.Store.Home = filepath.Join(home, ".distill")
	config.Store.TrashPath = filepath.Join(home, ".distill-trash")
	config.Checkout.Overwrite = "ask"
	config.Log.Format = "text"
	config.Log.Level = "info"
	config.Normalize.CRLFToLF = true
	config.Lang = "zh"
	return config
}

// LoadConfig 从指定路径加载 TOML 配置文件。
// 如果文件不存在或读取失败，返回空 Config 和 nil（不视为错误）。
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	config := &Config{}
	if _, err := toml.Decode(string(data), config); err != nil {
		return nil, err
	}
	return config, nil
}

// LoadConfigByHome 根据 storeHome 路径加载配置文件。
func LoadConfigByHome(storeHome string) (*Config, error) {
	return LoadConfig(filepath.Join(storeHome, "config", "config.toml"))
}

// SaveConfig 将配置写入指定路径的 TOML 文件。
func SaveConfig(config *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 使用 BurntSushi/toml 的 Marshal 保持格式一致
	var buf strings.Builder
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(config); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(buf.String()), 0644)
}

// SaveConfigByHome 根据 storeHome 路径保存配置文件。
func SaveConfigByHome(config *Config, storeHome string) error {
	return SaveConfig(config, filepath.Join(storeHome, "config", "config.toml"))
}

// ConfigKeyMap 定义点号分隔的配置 key 到 Config struct 字段的映射。
// 用于 config get/set 命令。
var ConfigKeyMap = map[string]string{
	"core.version":         "core.version",
	"core.objects_format":  "core.objects_format",
	"store.home":           "store.home",
	"store.trash_path":     "store.trash_path",
	"checkout.overwrite":   "checkout.overwrite",
	"log.format":           "log.format",
	"log.level":            "log.level",
	"normalize.crlf_to_lf": "normalize.crlf_to_lf",
	"lang":                 "lang",
}

// GetConfigValue 通过点号分隔的 key 从 Config 中获取值。
func GetConfigValue(config *Config, key string) (string, error) {
	switch key {
	case "core.version":
		return config.Core.Version, nil
	case "core.objects_format":
		return config.Core.ObjectsFormat, nil
	case "store.home":
		return config.Store.Home, nil
	case "store.trash_path":
		return config.Store.TrashPath, nil
	case "checkout.overwrite":
		return config.Checkout.Overwrite, nil
	case "log.format":
		return config.Log.Format, nil
	case "log.level":
		return config.Log.Level, nil
	case "normalize.crlf_to_lf":
		return fmt.Sprintf("%v", config.Normalize.CRLFToLF), nil
	case "lang":
		return config.Lang, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// SetConfigValue 通过点号分隔的 key 设置 Config 中的值。
func SetConfigValue(config *Config, key, value string) error {
	switch key {
	case "core.version":
		config.Core.Version = value
	case "core.objects_format":
		config.Core.ObjectsFormat = value
	case "store.home":
		config.Store.Home = value
	case "store.trash_path":
		config.Store.TrashPath = value
	case "checkout.overwrite":
		config.Checkout.Overwrite = value
	case "log.format":
		config.Log.Format = value
	case "log.level":
		config.Log.Level = value
	case "normalize.crlf_to_lf":
		config.Normalize.CRLFToLF = strings.ToLower(value) == "true"
	case "lang":
		config.Lang = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// AllConfigKeys 返回所有可用的配置 key 列表。
func AllConfigKeys() []string {
	return []string{
		"core.version",
		"core.objects_format",
		"store.home",
		"store.trash_path",
		"checkout.overwrite",
		"log.format",
		"log.level",
		"normalize.crlf_to_lf",
		"lang",
	}
}

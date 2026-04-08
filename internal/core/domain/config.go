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

// ConfigMeta 定义配置项的元数据，用于校验和展示。
type ConfigMeta struct {
	Key         string   // 配置键，如 "checkout.overwrite"
	Description string   // 配置项的中文描述
	Default     string   // 默认值
	ValidValues []string // 合法的可选值，空列表表示任意字符串
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

// ConfigMetaMap 定义所有配置项的元数据。
var ConfigMetaMap = map[string]ConfigMeta{
	"core.version": {
		Key:         "core.version",
		Description: "配置格式版本号",
		Default:     "1",
		ValidValues: []string{"1"},
	},
	"core.objects_format": {
		Key:         "core.objects_format",
		Description: "对象存储格式",
		Default:     "plain",
		ValidValues: []string{"plain"},
	},
	"store.home": {
		Key:         "store.home",
		Description: "仓库存储根目录",
		Default:     "~/.distill",
		ValidValues: []string{},
	},
	"store.trash_path": {
		Key:         "store.trash_path",
		Description: "回收站目录路径",
		Default:     "~/.distill-trash",
		ValidValues: []string{},
	},
	"checkout.overwrite": {
		Key:         "checkout.overwrite",
		Description: "导出时的覆盖策略",
		Default:     "ask",
		ValidValues: []string{"ask", "skip", "force"},
	},
	"log.format": {
		Key:         "log.format",
		Description: "日志输出格式",
		Default:     "text",
		ValidValues: []string{"text", "json"},
	},
	"log.level": {
		Key:         "log.level",
		Description: "日志级别",
		Default:     "info",
		ValidValues: []string{"debug", "info", "warn", "error"},
	},
	"normalize.crlf_to_lf": {
		Key:         "normalize.crlf_to_lf",
		Description: "导出时将 CRLF 转换为 LF",
		Default:     "true",
		ValidValues: []string{"true", "false"},
	},
	"lang": {
		Key:         "lang",
		Description: "界面语言",
		Default:     "zh",
		ValidValues: []string{"zh", "en"},
	},
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

// ValidateConfigValue 校验配置值是否合法。
func ValidateConfigValue(key, value string) error {
	meta, ok := ConfigMetaMap[key]
	if !ok {
		return nil // 未定义的 key 不校验
	}

	if len(meta.ValidValues) == 0 {
		return nil // 无可选值限制，不校验
	}

	for _, v := range meta.ValidValues {
		if value == v {
			return nil // 匹配到合法值
		}
	}

	// 构建可选值提示
	validStr := ""
	for i, v := range meta.ValidValues {
		if i > 0 {
			validStr += ", "
		}
		validStr += v
	}

	return fmt.Errorf("配置项 %s 的值 %q 无效，可选值: %s", key, value, validStr)
}

// SetConfigValue 通过点号分隔的 key 设置 Config 中的值。
func SetConfigValue(config *Config, key, value string) error {
	// 先校验值是否合法
	if err := ValidateConfigValue(key, value); err != nil {
		return err
	}

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

// FormatConfigItem 格式化单个配置项的输出，包含值、可选值和说明。
// 输出格式：
//
//	key=value [valid1, valid2, ...]
//	  description
//
// 如果没有可选值（路径类配置），则不显示可选值部分。
// 如果 key 未知，只输出 key=value。
func FormatConfigItem(key, value string) string {
	var sb strings.Builder

	// 第一行：key=value，如有可选值则追加
	sb.WriteString(key)
	sb.WriteByte('=')
	sb.WriteString(value)

	if meta, ok := ConfigMetaMap[key]; ok {
		if len(meta.ValidValues) > 0 {
			sb.WriteString(" [")
			for i, v := range meta.ValidValues {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(v)
			}
			sb.WriteByte(']')
		}
	}

	// 第二行：说明（缩进两个空格）
	if meta, ok := ConfigMetaMap[key]; ok && meta.Description != "" {
		sb.WriteByte('\n')
		sb.WriteString("  ")
		sb.WriteString(meta.Description)
	}

	return sb.String()
}

// FormatAllConfig 格式化所有配置项的输出，每个配置项之间用空行分隔。
func FormatAllConfig(config *Config) string {
	var sb strings.Builder
	keys := AllConfigKeys()
	for i, key := range keys {
		value, _ := GetConfigValue(config, key)
		sb.WriteString(FormatConfigItem(key, value))
		if i < len(keys)-1 {
			sb.WriteString("\n\n")
		}
	}
	return sb.String()
}

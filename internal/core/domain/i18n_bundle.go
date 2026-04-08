package domain

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

// globalBundle 是包级的 go-i18n Bundle 实例。
// 通过 InitBundle 初始化，为空时 T() 函数 fallback 到旧 map。
var (
	globalBundle *Bundle
	bundleMu     sync.RWMutex
)

// Bundle 封装 go-i18n 的 Bundle，提供 TOML 翻译文件的加载和查找。
type Bundle struct {
	inner *i18n.Bundle
}

// NewBundle 创建一个新 Bundle 实例，默认语言为中文。
func NewBundle() *Bundle {
	bundle := i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	return &Bundle{inner: bundle}
}

// LoadFromFS 从 fs.FS 中加载所有 .toml 翻译文件。
// 支持 os.DirFS 和 embed.FS。
func (b *Bundle) LoadFromFS(fsys fs.FS) error {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("read locale directory: %w", err)
	}

	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".toml" {
			continue
		}
		data, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			return fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		_, err = b.inner.ParseMessageFileBytes(data, entry.Name())
		if err != nil {
			return fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		loaded++
	}

	if loaded == 0 {
		return fmt.Errorf("no .toml files found in fs")
	}
	return nil
}

// LoadFromTOML 从指定目录加载所有 .toml 翻译文件。
// 文件名格式应为 {lang}.toml（如 zh.toml, en.toml）。
func (b *Bundle) LoadFromTOML(dir string) error {
	return b.LoadFromFS(os.DirFS(dir))
}

// Localize 根据语言和消息 ID 查找翻译。
// data 是模板变量，对应 go-i18n 的 TemplateData。
func (b *Bundle) Localize(lang, messageID string, data map[string]any) (string, error) {
	localizer := i18n.NewLocalizer(b.inner, lang)
	cfg := &i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	}
	return localizer.Localize(cfg)
}

// InitBundle 从指定目录加载 TOML 翻译文件并设置为全局 Bundle。
// 加载后 T() 函数将优先从 Bundle 查找，找不到则 fallback 到旧 map。
func InitBundle(localesDir string) error {
	b := NewBundle()
	if err := b.LoadFromTOML(localesDir); err != nil {
		return err
	}
	bundleMu.Lock()
	globalBundle = b
	bundleMu.Unlock()
	return nil
}

// InitBundleFromFS 从 fs.FS 加载翻译文件并设置为全局 Bundle。
// 用于 embed.FS 打包的场景。
func InitBundleFromFS(fsys fs.FS) error {
	b := NewBundle()
	if err := b.LoadFromFS(fsys); err != nil {
		return err
	}
	bundleMu.Lock()
	globalBundle = b
	bundleMu.Unlock()
	return nil
}

// getBundle 返回全局 Bundle，用于 T() 函数内部查找。
// 返回 nil 表示未初始化。
func getBundle() *Bundle {
	bundleMu.RLock()
	defer bundleMu.RUnlock()
	return globalBundle
}

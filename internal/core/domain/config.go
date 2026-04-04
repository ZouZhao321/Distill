// Package domain 定义 Distill 的核心领域实体、错误类型和配置结构。
package domain

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
}

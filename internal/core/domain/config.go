package domain

// Config represents the repository-level configuration.
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

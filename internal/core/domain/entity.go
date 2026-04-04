package domain

// Object represents a content-addressed object stored in the objects/ directory.
type Object struct {
	Hash string // SHA-256 hex string
	Size int64
}

// TreeNode represents a node in the manifest directory tree.
type TreeNode struct {
	Name     string     `json:"name"`     // file or directory name
	Type     string     `json:"type"`     // "file" or "directory"
	Size     int64      `json:"size"`     // file size (only for file type)
	Object   string     `json:"object"`   // object path, e.g. "ab/cdef..." (only for file type)
	Children []TreeNode `json:"children"` // child nodes (only for directory type)
}

// Manifest represents the complete metadata for an imported asset.
type Manifest struct {
	Hash         string   `json:"hash"`          // SHA-256 of manifest identity
	OriginalName string   `json:"original_name"` // user-specified name
	OriginalPath string   `json:"original_path"` // source path
	CreatedAt    string   `json:"created_at"`    // ISO 8601 timestamp
	FileCount    int      `json:"file_count"`    // total number of files
	TotalSize    int64    `json:"total_size"`    // original total size
	StoredSize   int64    `json:"stored_size"`   // actual storage size after dedup
	Status       string   `json:"status"`        // "active" or "trashed"
	Tree         TreeNode `json:"tree"`          // root directory tree node
}

// Ref represents a name-to-manifest-hash mapping.
type Ref struct {
	Name     string `json:"name"`
	Manifest string `json:"manifest"`
}

package domain

// Object 表示存储在 objects/ 目录下的内容寻址对象。
type Object struct {
	Hash string // SHA-256 十六进制字符串
	Size int64
}

// TreeNode 表示清单目录树中的一个节点。
type TreeNode struct {
	Name     string     `json:"name"`     // 文件或目录名称
	Type     string     `json:"type"`     // "file" 或 "directory"
	Size     int64      `json:"size"`     // 文件大小（仅文件类型有效）
	Object   string     `json:"object"`   // 对象路径，如 "ab/cdef..."（仅文件类型有效）
	Children []TreeNode `json:"children"` // 子节点（仅目录类型有效）
}

// Manifest 表示已导入资产的完整元数据。
type Manifest struct {
	Hash         string   `json:"hash"`          // 清单身份的 SHA-256 哈希
	OriginalName string   `json:"original_name"` // 用户指定的资产名称
	OriginalPath string   `json:"original_path"` // 来源路径
	CreatedAt    string   `json:"created_at"`    // ISO 8601 时间戳
	FileCount    int      `json:"file_count"`    // 文件总数
	TotalSize    int64    `json:"total_size"`    // 原始总大小
	StoredSize   int64    `json:"stored_size"`   // 去重后实际存储大小
	Status       string   `json:"status"`        // "active" 或 "trashed"
	Tree         TreeNode `json:"tree"`          // 根目录树节点
}

// Ref 表示资产名称到清单哈希的映射关系。
type Ref struct {
	Name     string `json:"name"`
	Manifest string `json:"manifest"`
}

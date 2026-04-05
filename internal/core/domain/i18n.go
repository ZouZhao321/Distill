// 提供 Distill CLI 的国际化支持。
// 支持中文（zh，默认）和英文（en）两种语言。
package domain

import (
	"fmt"
	"strings"
	"sync"
)

// Lang 表示支持的语言。
type Lang int

const (
	LangZh Lang = iota // 中文（默认）
	LangEn             // 英文
)

// MsgKey 是文案的唯一标识。
type MsgKey int

// 所有文案 key 的定义。
const (
	// 根命令
	MsgRootShort MsgKey = iota
	MsgRootLong
	MsgFlagHome
	MsgFlagLogFormat
	MsgFlagLogLevel
	MsgFlagLang

	// add 命令
	MsgCmdAddShort
	MsgCmdAddLong
	MsgFlagAs

	// checkout 命令
	MsgCmdCheckoutShort
	MsgCmdCheckoutLong
	MsgFlagOutput
	MsgFlagOverwrite

	// export 命令
	MsgCmdExportShort
	MsgCmdExportLong

	// gc 命令
	MsgCmdGcShort
	MsgCmdGcLong
	MsgFlagDryRun

	// init 命令
	MsgCmdInitShort
	MsgCmdInitLong
	MsgFlagTrash

	// list 命令
	MsgCmdListShort
	MsgCmdListLong
	MsgFlagFormat

	// remove 命令
	MsgCmdRemoveShort
	MsgCmdRemoveLong

	// 运行时输出
	MsgAdded
	MsgCheckedOut
	MsgExported
	MsgRemoved
	MsgGcNoOrphans
	MsgGcOrphanList
	MsgGcClean
	MsgGcAlreadyClean
	MsgInited
	MsgTrashPath
	MsgListEmpty

	// 错误信息
	MsgErrCannotAccess
	MsgErrNotRegularFile
	MsgErrCannotOpen
	MsgErrReadFailed
	MsgErrAddFailed
	MsgErrReadZipFailed
	MsgErrReadDirFailed
	MsgErrCheckoutFailed
	MsgErrExportFailed
	MsgErrGcDryRunFailed
	MsgErrGcFailed
	MsgErrCreateDirFailed
	MsgErrWriteConfigFailed
	MsgErrWriteRefsFailed
	MsgErrListFailed
	MsgErrAssetNotFound
	MsgErrReadManifestFailed
	MsgErrRemoveFailed
	MsgErrTrashBackupFailed

	// cobra 内置命令
	MsgCmdHelpShort
	MsgCmdCompletionShort

	// cobra 内置 flag
	MsgFlagHelp

	// cobra 帮助模板
	MsgHelpTip
	MsgFlagDefault

	// 交互提示
	MsgCheckoutFileExists
	MsgCheckoutOverwritePrompt
	MsgCheckoutSkipped
	MsgWarnLogOpenFailed
)

var (
	currentLang Lang
	langMu      sync.RWMutex
)

func init() {
	currentLang = LangZh // 默认中文
}

// zh 翻译表。
var zh = map[MsgKey]string{
	// 根命令
	MsgRootShort:     "Distill - 资产管理 CLI 工具",
	MsgRootLong:      "Distill 是一个 Go 语言 CLI 工具，用于资产的导入、管理和导出。基于内容寻址存储实现物理级去重。",
	MsgFlagHome:      "仓库路径",
	MsgFlagLogFormat: "日志格式 (text|json)",
	MsgFlagLogLevel:  "日志级别 (debug|info|warn|error)",
	MsgFlagLang:      "界面语言 (zh|en)",

	// add
	MsgCmdAddShort: "添加资产到仓库",
	MsgCmdAddLong:  "将文件、文件夹或 ZIP 包添加到 Distill 仓库。基于 SHA-256 内容寻址，相同内容只存储一份。",
	MsgFlagAs:      "指定资产名称",

	// checkout
	MsgCmdCheckoutShort: "从仓库还原资产到目录",
	MsgCmdCheckoutLong:  "将指定资产从仓库还原到目标目录，保留原始目录结构。",
	MsgFlagOutput:       "输出目录路径",
	MsgFlagOverwrite:    "覆盖策略 (skip|force|ask)",

	// export
	MsgCmdExportShort: "导出资产为 ZIP 压缩包",
	MsgCmdExportLong:  "将指定资产从仓库打包导出为 ZIP 文件。",

	// gc
	MsgCmdGcShort: "垃圾回收",
	MsgCmdGcLong:  "清理未被任何清单引用的孤立对象，释放磁盘空间。",
	MsgFlagDryRun: "仅列出孤立对象，不删除",

	// init
	MsgCmdInitShort: "初始化 Distill 仓库",
	MsgCmdInitLong:  "在指定路径创建 Distill 仓库目录结构、默认配置文件和空的引用索引。",
	MsgFlagTrash:    "回收站路径",

	// list
	MsgCmdListShort: "列出所有资产",
	MsgCmdListLong:  "显示仓库中所有已导入的资产列表。",
	MsgFlagFormat:   "输出格式 (table|json)",

	// remove
	MsgCmdRemoveShort: "移除资产",
	MsgCmdRemoveLong:  "从仓库中移除指定名称的资产。清单和引用将被删除，对象数据保留，等待 GC 清理。",

	// 运行时输出
	MsgAdded:          "已添加: %s (%d 文件, %d bytes)",
	MsgCheckedOut:     "已还原: %s -> %s",
	MsgExported:       "已导出: %s -> %s",
	MsgRemoved:        "已移除: %s",
	MsgGcNoOrphans:    "没有发现孤立对象。",
	MsgGcOrphanList:   "发现 %d 个孤立对象:",
	MsgGcClean:        "已清理 %d 个孤立对象。",
	MsgGcAlreadyClean: "仓库已是干净状态，无需清理。",
	MsgInited:         "Distill 仓库初始化完成: %s",
	MsgTrashPath:      "回收站路径: %s",
	MsgListEmpty:      "仓库为空，使用 distill add 添加资产",

	// 错误信息
	MsgErrCannotAccess:       "无法访问 %s: %w",
	MsgErrNotRegularFile:     "%s 不是普通文件、目录或 ZIP",
	MsgErrCannotOpen:         "无法打开 %s: %w",
	MsgErrReadFailed:         "读取 %s 失败: %w",
	MsgErrAddFailed:          "添加失败: %w",
	MsgErrReadZipFailed:      "读取 ZIP 失败: %w",
	MsgErrReadDirFailed:      "读取目录失败: %w",
	MsgErrCheckoutFailed:     "还原失败: %w",
	MsgErrExportFailed:       "导出失败: %w",
	MsgErrGcDryRunFailed:     "GC 预检失败: %w",
	MsgErrGcFailed:           "GC 执行失败: %w",
	MsgErrCreateDirFailed:    "创建目录 %s 失败: %w",
	MsgErrWriteConfigFailed:  "写入配置文件失败: %w",
	MsgErrWriteRefsFailed:    "写入引用文件失败: %w",
	MsgErrListFailed:         "列表查询失败: %w",
	MsgErrAssetNotFound:      "资产 %q 不存在",
	MsgErrReadManifestFailed: "无法读取清单: %w",
	MsgErrRemoveFailed:       "移除失败: %w",
	MsgErrTrashBackupFailed:  "回收站备份失败: %v",

	// cobra 内置命令
	MsgCmdHelpShort:       "查看命令帮助",
	MsgCmdCompletionShort: "生成自动补全脚本",

	// cobra 内置 flag
	MsgFlagHelp: "显示帮助信息",

	// cobra 帮助模板
	MsgHelpTip:     `使用 "distill [命令] --help" 获取更多关于某条命令的信息。`,
	MsgFlagDefault: "默认",

	// 交互提示
	MsgCheckoutFileExists:      "文件已存在: %s",
	MsgCheckoutOverwritePrompt: "是否覆盖？(y/N): ",
	MsgCheckoutSkipped:         "已跳过。",
	MsgWarnLogOpenFailed:       "警告: 打开日志文件失败: %v\n",
}

// en 翻译表。
var en = map[MsgKey]string{
	// 根命令
	MsgRootShort:     "Distill - Asset Management CLI",
	MsgRootLong:      "Distill is a Go CLI tool for importing, managing, and exporting assets. Uses content-addressed storage for physical deduplication.",
	MsgFlagHome:      "Repository path",
	MsgFlagLogFormat: "Log format (text|json)",
	MsgFlagLogLevel:  "Log level (debug|info|warn|error)",
	MsgFlagLang:      "UI language (zh|en)",

	// add
	MsgCmdAddShort: "Add asset to repository",
	MsgCmdAddLong:  "Add a file, folder, or ZIP archive to the Distill repository. Uses SHA-256 content addressing for deduplication.",
	MsgFlagAs:      "Specify asset name",

	// checkout
	MsgCmdCheckoutShort: "Restore asset from repository",
	MsgCmdCheckoutLong:  "Restore a specified asset from the repository to a target directory, preserving the original directory structure.",
	MsgFlagOutput:       "Output directory path",
	MsgFlagOverwrite:    "Overwrite policy (skip|force|ask)",

	// export
	MsgCmdExportShort: "Export asset as ZIP archive",
	MsgCmdExportLong:  "Package and export a specified asset from the repository as a ZIP file.",

	// gc
	MsgCmdGcShort: "Garbage collection",
	MsgCmdGcLong:  "Clean up orphaned objects not referenced by any manifest to free disk space.",
	MsgFlagDryRun: "List orphaned objects only, do not delete",

	// init
	MsgCmdInitShort: "Initialize Distill repository",
	MsgCmdInitLong:  "Create Distill repository directory structure, default config file, and empty reference index at the specified path.",
	MsgFlagTrash:    "Trash directory path",

	// list
	MsgCmdListShort: "List all assets",
	MsgCmdListLong:  "Display all imported assets in the repository.",
	MsgFlagFormat:   "Output format (table|json)",

	// remove
	MsgCmdRemoveShort: "Remove asset",
	MsgCmdRemoveLong:  "Remove a named asset from the repository. Manifests and references will be deleted; object data is retained for GC cleanup.",

	// 运行时输出
	MsgAdded:          "Added: %s (%d files, %d bytes)",
	MsgCheckedOut:     "Restored: %s -> %s",
	MsgExported:       "Exported: %s -> %s",
	MsgRemoved:        "Removed: %s",
	MsgGcNoOrphans:    "No orphaned objects found.",
	MsgGcOrphanList:   "Found %d orphaned object(s):",
	MsgGcClean:        "Cleaned up %d orphaned object(s).",
	MsgGcAlreadyClean: "Repository is already clean.",
	MsgInited:         "Distill repository initialized: %s",
	MsgTrashPath:      "Trash path: %s",
	MsgListEmpty:      "Repository is empty. Use \"distill add\" to add assets",

	// 错误信息
	MsgErrCannotAccess:       "Cannot access %s: %w",
	MsgErrNotRegularFile:     "%s is not a regular file, directory, or ZIP",
	MsgErrCannotOpen:         "Cannot open %s: %w",
	MsgErrReadFailed:         "Failed to read %s: %w",
	MsgErrAddFailed:          "Failed to add asset: %w",
	MsgErrReadZipFailed:      "Failed to read ZIP: %w",
	MsgErrReadDirFailed:      "Failed to read directory: %w",
	MsgErrCheckoutFailed:     "Failed to checkout: %w",
	MsgErrExportFailed:       "Failed to export: %w",
	MsgErrGcDryRunFailed:     "GC dry-run failed: %w",
	MsgErrGcFailed:           "GC failed: %w",
	MsgErrCreateDirFailed:    "Failed to create directory %s: %w",
	MsgErrWriteConfigFailed:  "Failed to write config file: %w",
	MsgErrWriteRefsFailed:    "Failed to write refs file: %w",
	MsgErrListFailed:         "Failed to list assets: %w",
	MsgErrAssetNotFound:      "Asset %q not found",
	MsgErrReadManifestFailed: "Failed to read manifest: %w",
	MsgErrRemoveFailed:       "Failed to remove asset: %w",
	MsgErrTrashBackupFailed:  "Warning: trash backup failed: %v",

	// cobra 内置命令
	MsgCmdHelpShort:       "Help about any command",
	MsgCmdCompletionShort: "Generate the autocompletion script for the specified shell",

	// cobra 内置 flag
	MsgFlagHelp: "Show help information",

	// cobra 帮助模板
	MsgHelpTip:     `Use "distill [command] --help" for more information about a command.`,
	MsgFlagDefault: "default",

	// 交互提示
	MsgCheckoutFileExists:      "File already exists: %s",
	MsgCheckoutOverwritePrompt: "Overwrite? (y/N): ",
	MsgCheckoutSkipped:         "Skipped.",
	MsgWarnLogOpenFailed:       "Warning: failed to open log file: %v\n",
}

// SetLang 设置全局语言。无效值默认回退到中文。
func SetLang(lang string) {
	langMu.Lock()
	defer langMu.Unlock()

	switch strings.ToLower(lang) {
	case "en":
		currentLang = LangEn
	default:
		currentLang = LangZh
	}
}

// CurrentLang 返回当前语言设置。
func CurrentLang() Lang {
	langMu.RLock()
	defer langMu.RUnlock()
	return currentLang
}

// ResetLang 重置语言为默认中文。
func ResetLang() {
	langMu.Lock()
	defer langMu.Unlock()
	currentLang = LangZh
}

// T 根据当前语言返回对应文案，支持 fmt.Sprintf 格式化参数。
func T(key MsgKey, args ...any) string {
	langMu.RLock()
	l := currentLang
	langMu.RUnlock()

	var text string
	switch l {
	case LangEn:
		text = en[key]
	default:
		text = zh[key]
	}

	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}
	return text
}

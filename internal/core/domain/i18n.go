// 提供 Distill CLI 的国际化支持。
// 所有翻译文案通过 go-i18n 从 TOML 翻译文件加载。
// 支持中文（zh，默认）和英文（en）两种语言。
package domain

import (
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
	MsgFlagVersion

	// config 命令
	MsgCmdConfigShort
	MsgCmdConfigLong
	MsgCmdConfigShowShort
	MsgCmdConfigShowLong
	MsgCmdConfigGetShort
	MsgCmdConfigGetLong
	MsgCmdConfigSetShort
	MsgCmdConfigSetLong

	// cobra 帮助模板
	MsgHelpTip
	MsgFlagDefault

	// 交互提示
	MsgCheckoutFileExists
	MsgCheckoutOverwritePrompt
	MsgCheckoutSkipped
	MsgExportFileExists
	MsgExportOverwritePrompt
	MsgExportSkipped
	MsgWarnLogOpenFailed
)

// P 是 T() 函数的具名参数类型，用于 go-i18n 模板插值。
// 键名与 TOML 翻译文件中的 {{.Key}} 占位符对应。
type P map[string]any

var (
	currentLang Lang
	langMu      sync.RWMutex
)

func init() {
	currentLang = LangZh // 默认中文
}

// msgKeyID 将 MsgKey 映射到 go-i18n 的字符串 ID。
// 所有 MsgKey 都有映射，翻译文案统一来自 TOML 文件。
var msgKeyID = map[MsgKey]string{
	// 根命令
	MsgRootShort:     "MsgRootShort",
	MsgRootLong:      "MsgRootLong",
	MsgFlagHome:      "MsgFlagHome",
	MsgFlagLogFormat: "MsgFlagLogFormat",
	MsgFlagLogLevel:  "MsgFlagLogLevel",
	MsgFlagLang:      "MsgFlagLang",

	// add 命令
	MsgCmdAddShort: "MsgCmdAddShort",
	MsgCmdAddLong:  "MsgCmdAddLong",
	MsgFlagAs:      "MsgFlagAs",
	MsgAdded:       "MsgAdded",

	// checkout 命令
	MsgCmdCheckoutShort:        "MsgCmdCheckoutShort",
	MsgCmdCheckoutLong:         "MsgCmdCheckoutLong",
	MsgFlagOutput:              "MsgFlagOutput",
	MsgFlagOverwrite:           "MsgFlagOverwrite",
	MsgCheckedOut:              "MsgCheckedOut",
	MsgCheckoutFileExists:      "MsgCheckoutFileExists",
	MsgCheckoutOverwritePrompt: "MsgCheckoutOverwritePrompt",
	MsgCheckoutSkipped:         "MsgCheckoutSkipped",

	// export 命令
	MsgCmdExportShort:        "MsgCmdExportShort",
	MsgCmdExportLong:         "MsgCmdExportLong",
	MsgExported:              "MsgExported",
	MsgExportFileExists:      "MsgExportFileExists",
	MsgExportOverwritePrompt: "MsgExportOverwritePrompt",
	MsgExportSkipped:         "MsgExportSkipped",

	// gc 命令
	MsgCmdGcShort:     "MsgCmdGcShort",
	MsgCmdGcLong:      "MsgCmdGcLong",
	MsgFlagDryRun:     "MsgFlagDryRun",
	MsgGcNoOrphans:    "MsgGcNoOrphans",
	MsgGcOrphanList:   "MsgGcOrphanList",
	MsgGcClean:        "MsgGcClean",
	MsgGcAlreadyClean: "MsgGcAlreadyClean",

	// init 命令
	MsgCmdInitShort: "MsgCmdInitShort",
	MsgCmdInitLong:  "MsgCmdInitLong",
	MsgFlagTrash:    "MsgFlagTrash",
	MsgInited:       "MsgInited",
	MsgTrashPath:    "MsgTrashPath",

	// list 命令
	MsgCmdListShort: "MsgCmdListShort",
	MsgCmdListLong:  "MsgCmdListLong",
	MsgFlagFormat:   "MsgFlagFormat",
	MsgListEmpty:    "MsgListEmpty",

	// remove 命令
	MsgCmdRemoveShort: "MsgCmdRemoveShort",
	MsgCmdRemoveLong:  "MsgCmdRemoveLong",
	MsgRemoved:        "MsgRemoved",

	// 错误信息
	MsgErrCannotAccess:       "MsgErrCannotAccess",
	MsgErrNotRegularFile:     "MsgErrNotRegularFile",
	MsgErrCannotOpen:         "MsgErrCannotOpen",
	MsgErrReadFailed:         "MsgErrReadFailed",
	MsgErrAddFailed:          "MsgErrAddFailed",
	MsgErrReadZipFailed:      "MsgErrReadZipFailed",
	MsgErrReadDirFailed:      "MsgErrReadDirFailed",
	MsgErrCheckoutFailed:     "MsgErrCheckoutFailed",
	MsgErrExportFailed:       "MsgErrExportFailed",
	MsgErrGcDryRunFailed:     "MsgErrGcDryRunFailed",
	MsgErrGcFailed:           "MsgErrGcFailed",
	MsgErrCreateDirFailed:    "MsgErrCreateDirFailed",
	MsgErrWriteConfigFailed:  "MsgErrWriteConfigFailed",
	MsgErrWriteRefsFailed:    "MsgErrWriteRefsFailed",
	MsgErrListFailed:         "MsgErrListFailed",
	MsgErrAssetNotFound:      "MsgErrAssetNotFound",
	MsgErrReadManifestFailed: "MsgErrReadManifestFailed",
	MsgErrRemoveFailed:       "MsgErrRemoveFailed",
	MsgErrTrashBackupFailed:  "MsgErrTrashBackupFailed",

	// cobra 内置命令
	MsgCmdHelpShort:       "MsgCmdHelpShort",
	MsgCmdCompletionShort: "MsgCmdCompletionShort",

	// cobra 内置 flag
	MsgFlagHelp:    "MsgFlagHelp",
	MsgFlagVersion: "MsgFlagVersion",

	// config 命令
	MsgCmdConfigShort:     "MsgCmdConfigShort",
	MsgCmdConfigLong:      "MsgCmdConfigLong",
	MsgCmdConfigShowShort: "MsgCmdConfigShowShort",
	MsgCmdConfigShowLong:  "MsgCmdConfigShowLong",
	MsgCmdConfigGetShort:  "MsgCmdConfigGetShort",
	MsgCmdConfigGetLong:   "MsgCmdConfigGetLong",
	MsgCmdConfigSetShort:  "MsgCmdConfigSetShort",
	MsgCmdConfigSetLong:   "MsgCmdConfigSetLong",

	// cobra 帮助模板
	MsgHelpTip:     "MsgHelpTip",
	MsgFlagDefault: "MsgFlagDefault",

	// 交互提示
	MsgWarnLogOpenFailed: "MsgWarnLogOpenFailed",
}

// langTag 将 Lang 转换为 go-i18n 的语言标签字符串。
func langTag(l Lang) string {
	switch l {
	case LangEn:
		return "en"
	default:
		return "zh"
	}
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

// T 根据当前语言返回对应文案。
// 使用 go-i18n Bundle 查找翻译，支持 text/template 插值语法（{{.Name}}）。
// params 为具名参数，与 TOML 翻译文件中的 {{.Key}} 占位符对应。
// 无参数时可直接调用 T(key)，无需传 nil。
func T(key MsgKey, params ...P) string {
	langMu.RLock()
	l := currentLang
	langMu.RUnlock()

	id, ok := msgKeyID[key]
	if !ok {
		return ""
	}

	var p P
	if len(params) > 0 {
		p = params[0]
	}

	if b := getBundle(); b != nil {
		text, err := b.Localize(langTag(l), id, p)
		if err == nil {
			return text
		}
	}
	return ""
}

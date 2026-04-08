package domain

import (
	"os"
	"path/filepath"
	"testing"
)

// projectLocalesDir 返回项目 locales 目录的路径。
// 测试文件位于 internal/core/domain/，locales/ 位于项目根目录。
var projectLocalesDir = filepath.Join("..", "..", "..", "locales")

func TestMain(m *testing.M) {
	// 确保所有测试开始时 Bundle 处于干净状态
	resetBundle()
	code := m.Run()
	os.Exit(code)
}

// initBundleFromProjectLocales 从项目的 locales 目录初始化 Bundle。
// 用于需要完整翻译的测试。
func initBundleFromProjectLocales(t *testing.T) {
	t.Helper()
	if err := InitBundle(projectLocalesDir); err != nil {
		t.Fatalf("InitBundle(%q) failed: %v", projectLocalesDir, err)
	}
}

func TestSetLang_ValidLang(t *testing.T) {
	tests := []struct {
		lang string
		want Lang
	}{
		{"zh", LangZh},
		{"en", LangEn},
		{"ZH", LangZh},
		{"EN", LangEn},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			SetLang(tt.lang)
			if CurrentLang() != tt.want {
				t.Errorf("SetLang(%q) = %v, want %v", tt.lang, CurrentLang(), tt.want)
			}
		})
	}
}

func TestSetLang_InvalidLang_DefaultsToZh(t *testing.T) {
	SetLang("fr")
	if CurrentLang() != LangZh {
		t.Errorf("SetLang(%q) should default to LangZh, got %v", "fr", CurrentLang())
	}
}

func TestT(t *testing.T) {
	resetBundle()
	defer resetBundle()
	initBundleFromProjectLocales(t)

	SetLang("zh")
	if got := T(MsgRootShort); got != "Distill - 资产管理 CLI 工具" {
		t.Errorf("T(MsgRootShort) zh = %q, want %q", got, "Distill - 资产管理 CLI 工具")
	}

	SetLang("en")
	if got := T(MsgRootShort); got != "Distill - Asset Management CLI" {
		t.Errorf("T(MsgRootShort) en = %q, want %q", got, "Distill - Asset Management CLI")
	}
}

func TestT_AllKeysHaveTranslations(t *testing.T) {
	resetBundle()
	defer resetBundle()
	initBundleFromProjectLocales(t)

	// 确保中英文对每个 key 都有翻译
	SetLang("zh")
	for _, key := range allKeys() {
		if got := T(key); got == "" {
			t.Errorf("missing zh translation for key %d", key)
		}
	}
	SetLang("en")
	for _, key := range allKeys() {
		if got := T(key); got == "" {
			t.Errorf("missing en translation for key %d", key)
		}
	}
}

func TestT_WithNamedParams(t *testing.T) {
	resetBundle()
	defer resetBundle()
	initBundleFromProjectLocales(t)

	SetLang("zh")
	got := T(MsgAdded, P{"Name": "test.zip", "FileCount": 3, "TotalSize": 1024})
	want := "已添加: test.zip (3 文件, 1024 bytes)"
	if got != want {
		t.Errorf("T(MsgAdded, ...) zh = %q, want %q", got, want)
	}

	SetLang("en")
	got = T(MsgAdded, P{"Name": "test.zip", "FileCount": 3, "TotalSize": 1024})
	want = "Added: test.zip (3 files, 1024 bytes)"
	if got != want {
		t.Errorf("T(MsgAdded, ...) en = %q, want %q", got, want)
	}
}

func TestResetLang(t *testing.T) {
	SetLang("en")
	ResetLang()
	if CurrentLang() != LangZh {
		t.Error("ResetLang() should reset to LangZh")
	}
}

// allKeys 返回所有已定义的文案 key，用于遍历检查翻译完整性。
func allKeys() []MsgKey {
	return []MsgKey{
		MsgRootShort,
		MsgRootLong,
		MsgFlagHome,
		MsgFlagLogFormat,
		MsgFlagLogLevel,
		MsgFlagLang,
		MsgCmdAddShort,
		MsgCmdAddLong,
		MsgFlagAs,
		MsgCmdCheckoutShort,
		MsgCmdCheckoutLong,
		MsgFlagOutput,
		MsgFlagOverwrite,
		MsgCmdExportShort,
		MsgCmdExportLong,
		MsgCmdGcShort,
		MsgCmdGcLong,
		MsgFlagDryRun,
		MsgCmdInitShort,
		MsgCmdInitLong,
		MsgFlagTrash,
		MsgCmdListShort,
		MsgCmdListLong,
		MsgFlagFormat,
		MsgCmdRemoveShort,
		MsgCmdRemoveLong,
		MsgAdded,
		MsgCheckedOut,
		MsgExported,
		MsgRemoved,
		MsgGcNoOrphans,
		MsgGcOrphanList,
		MsgGcClean,
		MsgGcAlreadyClean,
		MsgInited,
		MsgTrashPath,
		MsgListEmpty,
		MsgErrCannotAccess,
		MsgErrNotRegularFile,
		MsgErrCannotOpen,
		MsgErrReadFailed,
		MsgErrAddFailed,
		MsgErrReadZipFailed,
		MsgErrReadDirFailed,
		MsgErrCheckoutFailed,
		MsgErrExportFailed,
		MsgErrGcDryRunFailed,
		MsgErrGcFailed,
		MsgErrCreateDirFailed,
		MsgErrWriteConfigFailed,
		MsgErrWriteRefsFailed,
		MsgErrListFailed,
		MsgErrAssetNotFound,
		MsgErrReadManifestFailed,
		MsgErrRemoveFailed,
		MsgErrTrashBackupFailed,
		MsgCheckoutFileExists,
		MsgCheckoutOverwritePrompt,
		MsgCheckoutSkipped,
		MsgExportFileExists,
		MsgExportOverwritePrompt,
		MsgExportSkipped,
		MsgWarnLogOpenFailed,
		MsgCmdHelpShort,
		MsgCmdCompletionShort,
		MsgFlagHelp,
		MsgFlagVersion,
		MsgCmdConfigShort,
		MsgCmdConfigLong,
		MsgCmdConfigShowShort,
		MsgCmdConfigShowLong,
		MsgCmdConfigGetShort,
		MsgCmdConfigGetLong,
		MsgCmdConfigSetShort,
		MsgCmdConfigSetLong,
		MsgHelpTip,
		MsgFlagDefault,
	}
}

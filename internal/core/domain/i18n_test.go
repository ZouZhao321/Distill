package domain

import (
	"testing"
)

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

func TestT_WithFormatArgs(t *testing.T) {
	SetLang("zh")
	got := T(MsgAdded, "test.zip", 3, 1024)
	want := "已添加: test.zip (3 文件, 1024 bytes)"
	if got != want {
		t.Errorf("T(MsgAdded, ...) zh = %q, want %q", got, want)
	}

	SetLang("en")
	got = T(MsgAdded, "test.zip", 3, 1024)
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
		MsgWarnLogOpenFailed,
	}
}

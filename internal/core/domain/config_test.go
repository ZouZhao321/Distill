package domain

import (
	"strings"
	"testing"
)

func TestFormatConfigItem_EnumKey(t *testing.T) {
	// 有可选值的枚举类型配置项
	got := FormatConfigItem("checkout.overwrite", "ask")
	// 应包含 key=value, 可选值列表, 说明
	if !strings.Contains(got, "checkout.overwrite=ask") {
		t.Errorf("missing key=value, got:\n%s", got)
	}
	if !strings.Contains(got, "[ask, skip, force]") {
		t.Errorf("missing valid values, got:\n%s", got)
	}
	if !strings.Contains(got, "导出时的覆盖策略") {
		t.Errorf("missing description, got:\n%s", got)
	}
}

func TestFormatConfigItem_PathKey(t *testing.T) {
	// 路径类配置项：无可选值，只有说明
	got := FormatConfigItem("store.home", "C:/Users/test")
	if !strings.Contains(got, "store.home=C:/Users/test") {
		t.Errorf("missing key=value, got:\n%s", got)
	}
	if strings.Contains(got, "[") {
		t.Errorf("path key should not have valid values, got:\n%s", got)
	}
	if !strings.Contains(got, "仓库存储根目录") {
		t.Errorf("missing description, got:\n%s", got)
	}
}

func TestFormatConfigItem_BoolKey(t *testing.T) {
	got := FormatConfigItem("normalize.crlf_to_lf", "true")
	if !strings.Contains(got, "normalize.crlf_to_lf=true") {
		t.Errorf("missing key=value, got:\n%s", got)
	}
	if !strings.Contains(got, "[true, false]") {
		t.Errorf("missing valid values, got:\n%s", got)
	}
	if !strings.Contains(got, "导出时将 CRLF 转换为 LF") {
		t.Errorf("missing description, got:\n%s", got)
	}
}

func TestFormatConfigItem_UnknownKey(t *testing.T) {
	// 未定义的 key 应仍输出 key=value
	got := FormatConfigItem("unknown.key", "val")
	if !strings.Contains(got, "unknown.key=val") {
		t.Errorf("unknown key should still show key=value, got:\n%s", got)
	}
}

func TestFormatConfigItem_MultiLevelKey(t *testing.T) {
	got := FormatConfigItem("log.level", "warn")
	if !strings.Contains(got, "log.level=warn") {
		t.Errorf("missing key=value, got:\n%s", got)
	}
	if !strings.Contains(got, "[debug, info, warn, error]") {
		t.Errorf("missing valid values, got:\n%s", got)
	}
	if !strings.Contains(got, "日志级别") {
		t.Errorf("missing description, got:\n%s", got)
	}
}

func TestFormatConfigItem_LangKey(t *testing.T) {
	got := FormatConfigItem("lang", "zh")
	if !strings.Contains(got, "lang=zh") {
		t.Errorf("missing key=value, got:\n%s", got)
	}
	if !strings.Contains(got, "[zh, en]") {
		t.Errorf("missing valid values, got:\n%s", got)
	}
	if !strings.Contains(got, "界面语言") {
		t.Errorf("missing description, got:\n%s", got)
	}
}

func TestFormatAllConfig(t *testing.T) {
	config := DefaultConfig()
	got := FormatAllConfig(config)

	// 应包含所有配置项
	expectedKeys := []string{
		"core.version", "core.objects_format",
		"store.home", "store.trash_path",
		"checkout.overwrite", "log.format", "log.level",
		"normalize.crlf_to_lf", "lang",
	}
	for _, key := range expectedKeys {
		if !strings.Contains(got, key+"=") {
			t.Errorf("missing key %q in output:\n%s", key, got)
		}
	}

	// 枚举类型应有可选值
	if !strings.Contains(got, "[ask, skip, force]") {
		t.Errorf("missing checkout.overwrite valid values in output:\n%s", got)
	}

	// 路径类型不应有可选值括号
	lines := strings.Split(got, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "store.home=") && strings.Contains(line, "[") {
			t.Errorf("store.home should not have valid values, got:\n%s", line)
		}
	}

	// 应包含说明
	if !strings.Contains(got, "导出时的覆盖策略") {
		t.Errorf("missing description in output:\n%s", got)
	}
}

func TestFormatAllConfig_WithCustomValues(t *testing.T) {
	config := &Config{}
	config.Lang = "en"
	config.Log.Level = "debug"
	config.Checkout.Overwrite = "force"

	got := FormatAllConfig(config)

	if !strings.Contains(got, "lang=en") {
		t.Errorf("expected lang=en, got:\n%s", got)
	}
	if !strings.Contains(got, "log.level=debug") {
		t.Errorf("expected log.level=debug, got:\n%s", got)
	}
	if !strings.Contains(got, "checkout.overwrite=force") {
		t.Errorf("expected checkout.overwrite=force, got:\n%s", got)
	}
}

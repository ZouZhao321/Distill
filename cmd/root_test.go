package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestRootCmd_NoDuplicateError 验证子命令 RunE 返回错误时，
// Cobra 不会重复打印错误信息或 Usage。
//
// 背景：rootCmd 默认 SilenceErrors=false 和 SilenceUsage=false，
// Cobra 会自行打印 "Error: ..." + Usage，然后 Execute() 又用
// fmt.Fprintln(os.Stderr, err) 打印了一遍，导致错误信息重复两遍。
//
// 修复：给 rootCmd 设置 SilenceErrors=true 和 SilenceUsage=true，
// 让 Cobra 保持静默，由 Execute() 统一打印一次错误。
func TestRootCmd_NoDuplicateError(t *testing.T) {
	t.Setenv("DISTILL_HOME", t.TempDir())

	// 添加一个临时子命令，RunE 返回固定错误
	testCmd := &cobra.Command{
		Use: "__test_dup_error",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("unique marker error abc123")
		},
	}
	rootCmd.AddCommand(testCmd)
	defer rootCmd.RemoveCommand(testCmd)

	// 捕获 stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	rootCmd.SetArgs([]string{"__test_dup_error"})
	err := rootCmd.Execute()

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err == nil {
		t.Fatal("expected error from test command, got nil")
	}

	// Cobra 不应打印错误（SilenceErrors=true）
	if strings.Contains(output, "unique marker error abc123") {
		t.Errorf("Cobra should not print error to stderr (SilenceErrors=true), but output contains:\n%s", output)
	}

	// Cobra 不应为运行时错误打印 Usage（SilenceUsage=true）
	if strings.Contains(output, "Usage:") {
		t.Errorf("Cobra should not print Usage for runtime errors (SilenceUsage=true), but output contains:\n%s", output)
	}
}

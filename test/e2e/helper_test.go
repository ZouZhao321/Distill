//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// distillBin 是编译后的 distill 二进制路径。
// 由 TestMain 在测试开始前编译并设置。
var distillBin string

// TestMain 编译 distill 二进制并运行所有 e2e 测试。
func TestMain(m *testing.M) {
	// 确定项目根目录：从当前文件路径向上两级
	_, thisFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	projectRoot, _ = filepath.Abs(projectRoot)

	// 编译 distill 二进制到临时目录
	tmpDir, err := os.MkdirTemp("", "distill-e2e-build-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp build dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binName := "distill"
	if runtime.GOOS == "windows" {
		binName = "distill.exe"
	}
	distillBin = filepath.Join(tmpDir, binName)

	cmd := exec.Command("go", "build", "-o", distillBin, ".")
	cmd.Dir = projectRoot
	var buildOut strings.Builder
	cmd.Stdout = &buildOut
	cmd.Stderr = &buildOut
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build distill: %v\nbuild output:\n%s", err, buildOut.String())
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// Env 创建一个隔离的测试环境，包含临时 DISTILL_HOME 目录。
// 调用方不需要手动清理，t.TempDir() 会在测试结束后自动清理。
func Env(t *testing.T) *TestEnv {
	t.Helper()
	home := t.TempDir()
	return &TestEnv{
		Home:    home,
		T:       t,
		WorkDir: home,
	}
}

// TestEnv 封装一次 e2e 测试的隔离环境。
type TestEnv struct {
	Home    string // DISTILL_HOME 目录
	WorkDir string // 工作目录（用于创建测试文件）
	T       *testing.T
}

// Run 执行 distill 命令并返回结果。
// args 是传给 distill 的参数列表（不含 "distill" 本身）。
func (e *TestEnv) Run(args ...string) *Result {
	e.T.Helper()

	cmd := exec.Command(distillBin, args...)
	cmd.Dir = e.WorkDir
	// 完全隔离：不继承宿主环境变量，避免宿主 DISTILL_HOME/DISTILL_LANG
	// 污染测试（尤其在 Linux/macOS 上 os.Getenv 返回首个匹配，append 到末尾无效）。
	cmd.Env = []string{
		"DISTILL_HOME=" + e.Home,
		"DISTILL_LANG=zh",
		"PATH=" + os.Getenv("PATH"),
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			e.T.Fatalf("command failed to start: %v", err)
		}
	}

	return &Result{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}
}

// WriteFile 在 WorkDir 下创建文件并写入内容。
func (e *TestEnv) WriteFile(name, content string) string {
	e.T.Helper()
	path := filepath.Join(e.WorkDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		e.T.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		e.T.Fatalf("failed to write file: %v", err)
	}
	return path
}

// WriteBinaryFile 在 WorkDir 下创建二进制文件。
func (e *TestEnv) WriteBinaryFile(name string, data []byte) string {
	e.T.Helper()
	path := filepath.Join(e.WorkDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		e.T.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		e.T.Fatalf("failed to write file: %v", err)
	}
	return path
}

// Mkdir 在 WorkDir 下创建目录。
func (e *TestEnv) Mkdir(name string) string {
	e.T.Helper()
	path := filepath.Join(e.WorkDir, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		e.T.Fatalf("failed to create dir: %v", err)
	}
	return path
}

// ReadFile 读取指定路径的文件内容。
func (e *TestEnv) ReadFile(path string) string {
	e.T.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		e.T.Fatalf("failed to read file %s: %v", path, err)
	}
	return string(data)
}

// ReadBinaryFile 读取指定路径的文件二进制内容。
func (e *TestEnv) ReadBinaryFile(path string) []byte {
	e.T.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		e.T.Fatalf("failed to read file %s: %v", path, err)
	}
	return data
}

// FileExists 检查文件是否存在。
func (e *TestEnv) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists 检查目录是否存在。
func (e *TestEnv) DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// CountFiles 计算目录下的文件数量（递归）。
func (e *TestEnv) CountFiles(dir string) int {
	e.T.Helper()
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		e.T.Fatalf("CountFiles(%s) failed: %v", dir, err)
	}
	return count
}

// Result 封装命令执行结果。
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// Success 断言命令成功执行（退出码为 0）。
func (r *Result) Success(t *testing.T) *Result {
	t.Helper()
	if r.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s",
			r.ExitCode, r.Stdout, r.Stderr)
	}
	return r
}

// Fail 断言命令执行失败（退出码非 0）。
func (r *Result) Fail(t *testing.T) *Result {
	t.Helper()
	if r.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0\nstdout: %s", r.Stdout)
	}
	return r
}

// Contains 断言 stdout 包含指定子串。
func (r *Result) Contains(t *testing.T, substr string) *Result {
	t.Helper()
	if !strings.Contains(r.Stdout, substr) {
		t.Fatalf("stdout does not contain %q\nstdout: %s", substr, r.Stdout)
	}
	return r
}

// ContainsErr 断言 stderr 包含指定子串。
func (r *Result) ContainsErr(t *testing.T, substr string) *Result {
	t.Helper()
	if !strings.Contains(r.Stderr, substr) {
		t.Fatalf("stderr does not contain %q\nstderr: %s", substr, r.Stderr)
	}
	return r
}

// StdoutContains 断言 stdout 包含所有指定子串。
func (r *Result) StdoutContains(t *testing.T, substrs ...string) *Result {
	t.Helper()
	for _, s := range substrs {
		if !strings.Contains(r.Stdout, s) {
			t.Fatalf("stdout does not contain %q\nstdout: %s", s, r.Stdout)
		}
	}
	return r
}

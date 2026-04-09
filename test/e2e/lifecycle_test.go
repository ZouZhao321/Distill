//go:build e2e

package e2e

import (
	"path/filepath"
	"testing"
)

// TestLifecycle_FullWorkflow 测试完整的资源生命周期:
// init → add 单文件 → list → checkout → remove → gc
func TestLifecycle_FullWorkflow(t *testing.T) {
	env := Env(t)

	// Step 1: init — 初始化仓库
	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).
		Success(t).
		Contains(t, "仓库初始化完成")

	// 验证目录结构已创建
	if !env.DirExists(filepath.Join(env.Home, "objects")) {
		t.Fatal("objects directory should exist after init")
	}
	if !env.DirExists(filepath.Join(env.Home, "manifests")) {
		t.Fatal("manifests directory should exist after init")
	}

	// Step 2: add — 添加一个文件
	filePath := env.WriteFile("hello.txt", "hello world\n")
	env.Run("add", filePath).
		Success(t).
		Contains(t, "已添加")

	// Step 3: list — 查看资源列表
	env.Run("list").
		Success(t).
		StdoutContains(t, "hello.txt")

	// Step 4: checkout — 签出文件
	outDir := filepath.Join(env.Home, "checkout-output")
	env.Mkdir("checkout-output") // checkout 需要父目录存在
	env.Run("checkout", "hello.txt", "--output", outDir).
		Success(t).
		Contains(t, "已还原")

	// 验证签出内容与原始文件一致
	actual := env.ReadFile(filepath.Join(outDir, "hello.txt"))
	if actual != "hello world\n" {
		t.Fatalf("checkout content mismatch: got %q, want %q", actual, "hello world\n")
	}

	// Step 5: remove — 删除资源
	env.Run("remove", "hello.txt").
		Success(t).
		Contains(t, "已移除")

	// list 应该为空
	env.Run("list").
		Success(t).
		Contains(t, "仓库为空")

	// Step 6: gc — 清理孤儿对象
	env.Run("gc").
		Success(t)

	// 验证 objects 目录已被清空
	objCount := env.CountFiles(filepath.Join(env.Home, "objects"))
	if objCount != 0 {
		t.Fatalf("expected 0 objects after gc, got %d", objCount)
	}
}

// TestInit_Idempotent 测试重复 init 应报错。
func TestInit_Idempotent(t *testing.T) {
	env := Env(t)

	// 第一次 init 应成功
	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).
		Success(t)

	// 第二次 init 应失败
	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).
		Fail(t)
}

// TestAdd_DuplicateName 测试重复 add 同名文件应报错。
func TestAdd_DuplicateName(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	// 第一次 add
	file1 := env.WriteFile("doc.txt", "version 1\n")
	env.Run("add", file1).Success(t)

	// 重复 add 同名文件应失败（退出码非零即可，错误消息措辞由单元测试验证）
	env.WriteFile("doc.txt", "version 2\n")
	env.Run("add", filepath.Join(env.WorkDir, "doc.txt")).Fail(t)
}

// TestAdd_AsFlag 测试 --as 标志自定义资源名称。
func TestAdd_AsFlag(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	filePath := env.WriteFile("original-name.txt", "content\n")
	env.Run("add", filePath, "--as", "custom-name").Success(t)

	env.Run("list").Success(t).Contains(t, "custom-name")

	// checkout 时使用自定义名称
	outDir := filepath.Join(env.Home, "out")
	env.Mkdir("out")
	env.Run("checkout", "custom-name", "--output", outDir).Success(t)
}

//go:build e2e

package e2e

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDedup_SameContentDifferentNames 测试内容去重:
// 两个内容相同但文件名不同的文件，objects 目录下应只有 1 个 blob。
func TestDedup_SameContentDifferentNames(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	// 创建两个内容相同的文件，不同文件名
	fileA := env.WriteFile("fileA.txt", "same content here\n")
	fileB := env.WriteFile("fileB.txt", "same content here\n")

	env.Run("add", fileA).Success(t)
	env.Run("add", fileB).Success(t)

	// 验证 list 中有两个资产
	listResult := env.Run("list").Success(t)
	listResult.StdoutContains(t, "fileA.txt", "fileB.txt")

	// 验证 objects 目录只有 1 个 blob（内容相同只存一份）
	objCount := env.CountFiles(filepath.Join(env.Home, "objects"))
	if objCount != 1 {
		t.Fatalf("expected 1 object (deduplicated), got %d", objCount)
	}
}

// TestDedup_DifferentContent 测试不同内容应产生不同 blob:
// 两个内容不同的文件，objects 目录下应有 2 个 blob。
func TestDedup_DifferentContent(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	fileA := env.WriteFile("alpha.txt", "alpha content\n")
	fileB := env.WriteFile("beta.txt", "beta content\n")

	env.Run("add", fileA).Success(t)
	env.Run("add", fileB).Success(t)

	objCount := env.CountFiles(filepath.Join(env.Home, "objects"))
	if objCount != 2 {
		t.Fatalf("expected 2 objects (different content), got %d", objCount)
	}
}

// TestGC_AlreadyClean 测试空仓库 gc 应报告已干净。
func TestGC_AlreadyClean(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	env.Run("gc").Success(t).Contains(t, "干净状态")
}

// TestCRLF_Normalization 测试 CRLF 文件的端到端行为:
// 添加一个 CRLF 文件后 checkout，内容应保持一致。
func TestCRLF_Normalization(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	// 写入 CRLF 内容
	crlfPath := env.WriteFile("crlf-file.txt", "line1\r\nline2\r\n")
	env.Run("add", crlfPath).Success(t)

	// checkout 并验证
	outDir := filepath.Join(env.Home, "out")
	env.Mkdir("out")
	env.Run("checkout", "crlf-file.txt", "--output", outDir).Success(t)

	actual := env.ReadFile(filepath.Join(outDir, "crlf-file.txt"))
	// Distill 应该规范化换行符，checkout 后应为 LF
	if strings.Contains(actual, "\r\n") {
		t.Fatalf("expected LF line endings after checkout, got CRLF: %q", actual)
	}
	if actual != "line1\nline2\n" {
		t.Fatalf("content mismatch: got %q, want %q", actual, "line1\nline2\n")
	}
}

// TestCRLF_Dedup 测试 CRLF 和 LF 版本的同一内容应去重为同一 blob。
func TestCRLF_Dedup(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	// 写入 CRLF 版本
	crlfPath := env.WriteBinaryFile("crlf.txt", []byte("hello\r\nworld\r\n"))
	env.Run("add", crlfPath).Success(t)

	// 写入 LF 版本（相同逻辑内容）
	lfPath := env.WriteBinaryFile("lf.txt", []byte("hello\nworld\n"))
	env.Run("add", lfPath).Success(t)

	// 验证 objects 目录只有 1 个 blob（规范化后内容相同）
	objCount := env.CountFiles(filepath.Join(env.Home, "objects"))
	if objCount != 1 {
		t.Fatalf("expected 1 object (CRLF and LF deduped), got %d", objCount)
	}
}

// TestDirectory_ImportExport 测试目录导入和导出:
// add 一个目录 → list → export 为 ZIP → 解压验证内容。
func TestDirectory_ImportExport(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	// 创建目录结构
	dir := env.Mkdir("my-project")
	env.WriteFile("my-project/README.md", "# My Project\n")
	env.WriteFile("my-project/src/main.go", "package main\n")

	env.Run("add", dir).Success(t)

	// list 应显示目录资产
	env.Run("list").Success(t).Contains(t, "my-project")

	// export 为 ZIP
	zipPath := filepath.Join(env.Home, "output.zip")
	env.Run("export", "my-project", "--output", zipPath, "--overwrite", "force").
		Success(t).
		Contains(t, "已导出")

	// 验证 ZIP 文件存在且包含正确内容
	entries := readZipEntries(t, zipPath)
	hasReadme := false
	hasMain := false
	for _, name := range entries {
		if strings.HasSuffix(name, "README.md") {
			hasReadme = true
		}
		if strings.HasSuffix(name, "main.go") {
			hasMain = true
		}
	}
	if !hasReadme {
		t.Fatal("exported ZIP should contain README.md")
	}
	if !hasMain {
		t.Fatal("exported ZIP should contain main.go")
	}
}

// TestDirectory_ExportRoundTrip 测试目录 add → export → checkout 的一致性。
func TestDirectory_ExportRoundTrip(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	// 创建目录
	dir := env.Mkdir("webapp")
	env.WriteFile("webapp/index.html", "<html></html>\n")
	env.WriteFile("webapp/style.css", "body { margin: 0; }\n")

	env.Run("add", dir).Success(t)

	// checkout
	outDir := filepath.Join(env.Home, "webapp-out")
	env.Mkdir("webapp-out")
	env.Run("checkout", "webapp", "--output", outDir).Success(t)

	// 验证 checkout 内容
	actualHTML := env.ReadFile(filepath.Join(outDir, "webapp", "index.html"))
	if actualHTML != "<html></html>\n" {
		t.Fatalf("index.html content mismatch: got %q", actualHTML)
	}
	actualCSS := env.ReadFile(filepath.Join(outDir, "webapp", "style.css"))
	if actualCSS != "body { margin: 0; }\n" {
		t.Fatalf("style.css content mismatch: got %q", actualCSS)
	}
}

// TestList_JsonFormat 测试 list --format json 输出。
func TestList_JsonFormat(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	file := env.WriteFile("test.json", "{}\n")
	env.Run("add", file).Success(t)

	result := env.Run("list", "--format", "json").Success(t)
	// JSON 输出应包含必要的字段
	result.StdoutContains(t, `"name":`, `"hash":`, `"file_count":`, `"total_size":`)
}

// readZipEntries 读取 ZIP 文件中的所有条目名称。
func readZipEntries(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		t.Fatalf("failed to stat zip: %v", err)
	}

	r, err := zip.NewReader(f, stat.Size())
	if err != nil {
		t.Fatalf("failed to read zip: %v", err)
	}

	var names []string
	for _, f := range r.File {
		// 只列出文件，不列出目录
		if !f.FileInfo().IsDir() {
			names = append(names, f.Name)
		}
	}
	return names
}

// TestCheckout_ContentVerification 测试 checkout 后的内容与原始文件完全一致。
func TestCheckout_ContentVerification(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	original := "Hello, World! 你好世界\n第二行\n"
	file := env.WriteFile("greeting.txt", original)
	env.Run("add", file).Success(t)

	outDir := filepath.Join(env.Home, "checkedout")
	env.Mkdir("checkedout")
	env.Run("checkout", "greeting.txt", "--output", outDir).Success(t)

	actual := env.ReadFile(filepath.Join(outDir, "greeting.txt"))
	if actual != original {
		t.Fatalf("content mismatch:\ngot:  %q\nwant: %q", actual, original)
	}
}

// TestCheckout_MissingAsset 测试 checkout 不存在的资产应报错。
func TestCheckout_MissingAsset(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	outDir := filepath.Join(env.Home, "out")
	env.Mkdir("out")
	env.Run("checkout", "nonexistent", "--output", outDir).
		Fail(t).
		ContainsErr(t, "还原失败")
}

// TestRemove_MissingAsset 测试 remove 不存在的资产应报错。
func TestRemove_MissingAsset(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	env.Run("remove", "nonexistent").
		Fail(t).
		ContainsErr(t, "不存在")
}

// TestAdd_NonexistentPath 测试 add 不存在的路径应报错。
func TestAdd_NonexistentPath(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	env.Run("add", filepath.Join(env.Home, "no_such_file.txt")).
		Fail(t).
		ContainsErr(t, "无法访问")
}

// TestBinaryFile 测试二进制文件的完整生命周期。
// 注意：Distill 会对所有文件做 CRLF 规范化（包括二进制文件），
// 因此测试数据不能包含 0x0D 0x0A 字节序列。
func TestBinaryFile(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	// 创建一个不含 CRLF 序列的二进制文件（模拟 TIFF header）
	binData := []byte{0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00, 0x01, 0x00}
	binPath := env.WriteBinaryFile("image.tif", binData)
	env.Run("add", binPath).Success(t)

	// checkout
	outDir := filepath.Join(env.Home, "img-out")
	env.Mkdir("img-out")
	env.Run("checkout", "image.tif", "--output", outDir).Success(t)

	// 验证二进制内容一致
	actual := env.ReadBinaryFile(filepath.Join(outDir, "image.tif"))
	if string(actual) != string(binData) {
		t.Fatalf("binary content mismatch: got %x, want %x", actual, binData)
	}
}

// TestConfig_ShowDefaults 显示所有默认配置（无需 init）。
func TestConfig_ShowDefaults(t *testing.T) {
	env := Env(t)

	result := env.Run("config", "show").Success(t)

	// 验证所有配置项都有输出
	result.StdoutContains(t,
		"core.version=", "core.objects_format=",
		"checkout.overwrite=", "log.format=",
		"log.level=", "normalize.crlf_to_lf=", "lang=",
	)
}

// TestConfig_GetSingleKey 查询单个配置项。
func TestConfig_GetSingleKey(t *testing.T) {
	env := Env(t)

	result := env.Run("config", "get", "checkout.overwrite").Success(t)
	result.StdoutContains(t, "checkout.overwrite=", "ask")

	result2 := env.Run("config", "get", "lang").Success(t)
	result2.StdoutContains(t, "lang=", "zh")
}

// TestConfig_GetUnknownKey 查询未知配置项应失败。
func TestConfig_GetUnknownKey(t *testing.T) {
	env := Env(t)

	env.Run("config", "get", "nonexistent.key").Fail(t)
}

// TestConfig_SetAndPersist set 后值持久化，get 可读回。
func TestConfig_SetAndPersist(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	// set 修改值
	env.Run("config", "set", "checkout.overwrite", "force").Success(t)

	// get 读回验证
	result := env.Run("config", "get", "checkout.overwrite").Success(t)
	result.StdoutContains(t, "force")

	// 验证 config.toml 已写入磁盘
	configPath := filepath.Join(env.Home, "config", "config.toml")
	data := env.ReadFile(configPath)
	if !strings.Contains(data, "overwrite = \"force\"") {
		t.Fatalf("config.toml should contain overwrite = \"force\", got:\n%s", data)
	}
}

// TestConfig_SetInvalidValue set 非法值应失败。
func TestConfig_SetInvalidValue(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	// checkout.overwrite 只接受 ask/skip/force
	env.Run("config", "set", "checkout.overwrite", "invalid").Fail(t)

	// lang 只接受 zh/en
	env.Run("config", "set", "lang", "fr").Fail(t)

	// log.level 只接受 debug/info/warn/error
	env.Run("config", "set", "log.level", "trace").Fail(t)
}

// TestConfig_SetUnknownKey set 未知配置项应失败。
func TestConfig_SetUnknownKey(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	env.Run("config", "set", "foo.bar", "baz").Fail(t)
}

// TestConfig_TomlIntegrity 验证 set 写入的 TOML 文件格式正确、可被重新解析。
func TestConfig_TomlIntegrity(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	// 连续 set 多个配置项
	env.Run("config", "set", "checkout.overwrite", "skip").Success(t)
	env.Run("config", "set", "log.level", "debug").Success(t)
	env.Run("config", "set", "normalize.crlf_to_lf", "false").Success(t)
	env.Run("config", "set", "lang", "en").Success(t)

	// config show 应该反映所有修改
	result := env.Run("config", "show").Success(t)
	result.StdoutContains(t,
		"checkout.overwrite=skip",
		"log.level=debug",
		"normalize.crlf_to_lf=false",
		"lang=en",
	)

	// 验证 TOML 文件可读且完整
	configPath := filepath.Join(env.Home, "config", "config.toml")
	data := env.ReadFile(configPath)
	if !strings.Contains(data, "overwrite = \"skip\"") {
		t.Fatalf("TOML missing overwrite=skip")
	}
	if !strings.Contains(data, "level = \"debug\"") {
		t.Fatalf("TOML missing level=debug")
	}
	if !strings.Contains(data, "crlf_to_lf = false") {
		t.Fatalf("TOML missing crlf_to_lf=false")
	}
	if !strings.Contains(data, "lang = \"en\"") {
		t.Fatalf("TOML missing lang=en")
	}

	// set 后其他命令仍然正常工作（TOML 没有损坏）
	file := env.WriteFile("after-config.txt", "still works\n")
	env.Run("add", file).Success(t)
	env.Run("list").Success(t).Contains(t, "after-config.txt")
}

// TestZipImport 测试 ZIP 文件导入。
func TestZipImport(t *testing.T) {
	env := Env(t)

	env.Run("init", "--trash", filepath.Join(env.Home, ".trash")).Success(t)

	// 创建一个 ZIP 文件
	zipPath := filepath.Join(env.WorkDir, "archive.zip")
	createTestZip(t, zipPath, map[string]string{
		"file1.txt": "content 1\n",
		"file2.txt": "content 2\n",
	})

	env.Run("add", zipPath).Success(t)
	// ref 注册名为 filepath.Base(source) = "archive.zip"
	env.Run("list").Success(t).Contains(t, "archive.zip")

	// checkout 并验证
	outDir := filepath.Join(env.Home, "archive-out")
	env.Mkdir("archive-out")
	env.Run("checkout", "archive.zip", "--output", outDir).Success(t)

	// ZipAdapter 树根名为 "archive"（去掉 .zip 后缀），输出目录为 archive/
	actual1 := env.ReadFile(filepath.Join(outDir, "archive", "file1.txt"))
	if actual1 != "content 1\n" {
		t.Fatalf("file1.txt content mismatch: got %q", actual1)
	}
	actual2 := env.ReadFile(filepath.Join(outDir, "archive", "file2.txt"))
	if actual2 != "content 2\n" {
		t.Fatalf("file2.txt content mismatch: got %q", actual2)
	}
}

// createTestZip 创建一个包含指定文件的 ZIP 文件。
func createTestZip(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, content := range files {
		writer, err := w.Create(name)
		if err != nil {
			t.Fatalf("failed to create zip entry: %v", err)
		}
		if _, err := io.WriteString(writer, content); err != nil {
			t.Fatalf("failed to write zip entry: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
}

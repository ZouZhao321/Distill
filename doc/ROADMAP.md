# Distill 开发路线图

> 开发准则：严格遵循 `PRINCIPLES.md`，先建接口，再写实现。

## Phase 1: 夯实底座 - CAS 对象与头部机制

**目标**：解决最核心的“存”与“取”，特别是 Git 风格头部的拼接与剥离。

- [ ] 创建 `internal/cas/blob.go`：实现 `NormalizeText()` 文本规范化。
- [ ] 创建 `internal/cas/object.go`：
  - 实现 `Store(content []byte) (string, error)`：拼接头部，算哈希，流式写入 `.objects/`。
  - 实现 `Read(hash string) (io.Reader, error)`：读取对象，**返回一个自动剥离了头部的 `io.Reader`**（为后续流式导出做准备）。
- [ ] 更新 `cmd/hash.go`：调用底层 Store 方法。
**验收**：`add a.txt` 后，`.objects/` 生成文件；通过底层 Read 方法能读出不含头部的原始文本。

## Phase 2: 万物皆树 - 接口定义与基础适配器

**目标**：建立系统边界，将散装文件和目录转换为虚拟树。

- [ ] 创建 `internal/model/tree.go`：定义 `Tree` 结构体（Path, Type, Hash, Children）。
- [ ] 创建 `internal/adapter/source.go`：定义核心接口 `type Source interface { Parse() (*model.Tree, error) }`。
- [ ] 创建 `internal/adapter/file.go`：实现 `FileSource`，将单文件包装为扁平树。
- [ ] 创建 `internal/adapter/dir.go`：实现 `DirSource`，遍历目录生成树（调用 Phase 1 的 Store）。
- [ ] 创建 `cmd/scan.go`：接收路径，根据类型调度 Adapter，打印树的 JSON。
**验收**：`scan 1.txt` 和 `scan ./folder` 输出的都是标准化的 Tree JSON 结构。

## Phase 3: 穿透包装 - 内存虚拟解压

**目标**：消灭 ZIP 包的物理形态，将其在内存中拆解为树。

- [ ] 创建 `internal/adapter/zip.go`：实现 `ZipSource`。
  - 使用 `archive/zip` 在内存中遍历。
  - 遇到文件直接 `Open()` 拿到 Reader，送入 CAS 算哈希生成叶子节点。
  - 遇到目录生成树节点。
- [ ] 更新 `cmd/scan.go`：增加 ZIP 后缀判定，调度 `ZipSource`。
**验收**：`scan claude.zip` 输出的 JSON 树，与解压后 `scan ./claude` 的 JSON 树逻辑完全一致。

## Phase 4: 核心闭环 - 冗余判定与清单管理

**目标**：实现完整的 `add` 和 `list` 业务。

- [ ] 创建 `internal/cas/tree_hash.go`：实现递归计算整棵树的组合哈希。
- [ ] 创建 `internal/engine/manifest.go`：定义 Manifest 结构，实现清单的 JSON 序列化存入 `.manifests/`。
- [ ] 重构 `cmd/add.go`：
  - 调度 Adapter 拿到树。
  - 计算树哈希，扫描已有 Manifest。
  - 执行冗余判定（相同则追加来源，不同则新建）。
- [ ] 创建 `cmd/list.go`：遍历 `.manifests/` 格式化输出。
**验收**：连续 `add` 两个内容一样但名字不同的包，`list` 只显示一条记录，且来源数量为 2。

## Phase 5: 逆向还原 - Checkout 与 Export

**目标**：从图纸和碎片中拼装出可用资产。

- [ ] 创建 `internal/engine/restore.go`：
  - 实现 `Checkout(manifestPath, targetDir string)`：遍历树，调用 `cas.Read()` 获取剥离头部后的 Reader，写入目标文件。
  - 实现 `Export(manifestPath, outputZip string)`：使用 `archive/zip` 创建 Writer，将 `cas.Read()` 的 Reader 直接 Pipe 进压缩包。
- [ ] 创建 `cmd/checkout.go` 和 `cmd/export.go`。
**验收**：`export` 出来的 zip，解压后没有任何乱七八糟的嵌套目录，直接就是干净的源码根目录，且可用 IDE 正常打开。

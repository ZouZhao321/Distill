# 产品需求文档 (PRD)：Distill

## 1. 核心用户故事

- **【作为】技术研究者**，我执行 `add` 扔进多个同名源码包，系统自动识别冗余，底层只存一份，帮我理清脉络。
- **【作为】资料收集者**，我直接 `add` 一个单 PDF，系统自动将其视为完整资产并识别出各种改名的“同源 PDF”。
- **【作为】深度阅读者**，我通过 `checkout` 导出版本，得到无冗余嵌套的干净文件夹。
- **【作为】资产整理者**，我通过 `export` 将资产打包为标准 zip，方便归档分享。
- **【作为】保守型用户**，系统中无彻底删除功能，我只能移入回收站，由我决定最终生杀大权。

---

## 2. 核心业务流程

### 2.1 资产入库与去重 (`distill add <path>`)

此流程是产品核心。

````mermaid
flowchart TD
    Start([用户执行 distill add path]) --> Detect{路径类型判定}
    subgraph 边界外: 现实混沌域 [适配器层 Adapter]
        Detect -->|单文件| FileAdapter[FileAdapter: 向上组装为扁平树]
        Detect -->|文件夹| DirAdapter[DirAdapter: 映射为多级树]
        Detect -->|压缩包| ZipAdapter[ZipAdapter: 内存拆解为树]
        FileAdapter --> VTree
        DirAdapter --> VTree
        ZipAdapter --> VTree
        VTree[(输出: 标准虚拟树)]
    end
    VTree --> Engine
    subgraph 边界内: 纯粹秩序域 [引擎层 Engine]
        Engine --> Traverse[递归遍历虚拟树节点]
        Traverse --> TypeCheck{节点类型?}
        TypeCheck -->|文本文件| Normalize[文本规范化: 统一换行符等]
        TypeCheck -->|二进制文件| Raw[保持原样]
        Normalize --> AddHeader[拼接 Git 头部: blob size\0content]
        Raw --> AddHeader
        AddHeader --> CalcHash[计算 SHA-256]
        CalcHash --> Store{.objects/ 中已存在?}
        Store -->|否| WriteObj[写入 .objects/hash]
        Store -->|是| Skip[跳过写入]
        WriteObj --> BuildTree
        Skip --> BuildTree
        BuildTree[构建完整树指纹] --> Diff{与已有清单指纹对比}
        Diff -->|指纹一致: 冗余| AppendSource[在旧清单追加来源记录]
        Diff -->|指纹不一致: 新资产| GenManifest[生成新 .manifests/hash.json]
        AppendSource --> End([输出提纯报告])
        GenManifest --> End
    end
    classDef adapter fill:#f9f,stroke:#333,stroke-width:2px;
    classDef engine fill:#bbf,stroke:#333,stroke-width:2px;
    class FileAdapter,DirAdapter,ZipAdapter adapter;
    class Traverse,TypeCheck,Normalize,Raw,AddHeader,CalcHash,Store,WriteObj,Skip,BuildTree,Diff,AppendSource,GenManifest engine;
````

### 2.2 逆向组装导出 (`distill checkout <hash> <dir>`)

将图纸还原为散装的物理文件夹，核心在于“头部剥离”。

````mermaid
flowchart TD
    Start([用户执行 checkout]) --> ReadM[读取 .manifests/hash.json]
    ReadM --> CreateDir[在 target_dir 创建空文件夹骨架]
    CreateDir --> LoopNode{遍历清单文件节点}
    LoopNode --> GetHash[获取节点对应的内容哈希]
    GetHash --> ReadObj[从 .objects/ 读取数据流]
    ReadObj --> StripHeader["核心动作: 精准切掉 blob size\0 头部"]
    StripHeader --> WriteFile[将纯净内容写入目标路径]
    WriteFile --> LoopNode
    LoopNode -->|遍历结束| End([导出完成, 可用IDE打开])

````

### 2.3 标准打包导出 (`distill export <hash> <file.zip>`)

将图纸打包为单文件压缩包，核心在于“流式处理，零落盘”。

````mermaid
flowchart TD
    Start([用户执行 export]) --> ReadM[读取 .manifests/hash.json]
    ReadM --> InitZip[在内存初始化 zip.NewWriter]
    InitZip --> LoopNode{遍历清单文件节点}
    LoopNode --> ReadObj[流式打开 .objects/ 文件]
    ReadObj --> PipeSkip["核心动作: 流式跳过 blob size\0 头部"]
    PipeSkip --> ZipWrite[通过 io.Pipe 直接写入 Zip Writer]
    ZipWrite --> LoopNode
    LoopNode -->|遍历结束| CloseZip[关闭 Zip Writer]
    CloseZip --> FlushDisk[内存数据一次性 flush 到硬盘生成 zip]
    FlushDisk --> End([打包完成, 可直接分享])

````

### 2.4 安全移除 (`distill remove <hash>`)

严格遵守“系统无删除能力”的红线设计。

````mermaid
flowchart TD
    Start([用户执行 remove]) --> FindM[定位 .manifests/hash.json]
    FindM --> CheckTrash{检查 .trash/ 目录是否存在}
    CheckTrash -->|否| MkTrash[创建 .trash/ 目录]
    CheckTrash -->|是| MoveFile
    MkTrash --> MoveFile["执行 os.Rename (严禁 os.Remove)"]
    MoveFile --> Trash[清单移入 .trash/]
    Trash --> LockObj["铁律: 绝对不触碰 .objects/ 目录"]
    LockObj --> End([提示: 已移入回收站])

````

## 3. 技术与非功能约束

T1. 低依赖底线：核心逻辑（哈希、解压、IO）强制使用 Go 标准库（crypto/sha256, archive/zip, path/filepath 等

T2. 流式处理：大文件处理强制 io.Pipe 或 io.MultiWriter 实现流式穿透，防 OOM。

N1. 绝对安全（防误删防破坏）：代码层面严禁调用任何物理删除函数（如 os.Remove）。所有移除动作必须实现为 os.Rename 至 .trash/，将清理权交还用户。

N2. 原始数据 untouched：add 操作对用户指定的源 `<path>` 具有只读权限。

N3. 降级容错：遇到损坏文件或非支持格式，打印 Warn 日志并跳过，不中断整体流程

---

# Distill (蒸馏)

一个基于内容寻址存储（CAS）的本地数字资产提纯与结构去重引擎。

[English](../README.md)

## 灵感与定位

受 Git 底层机制启发。面对海量、杂乱、多重嵌套、高度重合的数字资产（如源码、报告、数据集），
Distill 能够**无视冗余外皮，提取纯净数字内核**，以极简物理空间存储，并支持按需逆向组装为干净的目录或压缩包。

## 核心特性

- **物理级去重**：基于 SHA-256 内容寻址，全盘只存一份数据。
- **结构级去重**：引入 Merkle Tree 思想，精准识别"换皮不换芯"的复合资产（如改名压缩包、改换行符的文本）。
- **万物皆树**：单文件、文件夹、ZIP 包在引擎眼中均为统一的"虚拟树"，架构极度解耦。
- **绝对安全**：系统无物理删除能力，仅提供移入回收站机制。
- **流式穿透**：处理几十万行源码或大压缩包时，纯内存流式处理，不落盘临时文件，拒绝 OOM。

## 快速开始

**编译：**

```bash
go build -o distill
```

**导入资产（支持单文件、文件夹、ZIP）：**

```bash
./distill add ./messy-source-code.zip
```

**查看已提纯的资产清单：**

```bash
./distill list
```

**导出为干净的可阅读文件夹：**

```bash
./distill checkout <manifest-hash> ./clean-code
```

**导出为干净的压缩包（适合归档分享）：**

```bash
./distill export <manifest-hash> ./archive.zip
```

## 文档

- [产品需求文档 (PRD)](PRD.md)
- [设计原则](PRINCIPLES.md)
- [路线图](ROADMAP.md)

## 贡献

请参阅 [贡献指南](../CONTRIBUTING.md)。

## 许可证

待定

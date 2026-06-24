# Distill

A content-addressable storage (CAS) engine for digital asset deduplication.

[中文文档](docs/README.zh-CN.md)

## Inspiration & Positioning

Inspired by Git's underlying mechanism. Distill extracts pure digital cores from messy, nested, highly overlapping digital assets (source code, reports, datasets), stores them with minimal physical space, and supports on-demand reverse assembly into clean directories or archives.

## Key Features

- **Physical-level deduplication**: SHA-256 content addressing, only one copy of data stored globally.
- **Structural-level deduplication**: Merkle Tree approach, accurately identifies "re-skinned" composite assets (renamed ZIPs, text with modified line breaks).
- **Everything is a tree**: Single files, directories, and ZIP archives are treated as unified "virtual trees" in the engine, achieving extreme decoupling.
- **Absolute safety**: No physical deletion capability, only recycle bin mechanism.
- **Streaming processing**: Pure memory streaming for large files/archives, no temporary files, no OOM.

## Quick Start

**Build:**

```bash
go build -o distill
```

**Import asset (supports single file, directory, or ZIP):**

```bash
./distill add ./messy-source-code.zip
```

**List imported assets:**

```bash
./distill list
```

**Export as clean directory:**

```bash
./distill checkout <manifest-hash> ./clean-code
```

**Export as clean archive:**

```bash
./distill export <manifest-hash> ./archive.zip
```

## Documentation

- [Product Requirements (PRD)](docs/PRD.md)
- [Design Principles](docs/PRINCIPLES.md)
- [Roadmap](docs/ROADMAP.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[GNU General Public License v3.0](LICENSE)

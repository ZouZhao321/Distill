# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Distill (蒸馏) is a content-addressable storage (CAS) engine for digital asset deduplication, written in Go. Inspired by Git's underlying mechanism, it extracts pure digital cores from messy, nested, highly overlapping assets (source code, reports, datasets) with minimal storage space, and supports on-demand reverse assembly into clean directories or archives.

Core features:
- **Physical-level deduplication**: SHA-256 content addressing, only one copy of data stored globally
- **Structural-level deduplication**: Merkle Tree思想, accurately identifies "re-skinned" composite assets (renamed zip files, text with modified line breaks)
- **Everything is a Tree**: Single files, directories, ZIP archives are treated as unified "Virtual Tree" in core logic
- **Absolute safety**: No physical deletion capability, only recycle bin mechanism
- **Streaming processing**: Pure memory streaming for large files/archives, no temporary files, no OOM

## Common Commands

### Make Commands
- Build binary: `make build` (outputs versioned binary with ldflags)
- Run all unit tests with coverage: `make test`
- Run E2E tests (e2e tag): `make test-e2e`
- Run linter (go vet): `make lint`
- Clean build artifacts: `make clean`
- Build cross-platform release binaries: `make release`
- Create version tag: `make tag VERSION=x.y.z`

### Direct Go Commands
- Simple build: `go build -o distill.exe`
- Run all unit tests: `go test ./...`
- Run a single test: `go test ./path/to/package -run TestName`
- Run E2E tests: `go test ./test/e2e/ -tags=e2e`

### Usage
- Import asset (file/folder/zip): `./distill.exe add <path/to/asset>`
- List assets: `./distill.exe list`
- Export asset to directory: `./distill.exe checkout <manifest-hash> <output-path>`
- Export asset to zip: `./distill.exe export <manifest-hash> <output.zip>`
- Run garbage collection: `./distill.exe gc`

## High-Level Architecture

Distill uses a hexagonal (ports & adapters) architecture pattern:

### Directory Structure
- **`cmd/`**: Cobra CLI layer, handles command routing and user input/output
- **`internal/adapter/`**: Input adapters that convert external assets (single files, directories, ZIP archives) into the unified Virtual Tree representation
- **`internal/core/`**: Pure business logic, no external dependencies
  - `domain/`: Core entities and value objects (Virtual Tree nodes, hashes, manifests)
  - `port/`: Abstract interfaces for external dependencies (storage, etc.)
  - `usecase/`: Implementation of all business operations (add, list, checkout, export, gc, remove)
- **`internal/infra/`**: Infrastructure implementations of core ports (on-disk storage, blob store)
- **`locales/`**: i18n translation files in TOML format
- **`test/e2e/`**: End-to-end black box tests covering full lifecycle scenarios

### Core Architectural Principle: "Everything is a Tree"
All external inputs are converted to a unified **Virtual Tree** data structure before entering the core domain. This means:
- Single files, directories, and ZIP archives are treated identically in core logic
- No format-specific checks (e.g. `if isZip`, `filepath.Ext()`) are allowed in `internal/core/`
- Core logic only operates on Virtual Trees, never on raw file formats
- Any new input type must implement an adapter that converts it to Virtual Tree format

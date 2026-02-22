# Copilot Instructions

## Build, test, and lint commands
- Go version: `1.25.3` (from `go.mod`).
- Sync module dependencies: `go mod tidy`
- Build all packages: `go build ./...`
- Run full test suite: `go test ./...`
- Run a single test: `go test ./path/to/package -run '^TestName$'`
- Static analysis (lint-style check used in this repo): `go vet ./...`
- Current repository state: no Go packages are checked in yet, so `./...` currently matches no packages.

## High-level architecture
- This repository is currently a minimal Go module bootstrap with dependency manifests only (`go.mod`, `go.sum`).
- No application/library packages, binaries, or entrypoints are committed yet.
- Dependency footprint indicates planned integration around:
  - `github.com/github/copilot-sdk/go`
  - `github.com/google/jsonschema-go`

## Key conventions
- Module path is `xtldr`; run Go commands from repository root.
- Dependency management is Go Modules only (no Makefile/task runner configured in-repo).
- Keep `go.mod` / `go.sum` authoritative by running `go mod tidy` after dependency or import changes.

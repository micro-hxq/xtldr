# xtldr

`xtldr` is a Copilot-powered terminal assistant that turns a natural-language request into executable command candidates.

## 🎬 Demo

[![asciicast](https://asciinema.org/a/fadSiG25I7sqHW2C.svg)](https://asciinema.org/a/fadSiG25I7sqHW2C)

## ✅ Prerequisite

- 🤖 GitHub Copilot CLI is required.
- 🔗 Install guide: [Installing GitHub Copilot CLI](https://docs.github.com/en/copilot/how-tos/copilot-cli/set-up-copilot-cli/install-copilot-cli)
- 🛠 Go environment is required (`go 1.25.3+` as defined in `go.mod`).
- 🔗 Install Go: [Download and install Go](https://go.dev/doc/install)

## ✨ Features

- 🤖 Generate up to 5 executable command candidates from your request
- ⌨️ Navigate candidates with `↑/↓` or `j/k`
- 📋 Press `Enter` to output the selected command to stdout and auto-copy it
- 🧠 Show a rich explanation panel with parameter meanings by default
- 🙈 Use `-e` to hide the explanation panel
- ⏳ Animated loading state while waiting for Copilot results

## 🚀 Usage

```bash
xtldr [flags] <request>
xtldr help
xtldr version
```

### Flags

- `-e` Hide the **Command Explanation** panel
- `-v`, `--version` Print version information

### Examples

```bash
xtldr "find large files in current directory"
xtldr "show top 10 processes by memory on macOS"
xtldr -e "show top 10 processes by memory on macOS"
xtldr help
xtldr version
```

## ℹ️ Help & Version Commands

- `xtldr help`: Show complete CLI help
- `xtldr version`: Show version, commit, and build date

## 🧭 Interactive Controls

- `↑/↓` or `j/k`: Move selection
- `Enter`: Print selected command to stdout + copy to clipboard
- `c`: Copy selected command
- `q` or `Ctrl+C`: Quit

## 🛠 Development

```bash
go test ./...
go build ./cmd/xtldr
```

## 📦 Release Build (Version Metadata)

```bash
go build -ldflags "-X main.Version=v1.0.0 -X main.Commit=$(git rev-parse --short HEAD) -X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" ./cmd/xtldr
```

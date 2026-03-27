# Contributing to Copilot CLI Session Monitor

Thanks for your interest in contributing! Here's how to get started.

## Getting Started

1. Fork and clone the repository
2. Ensure you have **Go 1.21+** installed
3. Build and run:
   ```bash
   make run
   ```

## Making Changes

1. Create a branch from `main`:
   ```bash
   git checkout -b my-feature
   ```
2. Make your changes
3. Ensure the project builds and passes vet:
   ```bash
   make build
   go vet ./...
   ```
4. Commit with a clear message describing the change
5. Open a pull request against `main`

## Project Structure

```
internal/
├── session/    # Reads Copilot session data from SQLite + lock files
├── menu/       # Builds the macOS menu bar UI via systray
├── monitor/    # Refresh timer that ties session reader to menu
└── terminal/   # Terminal.app tab switching via AppleScript
```

## Guidelines

- **Keep it lightweight** — this is a menu bar utility, not a full GUI app
- **Read-only** — never modify Copilot's files
- **No network calls** — all data comes from the local filesystem
- **Handle errors gracefully** — a broken session shouldn't crash the app
- **Clean up goroutines** — use the cancel channel pattern (see `menu/builder.go`)

## Reporting Issues

Please open a GitHub issue with:
- Your macOS version
- Go version (`go version`)
- Steps to reproduce
- Expected vs actual behaviour

## Code of Conduct

Be kind, be constructive. We're all here to build something useful.

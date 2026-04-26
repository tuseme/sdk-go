# Contributing to Tuseme Go SDK

Thank you for your interest in contributing!

## Development Setup

```bash
git clone https://github.com/tuseme/sdk-go.git
cd sdk-go
go mod download
```

## Running Tests

```bash
go test -race ./...
```

## Pull Requests

1. Fork the repo and create a feature branch from `main`.
2. Add tests for any new functionality.
3. Ensure `go vet ./...` and all tests pass.
4. Open a PR with a clear description of the change.

## Reporting Issues

Open an issue at [github.com/tuseme/sdk-go/issues](https://github.com/tuseme/sdk-go/issues) with:
- SDK version and Go version
- Minimal reproduction steps
- Expected vs. actual behavior

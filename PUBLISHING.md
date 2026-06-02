# Publishing `sdk-go`

This document describes how `sdk-go` is versioned and published for Go users.

Unlike Python, npm, NuGet, and Ruby, Go modules are generally published from the Git repository itself. There is no separate upload step to a central package registry.

## How Go publishing works

Users install directly from the module path:

```bash
go get github.com/tuseme/sdk-go
go get github.com/tuseme/sdk-go@v1.0.0
```

Go resolves module versions from Git tags and serves them through the Go module proxy/cache ecosystem.

## Module identity (must match exactly)

`go.mod` must use the repository path:

```go
module github.com/tuseme/sdk-go
go 1.22
```

If this path is wrong, installs fail or fetch a different module.

## Repository readiness checklist

Before releasing:

- `go.mod` exists and module path is correct
- `README.md` has install + usage examples
- `LICENSE` exists
- Public API is exported (`Client`, `NewClient`, etc.)
- CI is green
- Semantic version tag is planned (`vX.Y.Z`)

## Quality checks (required)

Run from repository root:

```bash
go fmt ./...
go vet ./...
go test ./...
```

Current repository status:

- `go fmt` succeeds
- `go vet` succeeds
- `go test` succeeds (`[no test files]` is acceptable if intentional)

## Release process

### 1) Prepare version

- Update `CHANGELOG.md` with release notes for `X.Y.Z`
- Ensure code on `main` is final and CI is passing

### 2) Create and push semantic tag

```bash
git checkout main
git pull origin main

git tag -a vX.Y.Z -m "vX.Y.Z — Release"
git push origin vX.Y.Z
```

Go module versioning depends on semantic tags with `v` prefix.

### 3) Publish GitHub Release

Create a GitHub Release from tag `vX.Y.Z`:

- Releases → Draft new release
- Select `vX.Y.Z`
- Add release notes
- Publish release

This is strongly recommended for visibility/changelog, though module availability comes from the tag itself.

## CI/CD behavior in this repo

### `ci.yml`

- Runs on push/PR to `main`
- Matrix tests with Go `1.21`, `1.22`, `1.23`
- Executes:
  - `go vet ./...`
  - `go test -race -coverprofile=coverage.out ./...`

### `publish.yml`

- Runs when GitHub Release is published
- Validates tag format
- Checks module with:
  - `go vet ./...`
  - `go test ./...`
- Prints verification URLs and module version context

Note: Go does not require uploading artifacts in this workflow.

## Verification after release

Wait a few minutes, then verify from a clean directory:

```bash
go list -m github.com/tuseme/sdk-go@vX.Y.Z
go get github.com/tuseme/sdk-go@vX.Y.Z
```

Optional proxy checks:

```bash
go env GOPROXY
curl "https://proxy.golang.org/github.com/tuseme/sdk-go/@v/vX.Y.Z.info"
```

Optional documentation check:

```text
https://pkg.go.dev/github.com/tuseme/sdk-go@vX.Y.Z
```

## Important command clarification

If you run:

```bash
go install github.com/tuseme/sdk-go@latest
```

and get:

```text
package github.com/tuseme/sdk-go is not a main package
```

that is expected and **not an error in the SDK**.

Reason:
- `go install` installs executable commands (`package main`)
- `sdk-go` is a library module, not a CLI binary

Use `go get` (or import in code) for library modules.

## Troubleshooting

### Module version not found

- Confirm tag exists on GitHub: `git ls-remote --tags origin`
- Ensure tag is semantic and prefixed (`v1.0.0`)
- Wait for proxy cache propagation (can take a few minutes)

### `go get` still pulls old version

- Clear module cache in local test environment:

```bash
go clean -modcache
```

- Retry with explicit version:

```bash
go get github.com/tuseme/sdk-go@vX.Y.Z
```

### Install confusion (`go install` vs `go get`)

- Use `go install ...` only for command-line tools (main packages)
- Use `go get ...` for libraries like this SDK

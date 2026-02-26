# Contributing to Oak

Thank you for considering a contribution! This document outlines how to get
started.

## Prerequisites

- **Go 1.25+** — [download](https://go.dev/dl/)
- **golangci-lint** — `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

## Getting Started

```bash
git clone https://github.com/ARTM2000/oak.git
cd oak
make test
```

## Development Workflow

1. **Fork** the repository and create a feature branch from `main`.
2. **Write code** — keep the public API minimal and backwards-compatible.
3. **Add tests** — aim for ≥ 95% coverage on new code. Run `make cover` to
   check.
4. **Lint** — run `make lint` before committing. CI enforces the same rules.
5. **Open a PR** — describe *what* changed and *why*.

## Running Tests

```bash
# Unit tests with race detector
make test

# Benchmarks
make bench

# Coverage report (opens coverage.html)
make cover
```

## Code Style

- Follow standard Go conventions (`gofmt`, `goimports`).
- Every exported type, function, and method must have a godoc comment.
- Avoid external dependencies — oak uses only the standard library.
- Keep the core API surface small; prefer additive changes.

## Commit Messages

Write clear, concise commit messages. Use the imperative mood:

```
Add transient lifetime support for named providers
Fix circular dependency detection when graph has multiple roots
```

## Versioning

Oak follows [Semantic Versioning](https://semver.org/):

- **v0.x.x** — API is evolving; breaking changes may occur in minor versions.
- **v1.0.0+** — public API is stable; breaking changes require a major bump.

## Reporting Issues

Open an issue with:

- Go version (`go version`)
- Minimal reproduction code
- Expected vs actual behaviour

## License

By contributing you agree that your contributions will be licensed under the
[MIT License](LICENSE).

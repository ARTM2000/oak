# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `Container.Shutdown(ctx)` for graceful shutdown of `io.Closer` singletons
  in reverse dependency order, with context-based timeout support.
- `ErrAlreadyShutdown` sentinel error for repeated shutdown calls.

## [0.1.0] - 2026-02-27

### Added
- `Container` interface with `Register`, `RegisterNamed`, `Build`, `Resolve`,
  and `ResolveNamed` methods.
- Generic helpers `oak.Resolve[T]` and `oak.ResolveNamed[T]` for type-safe
  resolution.
- `Singleton` (default) and `Transient` lifetime support via `WithLifetime`
  option.
- Named provider registration for multiple implementations of the same type.
- Circular dependency detection at build time with full chain in the error
  message.
- Duplicate provider detection at registration time.
- Build-time validation of the full dependency graph, including named
  providers.
- Thread-safe resolution after build.
- Comprehensive test suite (97%+ coverage) including concurrency and race
  detector tests.
- Benchmarks for `Register`, `Build`, `Resolve` (singleton & transient), and
  `ResolveNamed`.
- Testable examples for godoc.
- CI pipeline (GitHub Actions) with test, lint, and coverage.
- Release automation via GitHub Actions on tag push.

[Unreleased]: https://github.com/ARTM2000/oak/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ARTM2000/oak/releases/tag/v0.1.0

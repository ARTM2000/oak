# Oak

[![Go Reference](https://pkg.go.dev/badge/github.com/ARTM2000/oak.svg)](https://pkg.go.dev/github.com/ARTM2000/oak)
[![CI](https://github.com/ARTM2000/oak/actions/workflows/ci.yml/badge.svg)](https://github.com/ARTM2000/oak/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ARTM2000/oak)](https://goreportcard.com/report/github.com/ARTM2000/oak)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A lightweight dependency injection container for Go. Register constructors,
build once, resolve anywhere — with full type safety via generics.

## Features

- **Constructor injection** — dependencies are expressed as function parameters
- **Generics-first API** — `oak.Resolve[*DB](c)` with compile-time type safety
- **Singleton & Transient lifetimes** — one shared instance or a fresh one every time
- **Named providers** — multiple implementations of the same type
- **Circular dependency detection** — caught at build time with full chain in the error
- **Concurrency safe** — thread-safe resolution after build
- **Zero dependencies** — only the Go standard library

## Installation

```bash
go get github.com/ARTM2000/oak
```

Requires **Go 1.21** or later.

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/ARTM2000/oak"
)

type Logger struct{ Prefix string }
type Config struct{ DSN string }
type Database struct {
    Config *Config
    Logger *Logger
}

func main() {
    c := oak.New()

    // Register constructors — order does not matter.
    c.Register(func() *Config { return &Config{DSN: "postgres://localhost/app"} })
    c.Register(func() *Logger { return &Logger{Prefix: "app"} })
    c.Register(func(cfg *Config, log *Logger) *Database {
        return &Database{Config: cfg, Logger: log}
    })

    // Build validates the graph and creates all singletons.
    if err := c.Build(); err != nil {
        log.Fatal(err)
    }

    // Resolve with full type safety.
    db, err := oak.Resolve[*Database](c)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(db.Config.DSN)    // postgres://localhost/app
    fmt.Println(db.Logger.Prefix) // app
}
```

## Concepts

### Lifetimes

| Lifetime    | Behaviour                                        |
|-------------|--------------------------------------------------|
| `Singleton` | Created once during `Build()`. Same instance returned on every `Resolve()`. This is the **default**. |
| `Transient` | A new instance is constructed on every `Resolve()` call. |

```go
c.Register(NewLogger, oak.WithLifetime(oak.Transient))
```

### Named Providers

When you need several implementations of the same return type, register them
by name:

```go
c.RegisterNamed("mysql", NewMySQLDB)
c.RegisterNamed("postgres", NewPostgresDB)

// later …
db, _ := oak.ResolveNamed[*sql.DB](c, "postgres")
```

Named providers create a new instance on every `ResolveNamed` call. Their
dependencies are resolved from the typed provider pool.

### Build Phase

`Build()` does three things:

1. **Validates** the entire dependency graph — missing providers and circular
   dependencies are caught here, not at runtime.
2. **Instantiates** all singleton providers eagerly.
3. **Locks** the container — no further registrations are accepted.

### Constructor Signatures

Constructors must be functions with one of these return signatures:

```go
func(deps...) T
func(deps...) (T, error)
```

If a constructor returns `(T, error)` and the error is non-nil, `Build()`
(for singletons) or `Resolve()` (for transients) will propagate it.

## API Overview

| Function / Method                        | Description                                      |
|------------------------------------------|--------------------------------------------------|
| `oak.New() Container`                    | Create a new empty container                     |
| `c.Register(ctor, opts...) error`        | Register a typed constructor                     |
| `c.RegisterNamed(name, ctor, opts...) error` | Register a named constructor                 |
| `c.Build() error`                        | Validate graph and instantiate singletons        |
| `oak.Resolve[T](c) (T, error)`          | Resolve a type (generic, recommended)            |
| `oak.ResolveNamed[T](c, name) (T, error)` | Resolve a named provider (generic, recommended)|
| `c.Resolve(reflect.Type) (reflect.Value, error)` | Resolve by `reflect.Type`               |
| `c.ResolveNamed(name, reflect.Type) (reflect.Value, error)` | Resolve named by `reflect.Type` |

### Options

| Option                          | Description                                      |
|---------------------------------|--------------------------------------------------|
| `oak.WithLifetime(oak.Transient)` | Set the provider lifetime (default `Singleton`) |

### Sentinel Errors

All errors can be checked with `errors.Is`:

| Error                        | When                                            |
|------------------------------|--------------------------------------------------|
| `oak.ErrNotBuilt`           | `Resolve` called before `Build`                  |
| `oak.ErrAlreadyBuilt`       | `Register` or `Build` called after `Build`       |
| `oak.ErrProviderNotFound`   | No provider for the requested type or name       |
| `oak.ErrCircularDependency` | Dependency graph contains a cycle                |
| `oak.ErrDuplicateProvider`  | Same type or name registered twice               |

## Examples

See [`_examples/userapp`](_examples/userapp) for a runnable example that
wires up a small layered application.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md)
before opening a pull request.

## License

[MIT](LICENSE)

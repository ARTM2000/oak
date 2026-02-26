// Package oak provides a lightweight, reflection-based dependency injection
// container for Go.
//
// Oak uses constructor functions to wire dependencies automatically. Register
// constructors with the container, call [Container.Build] to validate the
// dependency graph, then retrieve fully-assembled objects with [Resolve] or
// [ResolveNamed].
//
// # Quick Start
//
//	c := oak.New()
//	c.Register(NewLogger)
//	c.Register(NewDatabase)
//	c.Build()
//
//	db, err := oak.Resolve[*Database](c)
//
// # Lifetimes
//
// [Singleton] (default) — one shared instance for the lifetime of the
// container.
//
// [Transient] — a fresh instance on every [Container.Resolve] call.
//
//	c.Register(NewLogger, oak.WithLifetime(oak.Transient))
//
// # Named Providers
//
// When you need several implementations of the same return type, use named
// registration:
//
//	c.RegisterNamed("mysql", NewMySQLDB)
//	c.RegisterNamed("postgres", NewPostgresDB)
//
//	db, _ := oak.ResolveNamed[Database](c, "postgres")
//
// # Graceful Shutdown
//
// Singleton providers that implement [io.Closer] are automatically tracked
// during [Container.Build]. Call [Container.Shutdown] to close them in reverse
// dependency order:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	if err := c.Shutdown(ctx); err != nil {
//	    log.Println("shutdown error:", err)
//	}
package oak

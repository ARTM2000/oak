package oak

import (
	"errors"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func TestRegister(t *testing.T) {
	t.Run("valid constructor", func(t *testing.T) {
		c := New()
		if err := c.Register(newTestLogger); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("constructor returning (T, error)", func(t *testing.T) {
		c := New()
		err := c.Register(func() (*testConfig, error) { return &testConfig{}, nil })
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("non-function is rejected", func(t *testing.T) {
		c := New()
		if err := c.Register("not a function"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("no return values rejected", func(t *testing.T) {
		c := New()
		if err := c.Register(func() {}); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("three return values rejected", func(t *testing.T) {
		c := New()
		if err := c.Register(func() (int, int, int) { return 0, 0, 0 }); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("second return not error rejected", func(t *testing.T) {
		c := New()
		if err := c.Register(func() (int, string) { return 0, "" }); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("after build returns ErrAlreadyBuilt", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustBuild(t, c)

		err := c.Register(newTestConfig)
		if !errors.Is(err, ErrAlreadyBuilt) {
			t.Fatalf("expected ErrAlreadyBuilt, got: %v", err)
		}
	})

	t.Run("duplicate type returns ErrDuplicateProvider", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)

		err := c.Register(func() *testLogger { return &testLogger{} })
		if !errors.Is(err, ErrDuplicateProvider) {
			t.Fatalf("expected ErrDuplicateProvider, got: %v", err)
		}
	})

	t.Run("with lifetime option", func(t *testing.T) {
		c := New()
		if err := c.Register(newTestLogger, WithLifetime(Transient)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// RegisterNamed
// ---------------------------------------------------------------------------

func TestRegisterNamed(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		c := New()
		if err := c.RegisterNamed("log", newTestLogger); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty name rejected", func(t *testing.T) {
		c := New()
		if err := c.RegisterNamed("", newTestLogger); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("duplicate name returns ErrDuplicateProvider", func(t *testing.T) {
		c := New()
		mustRegisterNamed(t, c, "log", newTestLogger)

		err := c.RegisterNamed("log", func() *testLogger { return &testLogger{} })
		if !errors.Is(err, ErrDuplicateProvider) {
			t.Fatalf("expected ErrDuplicateProvider, got: %v", err)
		}
	})

	t.Run("after build returns ErrAlreadyBuilt", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustBuild(t, c)

		err := c.RegisterNamed("log", newTestLogger)
		if !errors.Is(err, ErrAlreadyBuilt) {
			t.Fatalf("expected ErrAlreadyBuilt, got: %v", err)
		}
	})

	t.Run("same type can be named and typed", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		err := c.RegisterNamed("special", func() *testLogger { return &testLogger{Prefix: "special"} })
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// Build
// ---------------------------------------------------------------------------

func TestBuild(t *testing.T) {
	t.Run("empty container succeeds", func(t *testing.T) {
		c := New()
		if err := c.Build(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("single provider", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		if err := c.Build(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("dependency chain", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustRegister(t, c, newTestConfig)
		mustRegister(t, c, newTestDatabase)
		mustRegister(t, c, newTestUserRepo)
		mustRegister(t, c, newTestUserService)

		if err := c.Build(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("called twice returns ErrAlreadyBuilt", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustBuild(t, c)

		if err := c.Build(); !errors.Is(err, ErrAlreadyBuilt) {
			t.Fatalf("expected ErrAlreadyBuilt, got: %v", err)
		}
	})

	t.Run("missing dependency returns ErrProviderNotFound", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestDatabase) // needs *testConfig and *testLogger

		err := c.Build()
		if !errors.Is(err, ErrProviderNotFound) {
			t.Fatalf("expected ErrProviderNotFound, got: %v", err)
		}
	})

	t.Run("circular dependency detected", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestCircA)
		mustRegister(t, c, newTestCircB)
		mustRegister(t, c, newTestCircC)

		err := c.Build()
		if !errors.Is(err, ErrCircularDependency) {
			t.Fatalf("expected ErrCircularDependency, got: %v", err)
		}
	})

	t.Run("circular error includes chain", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestCircA)
		mustRegister(t, c, newTestCircB)
		mustRegister(t, c, newTestCircC)

		err := c.Build()
		if !strings.Contains(err.Error(), "->") {
			t.Fatalf("expected chain in error, got: %v", err)
		}
	})

	t.Run("constructor error propagates", func(t *testing.T) {
		c := New()
		mustRegister(t, c, func() (*testConfig, error) {
			return nil, errors.New("connection failed")
		})

		err := c.Build()
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "connection failed") {
			t.Fatalf("expected 'connection failed' in error, got: %v", err)
		}
	})

	t.Run("validates named provider dependencies", func(t *testing.T) {
		c := New()
		mustRegisterNamed(t, c, "order", newTestOrderService)

		err := c.Build()
		if !errors.Is(err, ErrProviderNotFound) {
			t.Fatalf("expected ErrProviderNotFound for named dep, got: %v", err)
		}
	})

	t.Run("named provider with satisfied deps builds", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustRegisterNamed(t, c, "order", newTestOrderService)

		if err := c.Build(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("transient provider validated but not instantiated", func(t *testing.T) {
		callCount := 0
		c := New()
		mustRegister(t, c, func() *testLogger {
			callCount++
			return &testLogger{Prefix: "app"}
		}, WithLifetime(Transient))
		mustBuild(t, c)

		if callCount != 0 {
			t.Fatalf("transient should not be constructed during build, called %d times", callCount)
		}
	})

	t.Run("singleton is eagerly instantiated", func(t *testing.T) {
		callCount := 0
		c := New()
		mustRegister(t, c, func() *testLogger {
			callCount++
			return &testLogger{Prefix: "app"}
		})
		mustBuild(t, c)

		if callCount != 1 {
			t.Fatalf("singleton should be constructed once during build, called %d times", callCount)
		}
	})
}

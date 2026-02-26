package oak

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// ---------------------------------------------------------------------------
// Resolve
// ---------------------------------------------------------------------------

func TestResolve(t *testing.T) {
	t.Run("before build returns ErrNotBuilt", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)

		_, err := c.Resolve(reflect.TypeOf((*testLogger)(nil)))
		if !errors.Is(err, ErrNotBuilt) {
			t.Fatalf("expected ErrNotBuilt, got: %v", err)
		}
	})

	t.Run("singleton returns same instance", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustBuild(t, c)

		v1, err := c.Resolve(reflect.TypeOf((*testLogger)(nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v2, _ := c.Resolve(reflect.TypeOf((*testLogger)(nil)))

		if v1.Pointer() != v2.Pointer() {
			t.Fatal("singleton should return the same instance")
		}
	})

	t.Run("transient returns different instances", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger, WithLifetime(Transient))
		mustBuild(t, c)

		v1, err := c.Resolve(reflect.TypeOf((*testLogger)(nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v2, _ := c.Resolve(reflect.TypeOf((*testLogger)(nil)))

		if v1.Pointer() == v2.Pointer() {
			t.Fatal("transient should return different instances")
		}
	})

	t.Run("transient constructor called each time", func(t *testing.T) {
		callCount := 0
		c := New()
		mustRegister(t, c, func() *testLogger {
			callCount++
			return &testLogger{}
		}, WithLifetime(Transient))
		mustBuild(t, c)

		callCount = 0
		_, _ = c.Resolve(reflect.TypeOf((*testLogger)(nil)))
		_, _ = c.Resolve(reflect.TypeOf((*testLogger)(nil)))
		_, _ = c.Resolve(reflect.TypeOf((*testLogger)(nil)))

		if callCount != 3 {
			t.Fatalf("expected 3 calls, got %d", callCount)
		}
	})

	t.Run("deep dependency chain fully resolved", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustRegister(t, c, newTestConfig)
		mustRegister(t, c, newTestDatabase)
		mustRegister(t, c, newTestUserRepo)
		mustRegister(t, c, newTestUserService)
		mustBuild(t, c)

		svc, err := Resolve[*testUserService](c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if svc.Repo == nil {
			t.Fatal("UserService.Repo is nil")
		}
		if svc.Repo.DB == nil {
			t.Fatal("UserRepo.DB is nil")
		}
		if svc.Repo.DB.Config == nil {
			t.Fatal("Database.Config is nil")
		}
		if svc.Repo.DB.Config.DSN != "postgres://localhost" {
			t.Fatalf("unexpected DSN: %s", svc.Repo.DB.Config.DSN)
		}
		if svc.Logger == nil {
			t.Fatal("UserService.Logger is nil")
		}
	})

	t.Run("singletons share instances across dependents", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustRegister(t, c, newTestConfig)
		mustRegister(t, c, newTestDatabase)
		mustRegister(t, c, newTestUserRepo)
		mustRegister(t, c, newTestUserService)
		mustBuild(t, c)

		svc, _ := Resolve[*testUserService](c)
		repo, _ := Resolve[*testUserRepo](c)
		logger, _ := Resolve[*testLogger](c)

		if svc.Logger != logger {
			t.Fatal("UserService should share Logger singleton")
		}
		if repo.Logger != logger {
			t.Fatal("UserRepo should share Logger singleton")
		}
		if repo.DB.Logger != logger {
			t.Fatal("Database should share Logger singleton")
		}
	})

	t.Run("transient with singleton dependency shares singleton", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustRegister(t, c, newTestOrderService, WithLifetime(Transient))
		mustBuild(t, c)

		s1, _ := Resolve[*testOrderService](c)
		s2, _ := Resolve[*testOrderService](c)

		if s1 == s2 {
			t.Fatal("transient should create different instances")
		}
		if s1.Logger != s2.Logger {
			t.Fatal("both transients should share the same singleton Logger")
		}
	})

	t.Run("singleton depending on transient captures one instance", func(t *testing.T) {
		callCount := 0
		c := New()
		mustRegister(t, c, func() *testLogger {
			callCount++
			return &testLogger{Prefix: fmt.Sprintf("v%d", callCount)}
		}, WithLifetime(Transient))
		mustRegister(t, c, newTestOrderService)
		mustBuild(t, c)

		s1, _ := Resolve[*testOrderService](c)
		s2, _ := Resolve[*testOrderService](c)

		if s1 != s2 {
			t.Fatal("singleton should return same instance")
		}
		if s1.Logger.Prefix != "v1" {
			t.Fatalf("singleton should capture first transient, got %q", s1.Logger.Prefix)
		}
	})

	t.Run("unregistered type returns ErrProviderNotFound", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustBuild(t, c)

		_, err := c.Resolve(reflect.TypeOf((*testConfig)(nil)))
		if !errors.Is(err, ErrProviderNotFound) {
			t.Fatalf("expected ErrProviderNotFound, got: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// Resolve — generic helper
// ---------------------------------------------------------------------------

func TestResolveGeneric(t *testing.T) {
	c := New()
	mustRegister(t, c, newTestLogger)
	mustBuild(t, c)

	logger, err := Resolve[*testLogger](c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logger.Prefix != "app" {
		t.Fatalf("expected prefix 'app', got %q", logger.Prefix)
	}
}

// ---------------------------------------------------------------------------
// Resolve — interface types
// ---------------------------------------------------------------------------

func TestResolve_Interface(t *testing.T) {
	c := New()
	mustRegister(t, c, func() testService {
		return &testUserService{Logger: &testLogger{Prefix: "iface"}}
	})
	mustBuild(t, c)

	svc, err := Resolve[testService](c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.Name() != "user" {
		t.Fatalf("expected 'user', got %q", svc.Name())
	}
}

// ---------------------------------------------------------------------------
// ResolveNamed
// ---------------------------------------------------------------------------

func TestResolveNamed(t *testing.T) {
	t.Run("before build returns ErrNotBuilt", func(t *testing.T) {
		c := New()
		mustRegisterNamed(t, c, "log", newTestLogger)

		_, err := c.ResolveNamed("log", reflect.TypeOf((*testLogger)(nil)))
		if !errors.Is(err, ErrNotBuilt) {
			t.Fatalf("expected ErrNotBuilt, got: %v", err)
		}
	})

	t.Run("resolves no-dep named provider", func(t *testing.T) {
		c := New()
		mustRegisterNamed(t, c, "log", newTestLogger)
		mustBuild(t, c)

		val, err := c.ResolveNamed("log", reflect.TypeOf((*testLogger)(nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		logger := val.Interface().(*testLogger)
		if logger.Prefix != "app" {
			t.Fatalf("expected prefix 'app', got %q", logger.Prefix)
		}
	})

	t.Run("unknown name returns ErrProviderNotFound", func(t *testing.T) {
		c := New()
		mustBuild(t, c)

		_, err := c.ResolveNamed("missing", reflect.TypeOf((*testLogger)(nil)))
		if !errors.Is(err, ErrProviderNotFound) {
			t.Fatalf("expected ErrProviderNotFound, got: %v", err)
		}
	})

	t.Run("named provider with dependencies", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustRegisterNamed(t, c, "order", newTestOrderService)
		mustBuild(t, c)

		svc, err := ResolveNamed[*testOrderService](c, "order")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if svc.Logger == nil {
			t.Fatal("named provider dependency not resolved")
		}
	})

	t.Run("named constructor error is propagated", func(t *testing.T) {
		c := New()
		mustRegisterNamed(t, c, "bad", func() (*testConfig, error) {
			return nil, errors.New("init failed")
		})
		mustBuild(t, c)

		_, err := ResolveNamed[*testConfig](c, "bad")
		if err == nil {
			t.Fatal("expected error from constructor")
		}
		if !strings.Contains(err.Error(), "init failed") {
			t.Fatalf("expected 'init failed' in error, got: %v", err)
		}
	})

	t.Run("type mismatch returns error", func(t *testing.T) {
		c := New()
		mustRegisterNamed(t, c, "log", newTestLogger)
		mustBuild(t, c)

		_, err := c.ResolveNamed("log", reflect.TypeOf((*testConfig)(nil)))
		if err == nil {
			t.Fatal("expected type-mismatch error")
		}
	})

	t.Run("multiple implementations via named providers", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustRegisterNamed(t, c, "user-svc", func(l *testLogger) testService {
			return &testUserService{Logger: l}
		})
		mustRegisterNamed(t, c, "order-svc", func(l *testLogger) testService {
			return &testOrderService{Logger: l}
		})
		mustBuild(t, c)

		userSvc, err := ResolveNamed[testService](c, "user-svc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if userSvc.Name() != "user" {
			t.Fatalf("expected 'user', got %q", userSvc.Name())
		}

		orderSvc, err := ResolveNamed[testService](c, "order-svc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if orderSvc.Name() != "order" {
			t.Fatalf("expected 'order', got %q", orderSvc.Name())
		}
	})

	t.Run("named provider creates new instance each call", func(t *testing.T) {
		c := New()
		mustRegisterNamed(t, c, "log", newTestLogger)
		mustBuild(t, c)

		v1, _ := c.ResolveNamed("log", reflect.TypeOf((*testLogger)(nil)))
		v2, _ := c.ResolveNamed("log", reflect.TypeOf((*testLogger)(nil)))

		if v1.Pointer() == v2.Pointer() {
			t.Fatal("named provider should create a new instance each call")
		}
	})

	t.Run("named provider shares singleton deps", func(t *testing.T) {
		c := New()
		mustRegister(t, c, newTestLogger)
		mustRegisterNamed(t, c, "o1", newTestOrderService)
		mustRegisterNamed(t, c, "o2", newTestOrderService)
		mustBuild(t, c)

		o1, _ := ResolveNamed[*testOrderService](c, "o1")
		o2, _ := ResolveNamed[*testOrderService](c, "o2")

		if o1.Logger != o2.Logger {
			t.Fatal("named providers should share singleton dependencies")
		}
	})
}

func TestResolveNamedGeneric(t *testing.T) {
	c := New()
	mustRegisterNamed(t, c, "log", func() *testLogger { return &testLogger{Prefix: "named"} })
	mustBuild(t, c)

	logger, err := ResolveNamed[*testLogger](c, "log")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logger.Prefix != "named" {
		t.Fatalf("expected prefix 'named', got %q", logger.Prefix)
	}
}

// ---------------------------------------------------------------------------
// Concurrency
// ---------------------------------------------------------------------------

func TestResolve_Concurrent(t *testing.T) {
	c := New()
	mustRegister(t, c, newTestLogger)
	mustRegister(t, c, newTestConfig)
	mustRegister(t, c, newTestDatabase)
	mustRegister(t, c, newTestOrderService, WithLifetime(Transient))
	mustBuild(t, c)

	const goroutines = 100
	var wg sync.WaitGroup
	errs := make(chan error, goroutines*2)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			logger, err := Resolve[*testLogger](c)
			if err != nil {
				errs <- fmt.Errorf("Logger: %w", err)
				return
			}
			if logger.Prefix != "app" {
				errs <- fmt.Errorf("Logger.Prefix = %q", logger.Prefix)
				return
			}

			svc, err := Resolve[*testOrderService](c)
			if err != nil {
				errs <- fmt.Errorf("OrderService: %w", err)
				return
			}
			if svc.Logger == nil {
				errs <- fmt.Errorf("OrderService.Logger is nil")
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent error: %v", err)
	}
}

func TestResolveNamed_Concurrent(t *testing.T) {
	c := New()
	mustRegister(t, c, newTestLogger)
	mustRegisterNamed(t, c, "order", newTestOrderService)
	mustBuild(t, c)

	const goroutines = 100
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			svc, err := ResolveNamed[*testOrderService](c, "order")
			if err != nil {
				errs <- err
				return
			}
			if svc.Logger == nil {
				errs <- fmt.Errorf("Logger is nil")
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestResolve_TransientDependsOnTransient(t *testing.T) {
	c := New()
	mustRegister(t, c, newTestLogger, WithLifetime(Transient))
	mustRegister(t, c, newTestOrderService, WithLifetime(Transient))
	mustBuild(t, c)

	s1, err := Resolve[*testOrderService](c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s2, _ := Resolve[*testOrderService](c)

	if s1 == s2 {
		t.Fatal("expected different OrderService instances")
	}
	if s1.Logger == s2.Logger {
		t.Fatal("expected different Logger instances for transient chain")
	}
}

func TestResolve_TransientConstructorReturningError(t *testing.T) {
	c := New()
	mustRegister(t, c, func() *testLogger { return &testLogger{} }, WithLifetime(Transient))
	mustRegister(t, c, func(l *testLogger) (*testOrderService, error) {
		return nil, errors.New("service init failed")
	}, WithLifetime(Transient))
	mustBuild(t, c)

	_, err := Resolve[*testOrderService](c)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "service init failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolve_ErrorChainShowsResolutionPath(t *testing.T) {
	t.Run("transient chain includes resolution path", func(t *testing.T) {
		c := New()
		mustRegister(t, c, func() (*testConfig, error) {
			return nil, errors.New("connection refused")
		}, WithLifetime(Transient))
		mustRegister(t, c, func(cfg *testConfig) *testDatabase {
			return &testDatabase{Config: cfg}
		}, WithLifetime(Transient))
		mustRegister(t, c, func(db *testDatabase) *testUserRepo {
			return &testUserRepo{DB: db}
		}, WithLifetime(Transient))
		mustBuild(t, c)

		_, err := Resolve[*testUserRepo](c)
		if err == nil {
			t.Fatal("expected error")
		}

		msg := err.Error()
		if !strings.Contains(msg, "resolving *oak.testDatabase") {
			t.Errorf("expected chain to mention *oak.testDatabase, got: %s", msg)
		}
		if !strings.Contains(msg, "resolving *oak.testConfig") {
			t.Errorf("expected chain to mention *oak.testConfig, got: %s", msg)
		}
		if !strings.Contains(msg, "connection refused") {
			t.Errorf("expected root cause 'connection refused', got: %s", msg)
		}
	})

	t.Run("singleton with failing transient dep shows chain", func(t *testing.T) {
		c := New()
		mustRegister(t, c, func() (*testLogger, error) {
			return nil, errors.New("logger init failed")
		}, WithLifetime(Transient))
		mustRegister(t, c, newTestOrderService) // singleton depends on transient *testLogger

		err := c.Build()
		if err == nil {
			t.Fatal("expected error")
		}

		msg := err.Error()
		if !strings.Contains(msg, "resolving *oak.testLogger") {
			t.Errorf("expected chain to mention *oak.testLogger, got: %s", msg)
		}
		if !strings.Contains(msg, "logger init failed") {
			t.Errorf("expected root cause 'logger init failed', got: %s", msg)
		}
	})
}

func TestResolve_ZeroArgConstructor(t *testing.T) {
	c := New()
	mustRegister(t, c, func() int { return 42 })
	mustBuild(t, c)

	val, err := Resolve[int](c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 42 {
		t.Fatalf("expected 42, got %d", val)
	}
}

func TestResolve_ValueType(t *testing.T) {
	type settings struct {
		Debug bool
		Port  int
	}

	c := New()
	mustRegister(t, c, func() settings {
		return settings{Debug: true, Port: 8080}
	})
	mustBuild(t, c)

	s, err := Resolve[settings](c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.Debug || s.Port != 8080 {
		t.Fatalf("unexpected settings: %+v", s)
	}
}

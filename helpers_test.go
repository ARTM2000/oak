package oak

import (
	"errors"
	"testing"
)

// Shared test types and constructors used across test files.

// mustRegister calls t.Fatal if registration fails.
func mustRegister(t *testing.T, c Container, constructor interface{}, opts ...Option) {
	t.Helper()
	if err := c.Register(constructor, opts...); err != nil {
		t.Fatalf("Register: %v", err)
	}
}

// mustRegisterNamed calls t.Fatal if named registration fails.
func mustRegisterNamed(t *testing.T, c Container, name string, constructor interface{}, opts ...Option) {
	t.Helper()
	if err := c.RegisterNamed(name, constructor, opts...); err != nil {
		t.Fatalf("RegisterNamed(%q): %v", name, err)
	}
}

// mustBuild calls t.Fatal if build fails.
func mustBuild(t *testing.T, c Container) {
	t.Helper()
	if err := c.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}
}

type testLogger struct{ Prefix string }
type testConfig struct{ DSN string }

type testDatabase struct {
	Config *testConfig
	Logger *testLogger
}

type testUserRepo struct {
	DB     *testDatabase
	Logger *testLogger
}

type testService interface {
	Name() string
}

type testUserService struct {
	Repo   *testUserRepo
	Logger *testLogger
}

func (s *testUserService) Name() string { return "user" }

type testOrderService struct{ Logger *testLogger }

func (s *testOrderService) Name() string { return "order" }

type testCircA struct{ B *testCircB }
type testCircB struct{ C *testCircC }
type testCircC struct{ A *testCircA }

func newTestLogger() *testLogger           { return &testLogger{Prefix: "app"} }
func newTestConfig() *testConfig           { return &testConfig{DSN: "postgres://localhost"} }
func newTestCircA(b *testCircB) *testCircA { return &testCircA{B: b} }
func newTestCircB(c *testCircC) *testCircB { return &testCircB{C: c} }
func newTestCircC(a *testCircA) *testCircC { return &testCircC{A: a} }

func newTestDatabase(cfg *testConfig, log *testLogger) *testDatabase {
	return &testDatabase{Config: cfg, Logger: log}
}

func newTestUserRepo(db *testDatabase, log *testLogger) *testUserRepo {
	return &testUserRepo{DB: db, Logger: log}
}

func newTestUserService(repo *testUserRepo, log *testLogger) *testUserService {
	return &testUserService{Repo: repo, Logger: log}
}

func newTestOrderService(log *testLogger) *testOrderService {
	return &testOrderService{Logger: log}
}

// testClosable is a singleton that implements io.Closer for shutdown tests.
type testClosable struct {
	Name   string
	Closed bool
	Order  *[]string // shared slice to record close order
}

func (c *testClosable) Close() error {
	c.Closed = true
	if c.Order != nil {
		*c.Order = append(*c.Order, c.Name)
	}
	return nil
}

// testFailCloser implements io.Closer but returns an error.
type testFailCloser struct{}

func (f *testFailCloser) Close() error {
	return errors.New("close failed")
}

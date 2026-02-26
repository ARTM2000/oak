package oak

// Shared test types and constructors used across test files.

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

func newTestLogger() *testLogger                                      { return &testLogger{Prefix: "app"} }
func newTestConfig() *testConfig                                      { return &testConfig{DSN: "postgres://localhost"} }
func newTestDatabase(cfg *testConfig, log *testLogger) *testDatabase   { return &testDatabase{Config: cfg, Logger: log} }
func newTestUserRepo(db *testDatabase, log *testLogger) *testUserRepo  { return &testUserRepo{DB: db, Logger: log} }
func newTestUserService(repo *testUserRepo, log *testLogger) *testUserService {
	return &testUserService{Repo: repo, Logger: log}
}
func newTestOrderService(log *testLogger) *testOrderService { return &testOrderService{Logger: log} }

func newTestCircA(b *testCircB) *testCircA { return &testCircA{B: b} }
func newTestCircB(c *testCircC) *testCircB { return &testCircB{C: c} }
func newTestCircC(a *testCircA) *testCircC { return &testCircC{A: a} }

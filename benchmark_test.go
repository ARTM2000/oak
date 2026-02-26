package oak

import "testing"

func BenchmarkRegister(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := New()
		_ = c.Register(newTestLogger)
		_ = c.Register(newTestConfig)
		_ = c.Register(newTestDatabase)
	}
}

func BenchmarkBuild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := New()
		_ = c.Register(newTestLogger)
		_ = c.Register(newTestConfig)
		_ = c.Register(newTestDatabase)
		_ = c.Register(newTestUserRepo)
		_ = c.Register(newTestUserService)
		_ = c.Build()
	}
}

func BenchmarkResolve_Singleton(b *testing.B) {
	c := New()
	_ = c.Register(newTestLogger)
	_ = c.Register(newTestConfig)
	_ = c.Register(newTestDatabase)
	_ = c.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Resolve[*testDatabase](c)
	}
}

func BenchmarkResolve_Transient(b *testing.B) {
	c := New()
	_ = c.Register(newTestLogger)
	_ = c.Register(func(l *testLogger) *testOrderService {
		return &testOrderService{Logger: l}
	}, WithLifetime(Transient))
	_ = c.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Resolve[*testOrderService](c)
	}
}

func BenchmarkResolveNamed(b *testing.B) {
	c := New()
	_ = c.Register(newTestLogger)
	_ = c.RegisterNamed("order", newTestOrderService)
	_ = c.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ResolveNamed[*testOrderService](c, "order")
	}
}

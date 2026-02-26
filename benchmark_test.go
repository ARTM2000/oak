package oak

import "testing"

func BenchmarkRegister(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := New()
		c.Register(newTestLogger)
		c.Register(newTestConfig)
		c.Register(newTestDatabase)
	}
}

func BenchmarkBuild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := New()
		c.Register(newTestLogger)
		c.Register(newTestConfig)
		c.Register(newTestDatabase)
		c.Register(newTestUserRepo)
		c.Register(newTestUserService)
		c.Build()
	}
}

func BenchmarkResolve_Singleton(b *testing.B) {
	c := New()
	c.Register(newTestLogger)
	c.Register(newTestConfig)
	c.Register(newTestDatabase)
	c.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Resolve[*testDatabase](c)
	}
}

func BenchmarkResolve_Transient(b *testing.B) {
	c := New()
	c.Register(newTestLogger)
	c.Register(func(l *testLogger) *testOrderService {
		return &testOrderService{Logger: l}
	}, WithLifetime(Transient))
	c.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Resolve[*testOrderService](c)
	}
}

func BenchmarkResolveNamed(b *testing.B) {
	c := New()
	c.Register(newTestLogger)
	c.RegisterNamed("order", newTestOrderService)
	c.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ResolveNamed[*testOrderService](c, "order")
	}
}

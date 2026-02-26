package oak_test

import (
	"fmt"

	"github.com/ARTM2000/oak"
)

// Types used in examples only.
type Logger struct{ Prefix string }
type Config struct{ DSN string }
type Database struct {
	Config *Config
	Logger *Logger
}

type Greeter interface {
	Greet() string
}
type englishGreeter struct{}

func (g *englishGreeter) Greet() string { return "hello" }

type spanishGreeter struct{}

func (g *spanishGreeter) Greet() string { return "hola" }

func ExampleNew() {
	c := oak.New()

	_ = c.Register(func() *Logger { return &Logger{Prefix: "app"} })
	if err := c.Build(); err != nil {
		panic(err)
	}

	logger, _ := oak.Resolve[*Logger](c)
	fmt.Println(logger.Prefix)
	// Output: app
}

func ExampleWithLifetime() {
	c := oak.New()
	_ = c.Register(
		func() *Logger { return &Logger{Prefix: "app"} },
		oak.WithLifetime(oak.Transient),
	)
	_ = c.Build()

	l1, _ := oak.Resolve[*Logger](c)
	l2, _ := oak.Resolve[*Logger](c)
	fmt.Println(l1 == l2)
	// Output: false
}

func ExampleResolve() {
	c := oak.New()
	_ = c.Register(func() *Config { return &Config{DSN: "postgres://localhost"} })
	_ = c.Register(func() *Logger { return &Logger{Prefix: "app"} })
	_ = c.Register(func(cfg *Config, log *Logger) *Database {
		return &Database{Config: cfg, Logger: log}
	})
	_ = c.Build()

	db, err := oak.Resolve[*Database](c)
	if err != nil {
		panic(err)
	}
	fmt.Println(db.Config.DSN)
	fmt.Println(db.Logger.Prefix)
	// Output:
	// postgres://localhost
	// app
}

func ExampleContainer_RegisterNamed() {
	c := oak.New()
	_ = c.RegisterNamed("dev", func() *Config { return &Config{DSN: "localhost"} })
	_ = c.RegisterNamed("prod", func() *Config { return &Config{DSN: "prod-host"} })
	_ = c.Build()

	dev, _ := oak.ResolveNamed[*Config](c, "dev")
	prod, _ := oak.ResolveNamed[*Config](c, "prod")
	fmt.Println(dev.DSN)
	fmt.Println(prod.DSN)
	// Output:
	// localhost
	// prod-host
}

func ExampleResolveNamed() {
	c := oak.New()
	_ = c.RegisterNamed("en", func() Greeter { return &englishGreeter{} })
	_ = c.RegisterNamed("es", func() Greeter { return &spanishGreeter{} })
	_ = c.Build()

	en, _ := oak.ResolveNamed[Greeter](c, "en")
	es, _ := oak.ResolveNamed[Greeter](c, "es")
	fmt.Println(en.Greet())
	fmt.Println(es.Greet())
	// Output:
	// hello
	// hola
}

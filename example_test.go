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

	c.Register(func() *Logger { return &Logger{Prefix: "app"} })
	if err := c.Build(); err != nil {
		panic(err)
	}

	logger, _ := oak.Resolve[*Logger](c)
	fmt.Println(logger.Prefix)
	// Output: app
}

func ExampleWithLifetime() {
	c := oak.New()
	c.Register(
		func() *Logger { return &Logger{Prefix: "app"} },
		oak.WithLifetime(oak.Transient),
	)
	c.Build()

	l1, _ := oak.Resolve[*Logger](c)
	l2, _ := oak.Resolve[*Logger](c)
	fmt.Println(l1 == l2)
	// Output: false
}

func ExampleResolve() {
	c := oak.New()
	c.Register(func() *Config { return &Config{DSN: "postgres://localhost"} })
	c.Register(func() *Logger { return &Logger{Prefix: "app"} })
	c.Register(func(cfg *Config, log *Logger) *Database {
		return &Database{Config: cfg, Logger: log}
	})
	c.Build()

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
	c.RegisterNamed("dev", func() *Config { return &Config{DSN: "localhost"} })
	c.RegisterNamed("prod", func() *Config { return &Config{DSN: "prod-host"} })
	c.Build()

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
	c.RegisterNamed("en", func() Greeter { return &englishGreeter{} })
	c.RegisterNamed("es", func() Greeter { return &spanishGreeter{} })
	c.Build()

	en, _ := oak.ResolveNamed[Greeter](c, "en")
	es, _ := oak.ResolveNamed[Greeter](c, "es")
	fmt.Println(en.Greet())
	fmt.Println(es.Greet())
	// Output:
	// hello
	// hola
}

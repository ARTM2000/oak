// Command userapp demonstrates how to wire a small layered application with
// oak, including graceful shutdown. Run it with:
//
//	cd _examples/userapp && go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ARTM2000/oak"
)

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

type Config struct {
	DatabaseURL string
	LogLevel    string
}

type Logger struct {
	Level string
}

func (l *Logger) Info(msg string) {
	fmt.Printf("[%s] %s\n", l.Level, msg)
}

// Database implements io.Closer — oak will call Close() on shutdown
// automatically, in the correct order.
type Database struct {
	URL    string
	Logger *Logger
}

func (db *Database) Query(q string) string {
	db.Logger.Info("query: " + q)
	return "row-result"
}

func (db *Database) Close() error {
	db.Logger.Info("database connection closed")
	return nil
}

type UserRepository struct {
	DB *Database
}

func (r *UserRepository) FindByID(id int) string {
	return r.DB.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %d", id))
}

type UserService struct {
	Repo   *UserRepository
	Logger *Logger
}

func (s *UserService) GetUser(id int) string {
	s.Logger.Info(fmt.Sprintf("looking up user %d", id))
	return s.Repo.FindByID(id)
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

func NewConfig() *Config {
	return &Config{
		DatabaseURL: env("DATABASE_URL", "postgres://localhost:5432/app"),
		LogLevel:    env("LOG_LEVEL", "info"),
	}
}

func NewLogger(cfg *Config) *Logger {
	return &Logger{Level: cfg.LogLevel}
}

func NewDatabase(cfg *Config, l *Logger) *Database {
	return &Database{URL: cfg.DatabaseURL, Logger: l}
}

func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{DB: db}
}

func NewUserService(repo *UserRepository, l *Logger) *UserService {
	return &UserService{Repo: repo, Logger: l}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	c := oak.New()

	// Registration order does not matter.
	c.Register(NewConfig)
	c.Register(NewLogger)
	c.Register(NewDatabase)
	c.Register(NewUserRepository)
	c.Register(NewUserService)

	if err := c.Build(); err != nil {
		log.Fatal(err)
	}

	svc, err := oak.Resolve[*UserService](c)
	if err != nil {
		log.Fatal(err)
	}

	result := svc.GetUser(42)
	fmt.Println("result:", result)

	// Graceful shutdown: closes Database (and any other io.Closer
	// singletons) in reverse dependency order.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Shutdown(ctx); err != nil {
		log.Fatal("shutdown error:", err)
	}
}

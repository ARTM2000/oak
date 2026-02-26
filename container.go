package oak

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
)

// Container defines the interface for the dependency injection container.
// Use [New] to create an instance.
type Container interface {
	// Register adds a constructor to the container. The constructor must be a
	// function with the signature func(deps...) T or func(deps...) (T, error).
	// Dependencies are expressed as function parameters and resolved by type.
	Register(constructor interface{}, opts ...Option) error

	// RegisterNamed adds a named constructor. Named providers live in a
	// separate namespace and are resolved via [Container.ResolveNamed] or
	// the generic [ResolveNamed] helper.
	RegisterNamed(name string, constructor interface{}, opts ...Option) error

	// Build validates the full dependency graph — detecting missing providers
	// and circular dependencies — and eagerly instantiates all [Singleton]
	// providers. After Build succeeds the container is immutable; no further
	// registrations are accepted.
	Build() error

	// Resolve returns the value for the given type. For [Singleton] providers
	// the cached instance is returned; for [Transient] providers a new
	// instance is constructed on each call. Prefer the generic [Resolve]
	// helper over calling this method directly.
	Resolve(t reflect.Type) (reflect.Value, error)

	// ResolveNamed returns the value for the named provider. The requested
	// type t must be assignable from the provider's return type. Prefer the
	// generic [ResolveNamed] helper over calling this method directly.
	ResolveNamed(name string, t reflect.Type) (reflect.Value, error)

	// Shutdown gracefully closes all singleton providers that implement
	// [io.Closer], in reverse dependency order (dependents are closed before
	// their dependencies). The context controls the overall deadline; if it
	// expires, remaining closers are skipped and the context error is
	// included in the result.
	//
	// Shutdown is safe to call multiple times; subsequent calls return
	// [ErrAlreadyShutdown]. It is the caller's responsibility to stop
	// calling [Container.Resolve] before or during shutdown.
	Shutdown(ctx context.Context) error
}

type container struct {
	mu sync.RWMutex

	providers  map[reflect.Type]provider
	named      map[string]provider
	singletons map[reflect.Type]reflect.Value

	// closers holds singletons that implement io.Closer, recorded in
	// dependency order during Build. Shutdown iterates them in reverse.
	closers []io.Closer

	built    bool
	shutdown bool
}

// New creates an empty [Container] ready for registration.
func New() Container {
	return &container{
		providers:  make(map[reflect.Type]provider),
		named:      make(map[string]provider),
		singletons: make(map[reflect.Type]reflect.Value),
	}
}

func (c *container) Register(constructor interface{}, opts ...Option) error {
	return c.register("", constructor, opts...)
}

func (c *container) RegisterNamed(name string, constructor interface{}, opts ...Option) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	return c.register(name, constructor, opts...)
}

func (c *container) register(name string, constructor interface{}, opts ...Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.built {
		return ErrAlreadyBuilt
	}

	val := reflect.ValueOf(constructor)
	typ := val.Type()

	if typ.Kind() != reflect.Func {
		return errors.New("constructor must be a function")
	}

	if typ.NumOut() == 0 || typ.NumOut() > 2 {
		return errors.New("constructor must return (T) or (T, error)")
	}

	if typ.NumOut() == 2 {
		errType := reflect.TypeOf((*error)(nil)).Elem()
		if !typ.Out(1).Implements(errType) {
			return errors.New("second return value must implement error")
		}
	}

	p := provider{
		constructor: val,
		lifetime:    Singleton,
		name:        name,
		outType:     typ.Out(0),
	}

	for _, opt := range opts {
		opt(&p)
	}

	if name != "" {
		if _, exists := c.named[name]; exists {
			return fmt.Errorf("%w: named %q", ErrDuplicateProvider, name)
		}
		c.named[name] = p
		return nil
	}

	outType := typ.Out(0)
	if _, exists := c.providers[outType]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateProvider, outType)
	}
	c.providers[outType] = p
	return nil
}

// ---------------------------------------------------------------------------
// Build
// ---------------------------------------------------------------------------

type buildState int

const (
	unvisited buildState = iota
	visiting
	visited
)

func (c *container) Build() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.built {
		return ErrAlreadyBuilt
	}

	states := make(map[reflect.Type]buildState)

	for t := range c.providers {
		if err := c.buildResolve(t, states, nil); err != nil {
			return err
		}
	}

	for name, p := range c.named {
		if err := c.validateNamedProvider(name, p); err != nil {
			return err
		}
	}

	c.built = true
	return nil
}

// buildResolve walks the dependency graph depth-first using a local state map
// and stack. Singletons are instantiated and cached; transients are only
// validated.
func (c *container) buildResolve(t reflect.Type, states map[reflect.Type]buildState, stack []reflect.Type) error {
	switch states[t] {
	case visiting:
		return c.circularError(t, stack)
	case visited:
		return nil
	}

	p, ok := c.providers[t]
	if !ok {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, t)
	}

	states[t] = visiting
	stack = append(stack, t)

	fnType := p.constructor.Type()
	for i := 0; i < fnType.NumIn(); i++ {
		if err := c.buildResolve(fnType.In(i), states, stack); err != nil {
			return err
		}
	}

	if p.lifetime == Singleton {
		instance, err := c.construct(p)
		if err != nil {
			return fmt.Errorf("constructing %s: %w", t, err)
		}
		c.singletons[t] = instance

		if closer, ok := instance.Interface().(io.Closer); ok {
			c.closers = append(c.closers, closer)
		}
	}

	states[t] = visited
	return nil
}

func (c *container) validateNamedProvider(name string, p provider) error {
	fnType := p.constructor.Type()
	for i := 0; i < fnType.NumIn(); i++ {
		depType := fnType.In(i)
		if _, ok := c.providers[depType]; !ok {
			return fmt.Errorf("named provider %q: %w: %s", name, ErrProviderNotFound, depType)
		}
	}
	return nil
}

func (c *container) circularError(t reflect.Type, stack []reflect.Type) error {
	chain := make([]string, len(stack)+1)
	for i, s := range stack {
		chain[i] = s.String()
	}
	chain[len(stack)] = t.String()

	return fmt.Errorf("%w: %s", ErrCircularDependency, strings.Join(chain, " -> "))
}

// ---------------------------------------------------------------------------
// Shutdown
// ---------------------------------------------------------------------------

func (c *container) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.built {
		return ErrNotBuilt
	}

	if c.shutdown {
		return ErrAlreadyShutdown
	}

	c.shutdown = true

	var errs []error
	for i := len(c.closers) - 1; i >= 0; i-- {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			break
		}
		if err := c.closers[i].Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

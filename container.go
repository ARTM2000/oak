package oak

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var (
	ErrNotBuilt           = errors.New("container not built")
	ErrProviderNotFound   = errors.New("provider not found")
	ErrCircularDependency = errors.New("circular dependency detected")
)

type Lifetime int

const (
	Singleton Lifetime = iota
	Transient
)

type provider struct {
	constructor reflect.Value
	lifetime    Lifetime
	name        string
}

type buildState int

const (
	notVisited buildState = iota
	resolving
	resolved
)

type Container interface {
	Register(constructor interface{}, opts ...Option) error
	RegisterNamed(name string, constructor interface{}, opts ...Option) error
	Build() error
	Resolve(t reflect.Type) (reflect.Value, error)
	ResolveNamed(name string, t reflect.Type) (reflect.Value, error)
}

type container struct {
	mu sync.RWMutex

	providers map[reflect.Type][]provider
	named     map[string]provider

	instances map[reflect.Type]reflect.Value
	states    map[reflect.Type]buildState
	stack     []reflect.Type

	built bool
}

type Option func(*provider)

func WithLifetime(l Lifetime) Option {
	return func(p *provider) {
		p.lifetime = l
	}
}

func New() Container {
	return &container{
		providers: make(map[reflect.Type][]provider),
		named:     make(map[string]provider),
		instances: make(map[reflect.Type]reflect.Value),
		states:    make(map[reflect.Type]buildState),
	}
}

func (c *container) Register(constructor interface{}, opts ...Option) error {
	return c.registerInternal("", constructor, opts...)
}

func (c *container) RegisterNamed(name string, constructor interface{}, opts ...Option) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	return c.registerInternal(name, constructor, opts...)
}

func (c *container) registerInternal(name string, constructor interface{}, opts ...Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.built {
		return errors.New("cannot register after build")
	}

	val := reflect.ValueOf(constructor)
	typ := val.Type()

	if typ.Kind() != reflect.Func {
		return errors.New("constructor must be a function")
	}

	if typ.NumOut() == 0 || typ.NumOut() > 2 {
		return errors.New("constructor must return T or (T, error)")
	}

	if typ.NumOut() == 2 {
		if !typ.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return errors.New("second return value must be error")
		}
	}

	p := provider{
		constructor: val,
		lifetime:    Singleton,
		name:        name,
	}

	for _, opt := range opts {
		opt(&p)
	}

	outType := typ.Out(0)

	if name != "" {
		c.named[name] = p
		return nil
	}

	c.providers[outType] = append(c.providers[outType], p)
	return nil
}

func (c *container) Build() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for typ := range c.providers {
		if _, err := c.resolve(typ); err != nil {
			return err
		}
	}

	c.built = true
	return nil
}

func (c *container) resolve(t reflect.Type) (reflect.Value, error) {

	switch c.states[t] {
	case resolving:
		return reflect.Value{}, c.circularError(t)
	case resolved:
		return c.instances[t], nil
	}

	providers := c.providers[t]
	if len(providers) == 0 {
		return reflect.Value{}, fmt.Errorf("%w: %s", ErrProviderNotFound, t)
	}

	if len(providers) > 1 {
		return reflect.Value{}, fmt.Errorf("multiple providers found for %s (use named registration)", t)
	}

	p := providers[0]

	if p.lifetime == Singleton {
		if inst, ok := c.instances[t]; ok {
			return inst, nil
		}
	}

	c.states[t] = resolving
	c.stack = append(c.stack, t)

	args, err := c.resolveDependencies(p.constructor.Type())
	if err != nil {
		return reflect.Value{}, err
	}

	results := p.constructor.Call(args)

	if len(results) == 2 && !results[1].IsNil() {
		return reflect.Value{}, results[1].Interface().(error)
	}

	instance := results[0]

	if p.lifetime == Singleton {
		c.instances[t] = instance
		c.states[t] = resolved
	}

	c.stack = c.stack[:len(c.stack)-1]

	return instance, nil
}

func (c *container) resolveDependencies(fnType reflect.Type) ([]reflect.Value, error) {
	args := make([]reflect.Value, fnType.NumIn())

	for i := 0; i < fnType.NumIn(); i++ {
		depType := fnType.In(i)

		dep, err := c.resolve(depType)
		if err != nil {
			return nil, err
		}
		args[i] = dep
	}

	return args, nil
}

func (c *container) circularError(t reflect.Type) error {
	var chain []string
	for _, s := range c.stack {
		chain = append(chain, s.String())
	}
	chain = append(chain, t.String())

	return fmt.Errorf("%w: %s",
		ErrCircularDependency,
		strings.Join(chain, " -> "),
	)
}

func (c *container) Resolve(t reflect.Type) (reflect.Value, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.built {
		return reflect.Value{}, ErrNotBuilt
	}

	val, ok := c.instances[t]
	if !ok {
		return reflect.Value{}, fmt.Errorf("%w: %s", ErrProviderNotFound, t)
	}

	return val, nil
}

func (c *container) ResolveNamed(name string, t reflect.Type) (reflect.Value, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.built {
		return reflect.Value{}, ErrNotBuilt
	}

	p, ok := c.named[name]
	if !ok {
		return reflect.Value{}, fmt.Errorf("named provider %s not found", name)
	}

	args, err := c.resolveDependencies(p.constructor.Type())
	if err != nil {
		return reflect.Value{}, err
	}

	results := p.constructor.Call(args)
	return results[0], nil
}

func Resolve[T any](c Container) (T, error) {
	var zero T
	t := reflect.TypeOf((*T)(nil)).Elem()

	val, err := c.Resolve(t)
	if err != nil {
		return zero, err
	}

	return val.Interface().(T), nil
}

func ResolveNamed[T any](c Container, name string) (T, error) {
	var zero T
	t := reflect.TypeOf((*T)(nil)).Elem()

	val, err := c.ResolveNamed(name, t)
	if err != nil {
		return zero, err
	}

	return val.Interface().(T), nil
}

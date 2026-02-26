package oak

import (
	"fmt"
	"reflect"
)

// ---------------------------------------------------------------------------
// Container methods
// ---------------------------------------------------------------------------

func (c *container) Resolve(t reflect.Type) (reflect.Value, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.built {
		return reflect.Value{}, ErrNotBuilt
	}

	if inst, ok := c.singletons[t]; ok {
		return inst, nil
	}

	p, ok := c.providers[t]
	if !ok {
		return reflect.Value{}, fmt.Errorf("%w: %s", ErrProviderNotFound, t)
	}

	return c.construct(p)
}

func (c *container) ResolveNamed(name string, t reflect.Type) (reflect.Value, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.built {
		return reflect.Value{}, ErrNotBuilt
	}

	p, ok := c.named[name]
	if !ok {
		return reflect.Value{}, fmt.Errorf("%w: named %q", ErrProviderNotFound, name)
	}

	if !p.outType.AssignableTo(t) {
		return reflect.Value{}, fmt.Errorf("named provider %q returns %s, not assignable to %s", name, p.outType, t)
	}

	return c.construct(p)
}

// ---------------------------------------------------------------------------
// Generic helpers
// ---------------------------------------------------------------------------

// Resolve is a generic helper that resolves a typed provider from the
// container. It is the recommended way to retrieve values:
//
//	db, err := oak.Resolve[*Database](c)
func Resolve[T any](c Container) (T, error) {
	var zero T
	t := reflect.TypeOf((*T)(nil)).Elem()

	val, err := c.Resolve(t)
	if err != nil {
		return zero, err
	}

	out, ok := val.Interface().(T)
	if !ok {
		return zero, fmt.Errorf("cannot convert %s to %s", val.Type(), t)
	}

	return out, nil
}

// ResolveNamed is a generic helper that resolves a named provider from the
// container:
//
//	db, err := oak.ResolveNamed[*Database](c, "primary")
func ResolveNamed[T any](c Container, name string) (T, error) {
	var zero T
	t := reflect.TypeOf((*T)(nil)).Elem()

	val, err := c.ResolveNamed(name, t)
	if err != nil {
		return zero, err
	}

	out, ok := val.Interface().(T)
	if !ok {
		return zero, fmt.Errorf("named %q: cannot convert %s to %s", name, val.Type(), t)
	}

	return out, nil
}

// ---------------------------------------------------------------------------
// Internal
// ---------------------------------------------------------------------------

// construct creates a new instance by resolving all dependencies. Singleton
// deps come from the cache; transient deps are recursively constructed. This
// method only reads c.singletons and c.providers, so it is safe under a
// read-lock after Build.
func (c *container) construct(p provider) (reflect.Value, error) {
	fnType := p.constructor.Type()
	args := make([]reflect.Value, fnType.NumIn())

	for i := 0; i < fnType.NumIn(); i++ {
		depType := fnType.In(i)

		if inst, ok := c.singletons[depType]; ok {
			args[i] = inst
			continue
		}

		depProvider, ok := c.providers[depType]
		if !ok {
			return reflect.Value{}, fmt.Errorf("%w: %s", ErrProviderNotFound, depType)
		}

		inst, err := c.construct(depProvider)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("resolving %s: %w", depType, err)
		}
		args[i] = inst
	}

	results := p.constructor.Call(args)
	if len(results) == 2 && !results[1].IsNil() {
		return reflect.Value{}, results[1].Interface().(error)
	}

	return results[0], nil
}

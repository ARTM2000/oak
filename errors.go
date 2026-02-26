package oak

import "errors"

var (
	// ErrNotBuilt is returned when Resolve is called before Build.
	ErrNotBuilt = errors.New("container not built")

	// ErrAlreadyBuilt is returned when Register or Build is called after the
	// container has already been built.
	ErrAlreadyBuilt = errors.New("container already built")

	// ErrProviderNotFound is returned when no provider is registered for the
	// requested type or name.
	ErrProviderNotFound = errors.New("provider not found")

	// ErrCircularDependency is returned when the dependency graph contains a
	// cycle. The error message includes the full chain.
	ErrCircularDependency = errors.New("circular dependency detected")

	// ErrDuplicateProvider is returned when a provider for the same type or
	// name is registered more than once.
	ErrDuplicateProvider = errors.New("duplicate provider")
)

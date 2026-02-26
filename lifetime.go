package oak

// Lifetime controls how many instances of a provider the container creates.
type Lifetime int

const (
	// Singleton is the default lifetime. The constructor is called once during
	// [Container.Build] and the resulting instance is reused for every
	// subsequent [Container.Resolve] call.
	Singleton Lifetime = iota

	// Transient means a new instance is constructed on every
	// [Container.Resolve] call.
	Transient
)

// String returns the human-readable name of the lifetime.
func (l Lifetime) String() string {
	switch l {
	case Singleton:
		return "singleton"
	case Transient:
		return "transient"
	default:
		return "unknown"
	}
}

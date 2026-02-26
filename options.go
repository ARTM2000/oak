package oak

import "reflect"

// provider holds the metadata for a single registered constructor.
type provider struct {
	constructor reflect.Value
	lifetime    Lifetime
	name        string
	outType     reflect.Type
}

// Option configures a provider during registration.
type Option func(*provider)

// WithLifetime sets the [Lifetime] of the provider. The default is
// [Singleton].
func WithLifetime(l Lifetime) Option {
	return func(p *provider) {
		p.lifetime = l
	}
}

package provider

import (
	"fmt"

	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
)

// Registry manages payment provider instances by name.
type Registry struct {
	providers map[string]domain.PaymentProvider
}

// NewRegistry creates a provider registry and registers the given providers.
func NewRegistry(providers ...domain.PaymentProvider) *Registry {
	r := &Registry{providers: make(map[string]domain.PaymentProvider)}
	for _, p := range providers {
		r.providers[p.Name()] = p
	}
	return r
}

// Get returns the provider for the given name.
func (r *Registry) Get(name string) (domain.PaymentProvider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
	return p, nil
}

// All returns all registered providers as a map (for backwards compatibility).
func (r *Registry) All() map[string]domain.PaymentProvider {
	return r.providers
}

// Register adds a provider to the registry.
func (r *Registry) Register(p domain.PaymentProvider) {
	r.providers[p.Name()] = p
}

// Names returns the names of all registered providers.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

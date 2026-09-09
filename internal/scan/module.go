package scan

import (
	"context"

	"github.com/ashwinADHD/k-cleaner/internal/models"
)

// Module is the extension interface for future scan plugins (Homebrew, dev caches, PKG).
// Phase 1 uses built-in scanners; register modules in Scanner when implementing Phase 3.
type Module interface {
	Name() string
	Description() string
	Scan(ctx context.Context) ([]models.Item, error)
}

// Registry holds optional scan modules. Empty in v1; populated as plugins are added.
type Registry struct {
	modules []Module
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(m Module) {
	r.modules = append(r.modules, m)
}

func (r *Registry) Modules() []Module {
	return r.modules
}

func (r *Registry) ScanAll(ctx context.Context) ([]models.Item, error) {
	var all []models.Item
	for _, m := range r.modules {
		items, err := m.Scan(ctx)
		if err != nil {
			return all, err
		}
		all = append(all, items...)
	}
	return all, nil
}

package agents

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Registry struct {
	mu       sync.RWMutex
	adapters map[string]AgentAdapter
}

// NewRegistry creates a new empty agent registry.
func NewRegistry() *Registry {
	return &Registry{adapters: make(map[string]AgentAdapter)}
}

// Register adds an adapter to the registry under its name and aliases.
func (r *Registry) Register(a AgentAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[strings.ToLower(a.Name())] = a
	for _, alias := range a.Aliases() {
		r.adapters[strings.ToLower(alias)] = a
	}
}

// Get returns the adapter registered with the given name or alias.
func (r *Registry) Get(name string) (AgentAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.adapters[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("unsupported agent: '%s'", name)
	}
	return a, nil
}

// All returns a slice of all unique registered adapters sorted by name.
func (r *Registry) All() []AgentAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := make(map[string]bool)
	var list []AgentAdapter
	for _, a := range r.adapters {
		if !seen[a.Name()] {
			seen[a.Name()] = true
			list = append(list, a)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() < list[j].Name()
	})
	return list
}

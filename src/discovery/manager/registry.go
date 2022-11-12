package manager

import (
	"sync"
)

type Registry struct {
	mapping map[string]Harvester
	mu      sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{ mapping: make(map[string]Harvester) }
}

func (r *Registry) Get(key string) (Harvester, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	harv, ok := r.mapping[key]

	return harv, ok
}

func (r *Registry) Put(key string, harv Harvester) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mapping[key] = harv
}

func (r *Registry) Del(key string) Harvester {
	r.mu.Lock()
	defer r.mu.Unlock()

	harv, ok := r.mapping[key]

	if ok {
		delete(r.mapping, key)
	}

	return harv
}

func (r *Registry) List() []Harvester {
	r.mu.Lock()
	defer r.mu.Unlock()

	var l []Harvester

	for _, harv := range r.mapping {
		l = append(l, harv)
	}

	return l
}
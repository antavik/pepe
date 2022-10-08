package manager

import (
	"sync"

	"github.com/antibantique/pepe/src/source"
)

type Registry struct {
	mapping map[string]*source.S
	mu      sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{mapping: make(map[string]*source.S)}
}

func (r *Registry) Get(key string) (src *source.S, ok bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	src, ok = r.mapping[key]

	return src, ok
}

func (r *Registry) Put(key string, src *source.S) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mapping[key] = src
}

func (r *Registry) Del(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.mapping, key)
}
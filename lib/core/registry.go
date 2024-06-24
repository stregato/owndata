package core

import "sync"

type Registry[T any] struct {
	handles   map[uint64]T
	count     uint64
	countSync sync.Mutex
}

func (h *Registry[T]) Add(v T) uint64 {
	h.countSync.Lock()
	defer h.countSync.Unlock()
	h.count++
	if h.handles == nil {
		h.handles = make(map[uint64]T)
	}
	h.handles[h.count] = v
	return h.count
}

func (h *Registry[T]) Get(i uint64) (T, error) {
	v, ok := h.handles[i]
	if !ok {
		return v, Errorf("handle %d not found", i)
	}
	return v, nil
}

func (h *Registry[T]) Remove(i uint64) {
	delete(h.handles, i)
}

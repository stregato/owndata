package core

import "sync"

type Registry[T any] struct {
	handles   map[int]T
	count     int
	countSync sync.Mutex
}

func (h *Registry[T]) Add(v T) int {
	h.countSync.Lock()
	defer h.countSync.Unlock()
	h.count++
	h.handles[h.count] = v
	return h.count
}

func (h *Registry[T]) Get(i int) (T, error) {
	v, ok := h.handles[i]
	if !ok {
		return v, Errorf("handle %d not found", i)
	}
	return v, nil
}

func (h *Registry[T]) Remove(i int) {
	delete(h.handles, i)
}

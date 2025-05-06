package kyro

import (
	"errors"
	"sync"
)

// RoundRobin is a thread-safe wrapper for accessing slice elements
// in a round-robin fashion.
type RoundRobin[T any] struct {
	items []T
	mu    sync.Mutex
	index int
}

// NewRoundRobin creates a new RoundRobin wrapper.
// It returns an error if the input slice is empty.
func NewRoundRobin[T any](items []T) (*RoundRobin[T], error) {
	if len(items) == 0 {
		return nil, errors.New("cannot create RoundRobin with an empty slice")
	}

	itemsCopy := make([]T, len(items))
	copy(itemsCopy, items)

	return &RoundRobin[T]{items: itemsCopy, index: 0}, nil
}

// Next returns the next element from the slice in a round-robin fashion.
// This method is safe for concurrent use by multiple goroutines.
func (rr *RoundRobin[T]) Next() T {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	item := rr.items[rr.index]
	rr.index = (rr.index + 1) % len(rr.items)

	return item
}

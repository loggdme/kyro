package kyro

import "sync"

type SimpleSet[T comparable] struct {
	elements map[T]struct{}
	mu       sync.RWMutex
}

// NewSimpleSet creates a new SimpleSet with an initial capacity.
// The capacity is used to optimize memory allocation but does not limit the number of elements.
// The set can grow beyond the initial capacity as needed.
func NewSimpleSet[T comparable](capacity int) *SimpleSet[T] {
	return &SimpleSet[T]{
		elements: make(map[T]struct{}, capacity),
	}
}

// Add inserts an element into the set. If the element already exists, it will not be added again.
func (s *SimpleSet[T]) Add(value T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.elements[value] = struct{}{}
}

// Contains checks if the set contains the specified element.
func (s *SimpleSet[T]) Contains(value T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.elements[value]
	return exists
}

// Clear removes all elements from the set, effectively resetting it.
func (s *SimpleSet[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.elements = make(map[T]struct{})
}

// AsSlice returns all elements in the set as a slice.
// The order of elements in the slice is not guaranteed to be the same as the order of insertion.
func (s *SimpleSet[T]) AsSlice() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]T, 0, len(s.elements))
	for elem := range s.elements {
		keys = append(keys, elem)
	}

	return keys
}

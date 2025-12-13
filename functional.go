package kyro

func Map[T, V any](ts []T, fn func(val T, index int) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t, i)
	}
	return result
}

func FindFirst[T any](slice []T, predicate func(T) bool) *T {
	for _, item := range slice {
		if predicate(item) {
			return &item
		}
	}
	return nil
}

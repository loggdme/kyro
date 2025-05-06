package kyro

func Map[T, V any](ts []T, fn func(val T, index int) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t, i)
	}
	return result
}

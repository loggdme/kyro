package kyro

import (
	"os"
	"time"
)

func Map[T, V any](ts []T, fn func(val T, index int) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t, i)
	}
	return result
}

// ProgressNotifier is a function type for notifying the progress of the queue processing.
type ProgressNotifier func(curr int, duration time.Duration, itemsPerSecond float64)

// ErrorNotifier is a function type for notifying about errors during processing.
type ErrorNotifier[ITEM any] func(err error, item ITEM)

// ProcessFunc is a function type for processing an item.
type ProcessFunc[ITEM any] func(ITEM) error

func SafeRemoveFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return err
		}
		return err
	}

	return nil
}

package kyro

import (
	"time"
)

// ProgressNotifier is a function type for notifying the progress of the queue processing.
type ProgressNotifier func(curr int, duration time.Duration, itemsPerSecond float64)

// ErrorNotifier is a function type for notifying about errors during processing.
type ErrorNotifier[ITEM any] func(err error, item ITEM)

// ProcessFunc is a function type for processing an item.
type ProcessFunc[ITEM any] func(ITEM) error

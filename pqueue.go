package kyro

import (
	"fmt"
	"sync"
	"time"
)

// ParallelQueue represents a queue for processing items in parallel.
type ParallelQueue[ITEM any] struct {
	items           *[]ITEM
	numberOfWorkers int

	processFunc    ProcessFunc[ITEM]
	processed      int
	processedMutex sync.Mutex

	progressBatch int
	progressFunc  ProgressNotifier

	errorFunc ErrorNotifier[ITEM]
}

// NewParallelQueue creates a new ParallelQueue with the specified number of workers.
func NewParallelQueue[ITEM any](numberOfWorkers int) *ParallelQueue[ITEM] {
	return &ParallelQueue[ITEM]{
		numberOfWorkers: numberOfWorkers,
		progressBatch:   100,
	}
}

// WithItems sets the items to be processed by the queue.
func (c *ParallelQueue[ITEM]) WithItems(items *[]ITEM) *ParallelQueue[ITEM] {
	c.items = items
	return c
}

// OnProcessItem sets the function to be used for processing each item.
func (c *ParallelQueue[ITEM]) OnProcessItem(processFunc ProcessFunc[ITEM]) *ParallelQueue[ITEM] {
	c.processFunc = processFunc
	return c
}

// WithProgressNotifier sets the progress notification function and the batch size.
// batch is the number of items processed before the progress function is called.
func (c *ParallelQueue[ITEM]) WithProgressNotifier(batch int, progressFunc ProgressNotifier) *ParallelQueue[ITEM] {
	c.progressFunc = progressFunc
	c.progressBatch = batch
	return c
}

// WithErrorNotifier sets the error notification function.
// errorFunc is the function to call when an error occurs during processing.
func (c *ParallelQueue[ITEM]) WithErrorNotifier(errorFunc ErrorNotifier[ITEM]) *ParallelQueue[ITEM] {
	c.errorFunc = errorFunc
	return c
}

// Process starts the parallel processing of the enqueued items. It returns a slice of items
// that failed to process and an error if any critical error occurred during setup or processing.
func (c *ParallelQueue[ITEM]) Process() (*[]ITEM, error) {
	var erroredItems []ITEM

	if c.numberOfWorkers <= 0 {
		return &erroredItems, fmt.Errorf("number of workers must be positive")
	}

	if c.items == nil || len(*c.items) == 0 {
		return &erroredItems, fmt.Errorf("items must be non-nil and non-empty")
	}

	if c.processFunc == nil {
		return &erroredItems, fmt.Errorf("process function must be set")
	}

	itemCh := make(chan ITEM, c.numberOfWorkers)

	var wg sync.WaitGroup
	wg.Add(c.numberOfWorkers)

	// errCh is buffered to avoid blocking workers if the errorFunc is slow or the
	// error channel is not consumed quickly enough. The size is set to the total
	// number of items as a safe upper bound.
	errCh := make(chan ITEM, len(*c.items))

	startTime := time.Now()

	// worker is the function executed by each goroutine to process items from the item channel.
	worker := func() {
		defer wg.Done()
		for item := range itemCh {
			if err := c.processFunc(item); err != nil {
				select {
				// Attempt to send the errored item to the error channel.
				case errCh <- item:
					if c.errorFunc != nil {
						c.errorFunc(err, item)
					}
				// If the error channel is full, we report this as an error
				// before attempting to report the original processing error.
				default:
					if c.errorFunc != nil {
						c.errorFunc(fmt.Errorf("error channel is full"), item)
						c.errorFunc(err, item)
					}
				}
			}

			c.processedMutex.Lock()
			c.processed++
			currentProcessed := c.processed
			c.processedMutex.Unlock()

			if c.progressFunc != nil && currentProcessed%c.progressBatch == 0 {
				duration := time.Since(startTime)
				itemsPerSecond := float64(currentProcessed) / duration.Seconds()
				c.progressFunc(currentProcessed, duration, itemsPerSecond)
			}
		}
	}

	// Start the worker goroutines. We use c.numberOfWorkers to determine how many
	// goroutines to start. Each goroutine will process items from the item channel.
	for i := 0; i < c.numberOfWorkers; i++ {
		go worker()
	}

	// Goroutine to send items to the item channel. The channel gets
	// closed when all items have been sent.
	go func() {
		for _, item := range *c.items {
			itemCh <- item
		}
		close(itemCh)
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		erroredItems = append(erroredItems, err)
	}

	if len(erroredItems) > 0 {
		return &erroredItems, fmt.Errorf("encountered %d errors during processing", len(erroredItems))
	}

	return &erroredItems, nil
}

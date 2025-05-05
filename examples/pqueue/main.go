package main

import (
	"fmt"
	"log"
	"time"

	"github.com/loggdme/kyro"
)

func main() {
	var ids []int
	for i := 1; i <= 1000; i++ {
		ids = append(ids, i)
	}

	limiter := kyro.NewRateLimiter(40, 40)
	erroredItems, err := kyro.NewParallelQueue[int](40).
		Enqueue(&ids).
		WithProgressNotifier(100, func(curr int, duration time.Duration, itemsPerSecond float64) {
			log.Printf("Processed %d items in %s (%.2f items/sec)", curr, duration, itemsPerSecond)
		}).
		WithErrorNotifier(func(err error, item int) {
			log.Printf("Error processing item %d: %v", item, err)
		}).
		OnProcessQueue(func(item int) error {
			limiter.Wait()

			if item%420 == 0 {
				return fmt.Errorf("simulated error for item %d", item)
			}

			return nil
		}).
		Done()

	if err != nil {
		fmt.Printf("Error during data processing: %v", err)
		fmt.Println("Errored items:", erroredItems)
	}
}

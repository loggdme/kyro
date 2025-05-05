package pqueue_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/loggdme/kyro/pkg/pqueue"
	"github.com/stretchr/testify/assert"
)

func TestParallelQueue_Done_Success(t *testing.T) {
	q := pqueue.NewParallelQueue[int](3)
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	processedItems := []int{}
	var mu sync.Mutex

	q.Enqueue(&items).
		OnProcessQueue(func(item int) error {
			mu.Lock()
			processedItems = append(processedItems, item)
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			return nil
		})

	erroredItems, err := q.Done()

	assert.NoError(t, err)
	assert.Empty(t, *erroredItems)
	assert.Len(t, processedItems, len(items))

	// Check if all items were processed
	processedMap := make(map[int]bool)
	for _, item := range processedItems {
		processedMap[item] = true
	}

	for _, item := range items {
		assert.True(t, processedMap[item], fmt.Sprintf("Item %d was not processed", item))
	}
}

func TestParallelQueue_Done_WithError(t *testing.T) {
	q := pqueue.NewParallelQueue[int](2)
	items := []int{1, 2, 3, 4, 5}
	expectedError := errors.New("processing error")
	erroredItemsNotifier := []int{}
	var erroredItemsNotifierMu sync.Mutex

	q.Enqueue(&items).
		OnProcessQueue(func(item int) error {
			if item%2 == 0 {
				return expectedError
			}
			time.Sleep(10 * time.Millisecond)
			return nil
		}).
		WithErrorNotifier(func(err error, item int) {
			assert.Equal(t, expectedError, err)
			erroredItemsNotifierMu.Lock()
			erroredItemsNotifier = append(erroredItemsNotifier, item)
			erroredItemsNotifierMu.Unlock()
		})

	resultErroredItems, err := q.Done()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "encountered 2 errors during processing")

	// Items 2 and 4 should error
	assert.Len(t, *resultErroredItems, 2)

	// Check if the correct items errored
	erroredMap := make(map[int]bool)
	for _, item := range *resultErroredItems {
		erroredMap[item] = true
	}
	assert.True(t, erroredMap[2], "Item 2 should be in errored items")
	assert.True(t, erroredMap[4], "Item 4 should be in errored items")

	// Check if the error notifier was called for the correct items
	notifierErroredMap := make(map[int]bool)
	for _, item := range erroredItemsNotifier {
		notifierErroredMap[item] = true
	}
	assert.True(t, notifierErroredMap[2], "Error notifier should have been called for item 2")
	assert.True(t, notifierErroredMap[4], "Error notifier should have been called for item 4")
}

func TestParallelQueue_Done_NoWorkers(t *testing.T) {
	q := pqueue.NewParallelQueue[int](0)
	items := []int{1, 2}
	q.Enqueue(&items).OnProcessQueue(func(item int) error { return nil })

	erroredItems, err := q.Done()
	assert.Error(t, err)
	assert.EqualError(t, err, "number of workers must be positive")
	assert.Empty(t, *erroredItems)
}

func TestParallelQueue_Done_ProgressNotifier(t *testing.T) {
	q := pqueue.NewParallelQueue[int](2)
	items := make([]int, 200)
	for i := range items {
		items[i] = i + 1
	}

	progressNotifications := []int{}
	var progressMu sync.Mutex

	q.Enqueue(&items).
		OnProcessQueue(func(item int) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		}).
		WithProgressNotifier(50, func(curr int, duration time.Duration, itemsPerSecond float64) {
			progressMu.Lock()
			progressNotifications = append(progressNotifications, curr)
			progressMu.Unlock()
			assert.GreaterOrEqual(t, curr, 0)
			assert.Greater(t, duration, time.Duration(0))
			assert.GreaterOrEqual(t, itemsPerSecond, float64(0))
		})

	_, err := q.Done()
	assert.NoError(t, err)

	// Check if progress was notified at the correct batches
	expectedNotifications := map[int]bool{50: false, 100: false, 150: false, 200: false}
	for _, notification := range progressNotifications {
		if _, ok := expectedNotifications[notification]; ok {
			expectedNotifications[notification] = true
		}
	}

	for notificationPoint, found := range expectedNotifications {
		assert.True(t, found, fmt.Sprintf("Progress notification at %d was not received", notificationPoint))
	}

	assert.GreaterOrEqual(t, len(progressNotifications), len(expectedNotifications))
}

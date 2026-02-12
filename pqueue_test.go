package kyro_test

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/loggdme/kyro"
)

func TestParallelQueue_Done_Success(t *testing.T) {
	q := kyro.NewParallelQueue[int](3)
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	processedItems := []int{}
	var mu sync.Mutex

	q.WithItems(&items).
		OnProcessItem(func(item int) error {
			mu.Lock()
			processedItems = append(processedItems, item)
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			return nil
		})

	erroredItems, err := q.Process()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(*erroredItems) != 0 {
		t.Errorf("expected empty errored items, got %v", *erroredItems)
	}
	if len(processedItems) != len(items) {
		t.Errorf("expected %d processed items, got %d", len(items), len(processedItems))
	}

	// Check if all items were processed
	processedMap := make(map[int]bool)
	for _, item := range processedItems {
		processedMap[item] = true
	}

	for _, item := range items {
		if !processedMap[item] {
			t.Errorf("Item %d was not processed", item)
		}
	}
}

func TestParallelQueue_Done_WithError(t *testing.T) {
	q := kyro.NewParallelQueue[int](2)
	items := []int{1, 2, 3, 4, 5}
	expectedError := errors.New("processing error")
	erroredItemsNotifier := []int{}
	var erroredItemsNotifierMu sync.Mutex

	q.WithItems(&items).
		OnProcessItem(func(item int) error {
			if item%2 == 0 {
				return expectedError
			}
			time.Sleep(10 * time.Millisecond)
			return nil
		}).
		WithErrorNotifier(func(err error, item int) {
			if err != expectedError {
				t.Errorf("expected error %v, got %v", expectedError, err)
			}
			erroredItemsNotifierMu.Lock()
			erroredItemsNotifier = append(erroredItemsNotifier, item)
			erroredItemsNotifierMu.Unlock()
		})

	resultErroredItems, err := q.Process()
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "encountered 2 errors during processing") {
		t.Errorf("expected error to contain 'encountered 2 errors during processing', got: %v", err)
	}

	// Items 2 and 4 should error
	if len(*resultErroredItems) != 2 {
		t.Errorf("expected 2 errored items, got %d", len(*resultErroredItems))
	}

	// Check if the correct items errored
	erroredMap := make(map[int]bool)
	for _, item := range *resultErroredItems {
		erroredMap[item] = true
	}
	if !erroredMap[2] {
		t.Error("Item 2 should be in errored items")
	}
	if !erroredMap[4] {
		t.Error("Item 4 should be in errored items")
	}

	// Check if the error notifier was called for the correct items
	notifierErroredMap := make(map[int]bool)
	for _, item := range erroredItemsNotifier {
		notifierErroredMap[item] = true
	}
	if !notifierErroredMap[2] {
		t.Error("Error notifier should have been called for item 2")
	}
	if !notifierErroredMap[4] {
		t.Error("Error notifier should have been called for item 4")
	}
}

func TestParallelQueue_Done_NoWorkers(t *testing.T) {
	q := kyro.NewParallelQueue[int](0)
	items := []int{1, 2}
	q.WithItems(&items).OnProcessItem(func(item int) error { return nil })

	erroredItems, err := q.Process()
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != nil && err.Error() != "number of workers must be positive" {
		t.Errorf("expected error 'number of workers must be positive', got: %v", err)
	}
	if len(*erroredItems) != 0 {
		t.Errorf("expected empty errored items, got %v", *erroredItems)
	}
}

func TestParallelQueue_Done_ProgressNotifier(t *testing.T) {
	q := kyro.NewParallelQueue[int](2)
	items := make([]int, 200)
	for i := range items {
		items[i] = i + 1
	}

	progressNotifications := []int{}
	var progressMu sync.Mutex

	q.WithItems(&items).
		OnProcessItem(func(item int) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		}).
		WithProgressNotifier(50, func(curr int, duration time.Duration, itemsPerSecond float64) {
			progressMu.Lock()
			progressNotifications = append(progressNotifications, curr)
			progressMu.Unlock()
			if curr < 0 {
				t.Errorf("expected curr >= 0, got %d", curr)
			}
			if duration <= 0 {
				t.Errorf("expected duration > 0, got %v", duration)
			}
			if itemsPerSecond < 0 {
				t.Errorf("expected itemsPerSecond >= 0, got %f", itemsPerSecond)
			}
		})

	_, err := q.Process()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check if progress was notified at the correct batches
	expectedNotifications := map[int]bool{50: false, 100: false, 150: false, 200: false}
	for _, notification := range progressNotifications {
		if _, ok := expectedNotifications[notification]; ok {
			expectedNotifications[notification] = true
		}
	}

	for notificationPoint, found := range expectedNotifications {
		if !found {
			t.Errorf("Progress notification at %d was not received", notificationPoint)
		}
	}

	if len(progressNotifications) < len(expectedNotifications) {
		t.Errorf("expected at least %d progress notifications, got %d", len(expectedNotifications), len(progressNotifications))
	}
}

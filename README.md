# 🚀 Kyro

Welcome to the **Kyro** project! This repository contains a collection of utilities and examples for working with pipelines, parallel queues, and rate limiters in Go. It is designed to help you build efficient and scalable data processing workflows.

## 📦 Core Packages

### 1. **Pipeline** (`pkg/pipeline`)

The `pipeline` package provides utilities for creating and executing sequential and parallel workflows.

#### Example:

```go
result, err := pipeline.Execute(
    generateItems,
    pipeline.InSequence(
        stringLength,
        pipeline.InParallel(double, triple),
        add,
    ),
)
```

### 2. **Parallel Queue** (`pkg/pqueue`)

The `pqueue` package enables parallel processing of items with configurable workers, progress notifications, and error handling.

#### Example:

```go
erroredItems, err := pqueue.NewParallelQueue[int](40).
    Enqueue(&ids).
    WithProgressNotifier(100, func(curr int, duration time.Duration, itemsPerSecond float64) {
        log.Printf("Processed %d items in %s (%.2f items/sec)", curr, duration, itemsPerSecond)
    }).
    OnProcessQueue(func(item int) error {
        if item%420 == 0 {
            return fmt.Errorf("simulated error for item %d", item)
        }
        return nil
    }).
    Done()
```

---

### 3. **Rate Limiter** (`pkg/limiter`)

The `limiter` package wraps the `golang.org/x/time/rate` library to provide a simple interface for rate limiting.

#### Example:

```go
limiter := limiter.NewRateLimiter(2, 2)
if err := limiter.Wait(); err != nil {
    log.Fatalf("Rate limiter error: %v", err)
}
```

## 🧪 Examples

You can find various examples demonstrating the usage of the core packages in the `examples` directory. Each example is self-contained and can be run independently.

## 📄 License

This project is licensed under the MIT License. See the `LICENSE` file for details.

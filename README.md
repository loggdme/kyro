# 🚀 Kyro

Welcome to the **Kyro** project! This repository contains a collection of utilities and examples for working with pipelines, parallel queues, and rate limiters in Go. It is designed to help you build efficient and scalable data processing workflows.

## 📦 Core Packages

You can find various examples demonstrating the usage of the core packages in the `examples` directory. Each example is self-contained and can be run independently.

### 1. **Pipeline** (`pkg/pipeline`)

The `pipeline` package provides utilities for creating and executing sequential and parallel workflows.

```go
result, err := pipeline.Execute(
    pipeline.InSequence(
        generateItems,
        stringLength,
        pipeline.InParallel(double, triple),
        add,
    ),
)
```

### 2. **Parallel Queue** (`pkg/pqueue`)

The `pqueue` package enables parallel processing of items with configurable workers, progress notifications, and error handling.

```go
limiter := kyro.NewRateLimiter(40, 40)
erroredItems, err := pqueue.NewParallelQueue[int](40).
    Enqueue(&ids).
    WithProgressNotifier(100, func(curr int, duration time.Duration, itemsPerSecond float64) {
        log.Printf("Processed %d items in %s (%.2f items/sec)", curr, duration, itemsPerSecond)
    }).
    OnProcessQueue(func(item int) error {
        limiter.Wait()
        
        if item%420 == 0 {
            return fmt.Errorf("simulated error for item %d", item)
        }
        
        return nil
    }).
    Done()
```

## 📄 License

This project is licensed under the MIT License. See the `LICENSE` file for details.

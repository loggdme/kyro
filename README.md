<div align="center">

![Loggd Banner](./.github/assets/header.webp)
# Kyro

</div>

<br>

A collection of composable utilities for building efficient data processing pipelines, parallel task execution, and concurrent workflows in Go. Features type-safe generics, fluent APIs, and zero external dependencies (except for rate limiting).

## ⚙️ Prerequisites

Install all required tools at their tested versions using [mise](https://mise.jdx.dev/):

```bash
mise install
```

## 📦 Installation

```bash
go get github.com/loggdme/kyro
```

## 🚀 Features

### Pipeline Execution

Build composable data processing pipelines with type-safe sequential and parallel execution:

```go
generateItems := kyro.AsPipelineGenerator(func() (string, error) {
    return "Hello, Kyro!", nil
})

stringLength := kyro.AsPipelineStep(func(input string, err error) (int, error) {
    return len(input), err
})

double := kyro.AsPipelineStep(func(input int, err error) (int, error) {
    return input * 2, err
})

triple := kyro.AsPipelineStep(func(input int, err error) (int, error) {
    return input * 3, err
})

add := kyro.AsPipelineStep(func(input []any, err error) (int, error) {
    first, second := kyro.AssertIn[int](input[0]), kyro.AssertIn[int](input[1])
    return first + second, err
})

result, err := kyro.Execute(
    kyro.InSequence(
        generateItems,
        stringLength,
        kyro.InParallel(double, triple),
        add,
    ),
)
```

**Key Pipeline Features:**
- Sequential execution with `InSequence`
- Parallel execution with `InParallel`
- Type-safe step composition with generics
- Error propagation and exit-on-error support
- Built-in steps: `RemoveFileStep`, `ExitOnErrorStep`, `TakeFirstStep`, `TakeLastStep`, `TakeSubsetStep`

### Parallel Queue Processing

Process collections concurrently with configurable workers, progress tracking, and error handling:

```go
var ids []int
for i := 1; i <= 1000; i++ {
    ids = append(ids, i)
}

limiter := kyro.NewRateLimiter(40, 40)

erroredItems, err := kyro.NewParallelQueue[int](40).
    WithItems(&ids).
    WithProgressNotifier(100, func(curr int, duration time.Duration, itemsPerSecond float64) {
        log.Printf("Processed %d items in %s (%.2f items/sec)", curr, duration, itemsPerSecond)
    }).
    WithErrorNotifier(func(err error, item int) {
        log.Printf("Error processing item %d: %v", item, err)
    }).
    OnProcessItem(func(item int) error {
        limiter.Wait()
        
        // Process item logic here
        if item%420 == 0 {
            return fmt.Errorf("simulated error for item %d", item)
        }
        
        return nil
    }).
    Process()

if err != nil {
    log.Printf("Processing completed with errors: %v", err)
    log.Printf("Failed items: %v", erroredItems)
}
```

**Key Queue Features:**
- Configurable worker pool size
- Progress notifications at batch intervals
- Per-item error notifications
- Thread-safe processing
- Returns all errored items for retry

### Parallel File Processing

Process large files line-by-line with concurrent workers:

```go
erroredLines, err := kyro.NewParallelFileProcessor(20).
    WithFilePath("data.txt").
    WithProgressNotifier(1000, func(curr int, duration time.Duration, linesPerSecond float64) {
        log.Printf("Processed %d lines in %s (%.2f lines/sec)", curr, duration, linesPerSecond)
    }).
    OnProcessLine(func(line []byte) error {
        // Process each line
        return processLine(line)
    }).
    Process()
```

### Rate Limiting

Simple wrapper around `golang.org/x/time/rate` for rate limiting:

```go
limiter := kyro.NewRateLimiter(10, 10) // 10 events per second, burst of 10

for i := 0; i < 100; i++ {
    limiter.Wait() // Blocks until allowed
    // Perform rate-limited operation
}
```

### Client Rotation

Thread-safe round-robin client rotation for load balancing:

```go
clients := []HTTPClient{client1, client2, client3}
rotator, err := kyro.NewRoundRobin(clients)

// Each call returns the next client in rotation
client := rotator.Next() // Returns client1
client = rotator.Next()  // Returns client2
client = rotator.Next()  // Returns client3
client = rotator.Next()  // Returns client1 (wraps around)
```

### Collections

Thread-safe generic set implementation:

```go
set := kyro.NewSimpleSet[int](100)

set.Add(1)
set.Add(2)
set.Add(2) // Duplicates are ignored

if set.Contains(1) {
    // Element exists
}

slice := set.AsSlice() // Convert to slice
set.Clear()            // Remove all elements
```

### Functional Utilities

Common functional programming utilities:

```go
// Map
numbers := []int{1, 2, 3, 4, 5}
doubled := kyro.Map(numbers, func(val int, index int) int {
    return val * 2
})

// Filter
evens := kyro.Filter(numbers, func(val int) bool {
    return val%2 == 0
})

// FindFirst
first := kyro.FindFirst(numbers, func(val int) bool {
    return val > 3
})
```

### Transformation Utilities

Utilities for handling default values and calculations:

```go
// NilIfDefault - returns nil for zero values
val := kyro.NilIfDefault(0)     // returns nil
val = kyro.NilIfDefault(42)     // returns pointer to 42

// NullishCoalescing - returns default if value is nil or zero
result := kyro.NullishCoalescing(nil, ptr(42))  // returns 42

// CalculateWeightedProportion
checks := []kyro.WeightedProportionCheck{
    {Score: 10, Condition: true},
    {Score: 5, Condition: false},
    {Score: 15, Condition: true},
}
proportion := kyro.CalculateWeightedProportion(checks) // Returns 0.833

// CalculateWeightedSum
sumChecks := []kyro.WeightedSumCheck{
    {Weight: 0.5, Value: 10.0},
    {Weight: 0.3, Value: 20.0},
    {Weight: 0.2, Value: 15.0},
}
sum := kyro.CalculateWeightedSum(sumChecks) // Returns 14.0
```

### File Utilities

Safe file operations:

```go
// Safely remove a file (no error if file doesn't exist)
err := kyro.SafeRemoveFile("temp.txt")
```

## 📚 Examples

Each example in the `examples/` directory is self-contained and can be run independently:

```bash
# Pipeline example
go run examples/pipeline/main.go

# Parallel queue example
go run examples/pqueue/main.go

# File processor example
go run examples/file_processor/main.go
```


## 🧹 Development

Run tests and vet from the project root:

```bash
mise run go:test
mise run go:vet
```

Or without mise:

```bash
go test -race $(go list ./... | grep -v /examples/)
go vet ./...
```

## 📄 License

This project is licensed under the MIT License. See the `LICENSE` file for details.

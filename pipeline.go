package kyro

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

// PipelineStep defines the function signature for a single step in a pipeline.
// It takes an input of any type and returns an output of any type or an error.
type PipelineStep func(input any, lastErr error) (output any, err error)

// GeneratorStep defines the function signature for a step that generates the initial input
// for a pipeline. It takes no input and returns an output of any type or an error.
type GeneratorStep func() (output any, err error)

var errExit error = errors.New("exit error")

// Execute runs a generator step followed by a pipeline step.
// It first calls the generator to get the initial input, and then passes this
// input to the pipeline step. It returns the output of the pipeline step or an error.
func Execute(pipeline PipelineStep) (output any, err error) {
	return pipeline(nil, nil)
}

// AsGenerator is a generic helper function that converts a function with a specific
// output type into a GeneratorStep. This is useful when the generator produces
// a specific type but needs to be used in a pipeline that expects any type.
func AsPipelineGenerator[O any](step func() (output O, err error)) PipelineStep {
	return func(input any, lastErr error) (output any, err error) {
		return step()
	}
}

// AsPipelineStep is a generic helper function that converts a function with specific
// input and output types into a PipelineStep. It handles type assertion for the input.
// If the input type assertion fails, it will panic. This is useful when a step operates
// on specific types but needs to be used in a pipeline that expects any type.
func AsPipelineStep[I any, O any](step func(input I, lastErr error) (output O, err error)) PipelineStep {
	return func(input any, lastErr error) (output any, err error) {
		asserted := AssertIn[I](input)
		return step(asserted, lastErr)
	}
}

// AssertIn is a helper function that asserts the type of the input to a specific type.
// If the assertion fails, it panics with a descriptive error message.
func AssertIn[T any](input any) T {
	if input == nil {
		var zeroValue T
		return zeroValue
	}

	value, ok := input.(T)
	if !ok {
		var zeroValue T
		panic(fmt.Sprintf("expected type %T, got %T", zeroValue, input))
	}
	return value
}

// InSequence creates a single PipelineStep that runs a sequence of provided pipeline steps.
// The output of each step becomes the input for the next step.
// If any step in the sequence returns an error, the InSequence step will return that error immediately.
func InSequence(steps ...PipelineStep) PipelineStep {
	return func(input any, lastErr error) (output any, err error) {
		currentInput := input
		currentErr := lastErr
		beforeExitErr := currentErr

		for _, step := range steps {
			currentInput, currentErr = step(currentInput, currentErr)

			if currentErr != nil && errors.Is(currentErr, errExit) {
				return nil, beforeExitErr
			}

			beforeExitErr = currentErr
		}

		return currentInput, currentErr
	}
}

// InParallel creates a single PipelineStep that runs multiple provided pipeline steps concurrently
// with the same input.
// The output will be a slice []any containing the results of each parallel step
// in the order the steps were provided. If any parallel step returns an error,
// the InParallel step will return the first error encountered.
func InParallel(steps ...PipelineStep) PipelineStep {
	return func(input any, lastErr error) (output any, err error) {
		numSteps := len(steps)

		if numSteps == 0 {
			return nil, nil
		}

		results := make([]any, numSteps)
		errCh := make(chan error, numSteps)
		var wg sync.WaitGroup

		for i, step := range steps {
			wg.Add(1)
			go func(index int, s PipelineStep) {
				defer wg.Done()
				out, stepErr := s(input, lastErr)
				if stepErr != nil {
					select {
					case errCh <- stepErr:
					default:
						// Error channel is full, another error has already been sent.
						// We prioritize the first error.
					}
					return
				}
				results[index] = out
			}(i, step)
		}

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case stepErr := <-errCh:
			return nil, stepErr
		case <-done:
			return results, nil
		}
	}
}

/* ======================== STEPS ======================== */

// RemoveFileStep creates a PipelineStep that removes the file at the given path
// if it exists. The step passes the input and error through, only returning
// an error if the file removal fails.
func RemoveFileStep(path string) PipelineStep {
	return AsPipelineStep(func(input any, err error) (any, error) {
		if _, err := os.Stat(path); err == nil {
			if err := os.Remove(path); err != nil {
				return input, err
			}
		}

		return input, err
	})
}

// ExitOnErrorStep creates a PipelineStep that immediately stops the pipeline
// if the previous step returned an error.
func ExitOnErrorStep() PipelineStep {
	return AsPipelineStep(func(input any, err error) (any, error) {
		if err != nil {
			return nil, errExit
		}

		return input, nil
	})
}

// TakeFirstStep creates a PipelineStep that takes the first N elements
// from a slice.  The value of n is the number of elements to take, so ids[:n]
// will take the first n elements.
func TakeFirstStep[T any](n int) PipelineStep {
	return AsPipelineStep(func(ids []T, err error) ([]T, error) {
		return ids[:n], err
	})
}

// TakeLastStep creates a PipelineStep that takes the last N elements
// from a slice
func TakeLastStep[T any](n int) PipelineStep {
	return AsPipelineStep(func(ids []T, err error) ([]T, error) {
		return ids[len(ids)-n:], err
	})
}

// TakeSubsetStep creates a PipelineStep that takes a subset of elements
// from a slice based on the provided start and end indices (exclusive of end).
// It returns an error if the start or end indices are out of bounds.
func TakeSubsetStep[T any](start, end int) PipelineStep {
	return AsPipelineStep(func(ids []T, err error) ([]T, error) {
		if start < 0 || end > len(ids) {
			return nil, fmt.Errorf("invalid range: %d-%d", start, end)
		}

		return ids[start:end], err
	})
}

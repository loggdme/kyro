package kyro_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/loggdme/kyro"
)

/* ====== Helper Functions ====== */

type ComplexType struct {
	Number int
	Slice  []string
}

func intGenerator() (int, error) {
	return 10, nil
}

func intToStringStep(input int, err error) (string, error) {
	return fmt.Sprintf("%d", input), nil
}

func errorGenerator() (any, error) {
	return nil, errors.New("error from generator")
}

func addOneStep(input int, err error) (int, error) {
	return input + 1, nil
}

func multiplyByTwoStep(input int, err error) (int, error) {
	return input * 2, nil
}

func sleepAndReturnIntStep(input int, duration time.Duration) kyro.PipelineStep {
	return kyro.AsPipelineStep(func(i int, err error) (int, error) {
		time.Sleep(duration)
		return input, nil
	})
}

/* ====== Test Cases ====== */

func TestExecute_Success(t *testing.T) {
	p := kyro.InSequence(
		kyro.AsPipelineGenerator(intGenerator),
		kyro.AsPipelineStep(intToStringStep),
	)

	output, err := kyro.Execute(p)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if output != "10" {
		t.Errorf("expected output '10', got %v", output)
	}
}

func TestExecute_GeneratorError(t *testing.T) {
	p := kyro.InSequence(
		kyro.AsPipelineGenerator(errorGenerator),
		kyro.ExitOnErrorStep(),
		kyro.AsPipelineStep(intToStringStep),
	)

	output, err := kyro.Execute(p)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "error from generator") {
		t.Errorf("expected error to contain 'error from generator', got: %v", err)
	}
	if output != nil {
		t.Errorf("expected nil output, got %v", output)
	}
}

func TestExecute_PipelineError(t *testing.T) {
	p := kyro.InSequence(
		kyro.AsPipelineGenerator(intGenerator),
		kyro.AsPipelineStep(func(input any, err error) (any, error) {
			return nil, errors.New("error from pipeline step")
		}),
	)

	output, err := kyro.Execute(p)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "error from pipeline step") {
		t.Errorf("expected error to contain 'error from pipeline step', got: %v", err)
	}
	if output != nil {
		t.Errorf("expected nil output, got %v", output)
	}
}

func TestAsPipelineGenerator_TypeConversion(t *testing.T) {
	generator := kyro.AsPipelineGenerator(intGenerator)
	output, err := kyro.Execute(generator)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if output != 10 {
		t.Errorf("expected output 10, got %v", output)
	}
}

func TestAsPipelineStep_Success(t *testing.T) {
	pipeline := kyro.AsPipelineStep(intToStringStep)
	input := 25

	output, err := pipeline(input, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if output != "25" {
		t.Errorf("expected output '25', got %v", output)
	}
}

func TestAsPipelineStep_InputTypeMismatch_Panics(t *testing.T) {
	pipeline := kyro.AsPipelineStep(intToStringStep)
	input := "hello"

	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected panic")
				return
			}
			if r != "expected type int, got string" {
				t.Errorf("expected panic 'expected type int, got string', got: %v", r)
			}
		}()
		pipeline(input, nil)
	}()
}

func TestAssertIn_Success(t *testing.T) {
	input := 123
	assertedValue := kyro.AssertIn[int](input)
	if assertedValue != 123 {
		t.Errorf("expected 123, got %v", assertedValue)
	}

	inputString := "test"
	assertedString := kyro.AssertIn[string](inputString)
	if assertedString != "test" {
		t.Errorf("expected 'test', got %v", assertedString)
	}
}

func TestAssertIn_Failure_Panics(t *testing.T) {
	input := 123
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected panic")
				return
			}
			if r != "expected type string, got int" {
				t.Errorf("expected panic 'expected type string, got int', got: %v", r)
			}
		}()
		kyro.AssertIn[string](input)
	}()

	inputAny := any(456)
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected panic")
				return
			}
			if r != "expected type bool, got int" {
				t.Errorf("expected panic 'expected type bool, got int', got: %v", r)
			}
		}()
		kyro.AssertIn[bool](inputAny)
	}()
}

func TestInSequence_Success(t *testing.T) {
	step1 := kyro.AsPipelineStep(addOneStep)
	step2 := kyro.AsPipelineStep(multiplyByTwoStep)
	step3 := kyro.AsPipelineStep(intToStringStep)

	sequence := kyro.InSequence(step1, step2, step3)

	input := 5
	output, err := sequence(input, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if output != "12" {
		t.Errorf("expected output '12', got %v", output)
	}
}

func TestInSequence_ErrorInMiddle(t *testing.T) {
	step1 := kyro.AsPipelineStep(addOneStep)
	step2 := func(input any, err error) (any, error) {
		return nil, errors.New("error in sequence step")
	}
	step3 := kyro.AsPipelineStep(multiplyByTwoStep)

	sequenceWithErr := kyro.InSequence(step1, step2, kyro.ExitOnErrorStep(), step3)

	input := 5
	output, err := sequenceWithErr(input, nil)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "error in sequence step") {
		t.Errorf("expected error to contain 'error in sequence step', got: %v", err)
	}
	if output != nil {
		t.Errorf("expected nil output, got %v", output)
	}
}

func TestInSequence_EmptySequence(t *testing.T) {
	sequence := kyro.InSequence()
	input := "initial input"

	output, err := sequence(input, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if output != "initial input" {
		t.Errorf("expected output 'initial input', got %v", output) // Should return the original input
	}
}

func TestInSequence_SingleStep(t *testing.T) {
	step := kyro.AsPipelineStep(intToStringStep)
	sequence := kyro.InSequence(step)
	input := 42

	output, err := sequence(input, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if output != "42" {
		t.Errorf("expected output '42', got %v", output)
	}
}

func TestInParallel_Success(t *testing.T) {
	step1 := kyro.AsPipelineStep(func(input int, err error) (string, error) {
		return fmt.Sprintf("step1: %d", input), nil
	})
	step2 := kyro.AsPipelineStep(func(input int, err error) (int, error) {
		return input * 10, nil
	})
	step3 := kyro.AsPipelineStep(func(input int, err error) (bool, error) {
		return input > 5, nil
	})

	parallel := kyro.InParallel(step1, step2, step3)
	input := 7

	output, err := parallel(input, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	results, ok := output.([]any)
	if !ok {
		t.Error("expected output to be a slice of any")
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	if results[0] != "step1: 7" {
		t.Errorf("expected results[0] 'step1: 7', got %v", results[0])
	}
	if results[1] != 70 {
		t.Errorf("expected results[1] 70, got %v", results[1])
	}
	if results[2] != true {
		t.Errorf("expected results[2] true, got %v", results[2])
	}
}

func TestInParallel_ErrorInOneStep(t *testing.T) {
	step1 := kyro.AsPipelineStep(addOneStep)
	errorStep := func(input any, err error) (any, error) {
		return nil, errors.New("parallel error")
	}
	step3 := kyro.AsPipelineStep(multiplyByTwoStep)

	parallel := kyro.InParallel(step1, errorStep, step3)
	input := 10

	output, err := parallel(input, nil)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "parallel error") {
		t.Errorf("expected error to contain 'parallel error', got: %v", err)
	}
	if output != nil {
		t.Errorf("expected nil output, got %v", output)
	}
}

func TestInParallel_MultipleErrors(t *testing.T) {
	errorStep1 := func(input any, err error) (any, error) {
		time.Sleep(50 * time.Millisecond)
		return nil, errors.New("first parallel error")
	}
	errorStep2 := func(input any, err error) (any, error) {
		time.Sleep(10 * time.Millisecond)
		return nil, errors.New("second parallel error")
	}

	parallel := kyro.InParallel(errorStep1, errorStep2)
	input := "some input"

	output, err := parallel(input, nil)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "second parallel error") {
		t.Errorf("expected error to contain 'second parallel error', got: %v", err)
	}
	if output != nil {
		t.Errorf("expected nil output, got %v", output)
	}
}

func TestInParallel_EmptyParallel(t *testing.T) {
	parallel := kyro.InParallel()
	input := "initial input"

	output, err := parallel(input, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if output != nil {
		t.Errorf("expected nil output, got %v", output)
	}
}

func TestInParallel_ConcurrencyCheckInOrder(t *testing.T) {
	step1 := sleepAndReturnIntStep(1, 200*time.Millisecond)
	step2 := sleepAndReturnIntStep(2, 50*time.Millisecond)

	parallel := kyro.InParallel(step1, step2)
	input := 0

	startTime := time.Now()
	output, err := parallel(input, nil)
	endTime := time.Now()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	results, ok := output.([]any)
	if !ok {
		t.Error("expected output to be []any")
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	duration := endTime.Sub(startTime)
	if duration >= 300*time.Millisecond {
		t.Error("parallel steps should run concurrently")
	}

	if results[0] != 1 {
		t.Errorf("expected results[0] 1, got %v", results[0])
	}
	if results[1] != 2 {
		t.Errorf("expected results[1] 2, got %v", results[1])
	}
}

func TestInSequence_WithParallelSteps(t *testing.T) {
	// Step 1: Add 1 to input
	step1 := kyro.AsPipelineStep(addOneStep)

	// Step 2: Run two steps in parallel: multiply by 2 and convert to string
	parallelStep := kyro.InParallel(
		kyro.AsPipelineStep(multiplyByTwoStep),
		kyro.AsPipelineStep(intToStringStep),
	)

	// Step 3: Combine the results from the parallel step (assuming they are strings)
	// This requires a step that can handle the []any input from InParallel.
	// Let's create one that expects []any and casts its elements.
	combineResultsStep := func(input any, err error) (any, error) {
		results, ok := input.([]any)
		if !ok {
			return nil, errors.New("expected []any input for combining results")
		}

		if len(results) != 2 {
			return nil, errors.New("expected 2 results from parallel step")
		}
		num, ok := results[0].(int)
		if !ok {
			return nil, errors.New("expected first parallel result to be int")
		}
		str, ok := results[1].(string)
		if !ok {
			return nil, errors.New("expected second parallel result to be string")
		}
		return fmt.Sprintf("Num: %d, Str: %s", num, str), nil
	}

	sequence := kyro.InSequence(step1, parallelStep, combineResultsStep)

	input := 5 // 5 -> 6 -> parallel({12, "6"}) -> "Num: 12, Str: 6"

	output, err := sequence(input, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if output != "Num: 12, Str: 6" {
		t.Errorf("expected output 'Num: 12, Str: 6', got %v", output)
	}
}

func TestInParallel_InputPropagation(t *testing.T) {
	// Test that the same input is passed to each parallel step.
	step1 := kyro.AsPipelineStep(func(input int, err error) (int, error) {
		return input + 1, nil
	})
	step2 := kyro.AsPipelineStep(func(input int, err error) (int, error) {
		return input * 2, nil
	})

	parallel := kyro.InParallel(step1, step2)
	input := 10

	output, err := parallel(input, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	results, ok := output.([]any)
	if !ok {
		t.Error("expected output to be []any")
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	if results[0] != 11 {
		t.Errorf("expected results[0] 11, got %v", results[0])
	}
	if results[1] != 20 {
		t.Errorf("expected results[1] 20, got %v", results[1])
	}
}

func TestComplexTypePipeline(t *testing.T) {
	generator := kyro.AsPipelineGenerator(func() (ComplexType, error) {
		return ComplexType{Number: 10, Slice: []string{"a", "b"}}, nil
	})

	step1 := kyro.AsPipelineStep(func(input ComplexType, err error) (ComplexType, error) {
		input.Number = input.Number * 2
		input.Slice = append(input.Slice, "c")
		return input, nil
	})

	step2 := kyro.AsPipelineStep(func(input ComplexType, err error) (ComplexType, error) {
		input.Number = input.Number - 5
		input.Slice[0] = "z"
		return input, nil
	})

	p := kyro.InSequence(generator, step1, step2)

	output, err := kyro.Execute(p)

	if err != nil {
		t.Fatalf("pipeline execution failed: %v", err)
	}

	result, ok := output.(ComplexType)
	if !ok {
		t.Fatalf("expected output of type ComplexType, got %T", output)
	}

	expected := ComplexType{Number: 15, Slice: []string{"z", "b", "c"}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected output %v, got %v", expected, result)
	}
}

func TestComplexTypeParallelPipeline(t *testing.T) {
	generator := kyro.AsPipelineGenerator(func() (ComplexType, error) {
		return ComplexType{Number: 10, Slice: []string{"a", "b"}}, nil
	})

	stepNum := kyro.AsPipelineStep(func(input ComplexType, err error) (ComplexType, error) {
		return ComplexType{Number: input.Number * 3, Slice: input.Slice}, nil
	})

	stepSlice := kyro.AsPipelineStep(func(input ComplexType, err error) ([]string, error) {
		return append(input.Slice, "c", "d"), nil
	})

	p := kyro.InSequence(
		generator,
		kyro.InParallel(stepNum, stepSlice),
	)

	output, err := kyro.Execute(p)

	if err != nil {
		t.Fatalf("parallel pipeline execution failed: %v", err)
	}

	results, ok := output.([]any)
	if !ok {
		t.Fatalf("expected output of type []any, got %T", output)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	complexResult, ok := results[0].(ComplexType)
	if !ok {
		t.Fatalf("expected first result to be int, got %T", results[0])
	}
	expectedNum := 30
	if complexResult.Number != expectedNum {
		t.Errorf("expected number result %d, got %d", expectedNum, complexResult.Number)
	}

	sliceResult, ok := results[1].([]string)
	if !ok {
		t.Fatalf("expected second result to be []string, got %T", results[1])
	}
	expectedSlice := []string{"a", "b", "c", "d"}
	if !reflect.DeepEqual(sliceResult, expectedSlice) {
		t.Errorf("expected slice result %v, got %v", expectedSlice, sliceResult)
	}
}

func TestInParallel_NilInput(t *testing.T) {
	step1 := func(input any, err error) (any, error) {
		if input != nil {
			t.Errorf("expected nil input, got %v", input)
		}
		return "step1 received nil", nil
	}
	step2 := func(input any, err error) (any, error) {
		if input != nil {
			t.Errorf("expected nil input, got %v", input)
		}
		return "step2 received nil", nil
	}

	parallel := kyro.InParallel(step1, step2)

	output, err := parallel(nil, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	results, ok := output.([]any)
	if !ok {
		t.Error("expected output to be []any")
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	if results[0] != "step1 received nil" {
		t.Errorf("expected results[0] 'step1 received nil', got %v", results[0])
	}
	if results[1] != "step2 received nil" {
		t.Errorf("expected results[1] 'step2 received nil', got %v", results[1])
	}
}

func TestInSequence_NilInput(t *testing.T) {
	step1 := func(input any, err error) (any, error) {
		if input != nil {
			t.Errorf("expected nil input, got %v", input)
		}
		return "step1 received nil", nil
	}
	step2 := func(input any, err error) (any, error) {
		if input != "step1 received nil" {
			t.Errorf("expected input 'step1 received nil', got %v", input)
		}
		return "step2 received step1 output", nil
	}

	sequence := kyro.InSequence(step1, step2)

	output, err := sequence(nil, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if output != "step2 received step1 output" {
		t.Errorf("expected output 'step2 received step1 output', got %v", output)
	}
}

func TestInSequence_StepReturnsNilOutput(t *testing.T) {
	step1 := func(input any, err error) (any, error) {
		return "output from step1", nil
	}
	step2 := func(input any, err error) (any, error) {
		return nil, nil
	}
	step3 := func(input any, err error) (any, error) {
		if input != nil {
			t.Errorf("expected nil input, got %v", input)
		}
		return "step3 received nil", nil
	}

	sequence := kyro.InSequence(step1, step2, step3)
	input := "initial input"

	output, err := sequence(input, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if output != "step3 received nil" {
		t.Errorf("expected output 'step3 received nil', got %v", output)
	}
}

func TestInParallel_StepsReturnNilOutput(t *testing.T) {
	step1 := func(input any, err error) (any, error) {
		return "output 1", nil
	}
	step2 := func(input any, err error) (any, error) {
		return nil, nil
	}
	step3 := func(input any, err error) (any, error) {
		return "output 3", nil
	}

	parallel := kyro.InParallel(step1, step2, step3)
	input := "initial input"

	output, err := parallel(input, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	results, ok := output.([]any)
	if !ok {
		t.Error("expected output to be []any")
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	if results[0] != "output 1" {
		t.Errorf("expected results[0] 'output 1', got %v", results[0])
	}
	if results[1] != nil {
		t.Errorf("expected results[1] nil, got %v", results[1])
	}
	if results[2] != "output 3" {
		t.Errorf("expected results[2] 'output 3', got %v", results[2])
	}
}

func TestErrorHandler_Sequential(t *testing.T) {
	step1 := func(input any, err error) (any, error) {
		return nil, errors.New("error in step 1")
	}

	step2 := func(input any, err error) (any, error) {
		return "step 2 output", err
	}

	sequence := kyro.InSequence(step1, step2)
	input := "initial input"

	output, err := sequence(input, nil)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "error in step 1") {
		t.Errorf("expected error to contain 'error in step 1', got: %v", err)
	}
	if output != "step 2 output" {
		t.Errorf("expected output 'step 2 output', got %v", output)
	}
}

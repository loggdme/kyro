package kyro_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/loggdme/kyro"
	"github.com/stretchr/testify/assert"
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

	assert.NoError(t, err)
	assert.Equal(t, "10", output)
}

func TestExecute_GeneratorError(t *testing.T) {
	p := kyro.InSequence(
		kyro.AsPipelineGenerator(errorGenerator),
		kyro.ExitOnErrorStep(),
		kyro.AsPipelineStep(intToStringStep),
	)

	output, err := kyro.Execute(p)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error from generator")
	assert.Nil(t, output)
}

func TestExecute_PipelineError(t *testing.T) {
	p := kyro.InSequence(
		kyro.AsPipelineGenerator(intGenerator),
		kyro.AsPipelineStep(func(input any, err error) (any, error) {
			return nil, errors.New("error from pipeline step")
		}),
	)

	output, err := kyro.Execute(p)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error from pipeline step")
	assert.Nil(t, output)
}

func TestAsPipelineGenerator_TypeConversion(t *testing.T) {
	generator := kyro.AsPipelineGenerator(intGenerator)
	output, err := kyro.Execute(generator)

	assert.NoError(t, err)
	assert.Equal(t, 10, output)
}

func TestAsPipelineStep_Success(t *testing.T) {
	pipeline := kyro.AsPipelineStep(intToStringStep)
	input := 25

	output, err := pipeline(input, nil)

	assert.NoError(t, err)
	assert.Equal(t, "25", output)
}

func TestAsPipelineStep_InputTypeMismatch_Panics(t *testing.T) {
	pipeline := kyro.AsPipelineStep(intToStringStep)
	input := "hello"

	assert.PanicsWithValue(t, "expected type int, got string", func() {
		pipeline(input, nil)
	})
}

func TestAssertIn_Success(t *testing.T) {
	input := 123
	assertedValue := kyro.AssertIn[int](input)
	assert.Equal(t, 123, assertedValue)

	inputString := "test"
	assertedString := kyro.AssertIn[string](inputString)
	assert.Equal(t, "test", assertedString)
}

func TestAssertIn_Failure_Panics(t *testing.T) {
	input := 123
	assert.PanicsWithValue(t, "expected type string, got int", func() {
		kyro.AssertIn[string](input)
	})

	inputAny := any(456)
	assert.PanicsWithValue(t, "expected type bool, got int", func() {
		kyro.AssertIn[bool](inputAny)
	})
}

func TestInSequence_Success(t *testing.T) {
	step1 := kyro.AsPipelineStep(addOneStep)
	step2 := kyro.AsPipelineStep(multiplyByTwoStep)
	step3 := kyro.AsPipelineStep(intToStringStep)

	sequence := kyro.InSequence(step1, step2, step3)

	input := 5
	output, err := sequence(input, nil)

	assert.NoError(t, err)
	assert.Equal(t, "12", output)
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

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error in sequence step")
	assert.Nil(t, output)
}

func TestInSequence_EmptySequence(t *testing.T) {
	sequence := kyro.InSequence()
	input := "initial input"

	output, err := sequence(input, nil)

	assert.NoError(t, err)
	assert.Equal(t, "initial input", output) // Should return the original input
}

func TestInSequence_SingleStep(t *testing.T) {
	step := kyro.AsPipelineStep(intToStringStep)
	sequence := kyro.InSequence(step)
	input := 42

	output, err := sequence(input, nil)

	assert.NoError(t, err)
	assert.Equal(t, "42", output)
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

	assert.NoError(t, err)

	results, ok := output.([]any)
	assert.True(t, ok, "expected output to be a slice of any")
	assert.Len(t, results, 3)

	assert.Equal(t, "step1: 7", results[0])
	assert.Equal(t, 70, results[1])
	assert.Equal(t, true, results[2])
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

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parallel error")
	assert.Nil(t, output)
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

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "second parallel error")
	assert.Nil(t, output)
}

func TestInParallel_EmptyParallel(t *testing.T) {
	parallel := kyro.InParallel()
	input := "initial input"

	output, err := parallel(input, nil)

	assert.NoError(t, err)
	assert.Nil(t, output)
}

func TestInParallel_ConcurrencyCheckInOrder(t *testing.T) {
	step1 := sleepAndReturnIntStep(1, 200*time.Millisecond)
	step2 := sleepAndReturnIntStep(2, 50*time.Millisecond)

	parallel := kyro.InParallel(step1, step2)
	input := 0

	startTime := time.Now()
	output, err := parallel(input, nil)
	endTime := time.Now()

	assert.NoError(t, err)

	results, ok := output.([]any)
	assert.True(t, ok)
	assert.Len(t, results, 2)

	duration := endTime.Sub(startTime)
	assert.True(t, duration < 300*time.Millisecond, "parallel steps should run concurrently")

	assert.Equal(t, 1, results[0])
	assert.Equal(t, 2, results[1])
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

	assert.NoError(t, err)
	assert.Equal(t, "Num: 12, Str: 6", output)
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

	assert.NoError(t, err)
	results, ok := output.([]any)
	assert.True(t, ok)
	assert.Len(t, results, 2)

	assert.Equal(t, 11, results[0])
	assert.Equal(t, 20, results[1])
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
		assert.Nil(t, input)
		return "step1 received nil", nil
	}
	step2 := func(input any, err error) (any, error) {
		assert.Nil(t, input)
		return "step2 received nil", nil
	}

	parallel := kyro.InParallel(step1, step2)

	output, err := parallel(nil, nil)

	assert.NoError(t, err)
	results, ok := output.([]any)
	assert.True(t, ok)
	assert.Len(t, results, 2)

	assert.Equal(t, "step1 received nil", results[0])
	assert.Equal(t, "step2 received nil", results[1])
}

func TestInSequence_NilInput(t *testing.T) {
	step1 := func(input any, err error) (any, error) {
		assert.Nil(t, input)
		return "step1 received nil", nil
	}
	step2 := func(input any, err error) (any, error) {
		assert.Equal(t, "step1 received nil", input)
		return "step2 received step1 output", nil
	}

	sequence := kyro.InSequence(step1, step2)

	output, err := sequence(nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, "step2 received step1 output", output)
}

func TestInSequence_StepReturnsNilOutput(t *testing.T) {
	step1 := func(input any, err error) (any, error) {
		return "output from step1", nil
	}
	step2 := func(input any, err error) (any, error) {
		return nil, nil
	}
	step3 := func(input any, err error) (any, error) {
		assert.Nil(t, input)
		return "step3 received nil", nil
	}

	sequence := kyro.InSequence(step1, step2, step3)
	input := "initial input"

	output, err := sequence(input, nil)

	assert.NoError(t, err)
	assert.Equal(t, "step3 received nil", output)
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

	assert.NoError(t, err)
	results, ok := output.([]any)
	assert.True(t, ok)
	assert.Len(t, results, 3)

	assert.Equal(t, "output 1", results[0])
	assert.Nil(t, results[1])
	assert.Equal(t, "output 3", results[2])
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

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error in step 1")
	assert.Equal(t, output, "step 2 output")
}

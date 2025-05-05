package main

import (
	"fmt"

	"github.com/loggdme/kyro/pkg/pipeline"
)

func main() {
	generateItems := pipeline.AsPipelineGenerator(func() (string, error) {
		return "Hello, Kyro Pipeline!", nil
	})

	stringLength := pipeline.AsPipelineStep(func(input string) (int, error) {
		return len(input), nil
	})

	double := pipeline.AsPipelineStep(func(input int) (int, error) {
		return input * 2, nil
	})

	triple := pipeline.AsPipelineStep(func(input int) (int, error) {
		return input * 3, nil
	})

	add := pipeline.AsPipelineStep(func(input []any) (int, error) {
		first, second := pipeline.AssertIn[int](input[0]), pipeline.AssertIn[int](input[1])
		return first + second, nil
	})

	result, err := pipeline.Execute(
		generateItems,
		pipeline.InSequence(
			stringLength,
			pipeline.InParallel(double, triple),
			add,
		),
	)

	if err != nil {
		fmt.Printf("Pipeline execution failed: %v\n", err)
		return
	}

	fmt.Printf("Pipeline execution successful. Final result: %v\n", result)
}

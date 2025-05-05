package main

import (
	"fmt"

	"github.com/loggdme/kyro"
)

func main() {
	generateItems := kyro.AsPipelineGenerator(func() (string, error) {
		return "Hello, Kyro Pipeline!", nil
	})

	stringLength := kyro.AsPipelineStep(func(input string) (int, error) {
		return len(input), nil
	})

	double := kyro.AsPipelineStep(func(input int) (int, error) {
		return input * 2, nil
	})

	triple := kyro.AsPipelineStep(func(input int) (int, error) {
		return input * 3, nil
	})

	add := kyro.AsPipelineStep(func(input []any) (int, error) {
		first, second := kyro.AssertIn[int](input[0]), kyro.AssertIn[int](input[1])
		return first + second, nil
	})

	result, err := kyro.Execute(
		generateItems,
		kyro.InSequence(
			stringLength,
			kyro.InParallel(double, triple),
			add,
		),
	)

	if err != nil {
		fmt.Printf("Pipeline execution failed: %v\n", err)
		return
	}

	fmt.Printf("Pipeline execution successful. Final result: %v\n", result)
}

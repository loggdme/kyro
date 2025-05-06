package main

import (
	"fmt"

	"github.com/loggdme/kyro"
)

func main() {
	generateItems := kyro.AsPipelineGenerator(func() (string, error) {
		return "Hello, Kyro Pipeline!", nil
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

	if err != nil {
		fmt.Printf("Pipeline execution failed: %v\n", err)
		return
	}

	fmt.Printf("Pipeline execution successful. Final result: %v\n", result)
}

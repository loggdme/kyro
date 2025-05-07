package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/loggdme/kyro"
)

func main() {
	type Object struct {
		ID int `json:"id"`
	}

	erroredItems, err := kyro.NewParallelFileProcessor(5).
		WithFilePath("examples/file_processor/test.jsonl").
		WithProgressNotifier(10, func(curr int, duration time.Duration, itemsPerSecond float64) {
			log.Printf("Processed %d lines in %s (%.2f lines/sec)", curr, duration, itemsPerSecond)
		}).
		WithErrorNotifier(func(err error, line []byte) {
			log.Printf("Error processing line %s: %v", line, err)
		}).
		OnProcessLine(func(line []byte) error {
			var obj Object
			if err := json.Unmarshal(line, &obj); err != nil {
				fmt.Printf("Error unmarshaling JSON: %v\n", err)
			}
			return nil
		}).
		Process()

	if err != nil {
		fmt.Printf("Error during data processing: %v", err)
		fmt.Println("Errored lines:", erroredItems)
	}
}

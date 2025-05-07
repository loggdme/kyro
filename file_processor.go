package kyro

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ParallelFileProcessor represents a processor for reading and processing a file line by line in parallel.
type ParallelFileProcessor struct {
	filePath        string
	numberOfWorkers int

	processLineFunc ProcessFunc[[]byte]
	processed       int
	processedMutex  sync.Mutex

	progressBatch int
	progressFunc  ProgressNotifier

	errorFunc ErrorNotifier[[]byte]
}

// NewParallelFileProcessor creates a new ParallelFileProcessor with the specified number of workers.
func NewParallelFileProcessor(numberOfWorkers int) *ParallelFileProcessor {
	return &ParallelFileProcessor{
		numberOfWorkers: numberOfWorkers,
		progressBatch:   100,
	}
}

// WithFilePath sets the path to the file to be processed.
func (p *ParallelFileProcessor) WithFilePath(filePath string) *ParallelFileProcessor {
	p.filePath = filePath
	return p
}

// OnProcessLine sets the function to be used for processing each line.
func (p *ParallelFileProcessor) OnProcessLine(processLineFunc ProcessFunc[[]byte]) *ParallelFileProcessor {
	p.processLineFunc = processLineFunc
	return p
}

// WithProgressNotifier sets the progress notification function and the batch size.
// batch is the number of lines processed before the progress function is called.
func (p *ParallelFileProcessor) WithProgressNotifier(batch int, progressFunc ProgressNotifier) *ParallelFileProcessor {
	p.progressFunc = progressFunc
	p.progressBatch = batch
	return p
}

// WithErrorNotifier sets the error notification function.
// errorFunc is the function to call when an error occurs during processing.
func (p *ParallelFileProcessor) WithErrorNotifier(errorFunc ErrorNotifier[[]byte]) *ParallelFileProcessor {
	p.errorFunc = errorFunc
	return p
}

// Process starts the parallel processing of the file. It returns a slice of lines
// that failed to process and an error if any critical error occurred during setup or processing.
func (p *ParallelFileProcessor) Process() (*[][]byte, error) {
	var erroredLines [][]byte

	if p.numberOfWorkers <= 0 {
		return &erroredLines, fmt.Errorf("number of workers must be positive")
	}

	if p.filePath == "" {
		return &erroredLines, fmt.Errorf("file path must be set")
	}

	if p.processLineFunc == nil {
		return &erroredLines, fmt.Errorf("process line function must be set")
	}

	file, err := os.Open(p.filePath)
	if err != nil {
		return &erroredLines, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	lineCh := make(chan []byte, p.numberOfWorkers)
	errCh := make(chan []byte, p.numberOfWorkers)

	var wg sync.WaitGroup
	wg.Add(p.numberOfWorkers)

	startTime := time.Now()

	worker := func() {
		defer wg.Done()
		for line := range lineCh {
			if err := p.processLineFunc(line); err != nil {
				select {
				// Attempt to send the errored line to the error channel.
				case errCh <- line:
					if p.errorFunc != nil {
						p.errorFunc(err, line)
					}
				// If the error channel is full, we report this as an error
				// before attempting to report the original processing error.
				default:
					if p.errorFunc != nil {
						p.errorFunc(fmt.Errorf("error channel is full"), line)
						p.errorFunc(err, line)
					}
				}
			}

			p.processedMutex.Lock()
			p.processed++
			currentProcessed := p.processed
			p.processedMutex.Unlock()

			if p.progressFunc != nil && currentProcessed%p.progressBatch == 0 {
				duration := time.Since(startTime)
				linesPerSecond := float64(currentProcessed) / duration.Seconds()
				p.progressFunc(currentProcessed, duration, linesPerSecond)
			}
		}
	}

	for range p.numberOfWorkers {
		go worker()
	}

	go func() {
		reader := bufio.NewReader(file)

		for {
			lineBytes, err := reader.ReadBytes('\n')

			if err != nil {
				if err == io.EOF {
					break
				}

				fmt.Fprintf(os.Stderr, "read error: %v\n", err)
				break
			}

			if len(lineBytes) > 0 && lineBytes[len(lineBytes)-1] == '\n' {
				lineBytes = lineBytes[:len(lineBytes)-1]
			}

			lineCh <- lineBytes
		}
		close(lineCh)
	}()

	wg.Wait()
	close(errCh)

	for errLine := range errCh {
		erroredLines = append(erroredLines, errLine)
	}

	if len(erroredLines) > 0 {
		return &erroredLines, fmt.Errorf("encountered %d errors during line processing", len(erroredLines))
	}

	return &erroredLines, nil
}

package hasher

import (
	"context"
	"fmt"
	"golang.org/x/crypto/sha3"
	"runtime"
	"sync"
)

type job struct {
	index int
	value string
}

type result struct {
	index int
	hash  string
}

func HashStringsParallel(ctx context.Context, input []string) ([]string, error) {
	numWorkers := runtime.NumCPU()
	jobs := make(chan job, len(input))
	results := make(chan result, len(input))
	n := len(input)

	var wg sync.WaitGroup

	// fan-out
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for j := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}
				sum := sha3.Sum256([]byte(j.value))
				h := fmt.Sprintf("%x", sum)

				select {
				case <-ctx.Done():
					return
				case results <- result{index: j.index, hash: h}:
				}
			}
		}()

	}

	// send jobs
	go func() {
		for i, s := range input {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case jobs <- job{index: i, value: s}:
			}
		}
		close(jobs)
	}()

	// waiting for jobs to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// fan-in
	output := make([]string, n)
	received := 0
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r, ok := <-results:
			if !ok {
				return output, nil
			}
			output[r.index] = r.hash
			received++
			if received == n {
				return output, nil
			}
		}
	}
}

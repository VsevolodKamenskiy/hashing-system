package hasher

import (
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

func HashStringsParallel(input []string) []string {
	numWorkers := runtime.NumCPU()
	jobs := make(chan job, len(input))
	results := make(chan result, len(input))

	var wg sync.WaitGroup

	// fan-out
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for j := range jobs {
				hash := sha3.Sum256([]byte(j.value))
				results <- result{
					index: j.index,
					hash:  fmt.Sprintf("%x", hash),
				}
			}
		}()

	}

	// отправляем задания
	for i, str := range input {
		jobs <- job{index: i, value: str}
	}
	close(jobs)

	// закрываем results, когда все воркеры закончат
	go func() {
		wg.Wait()
		close(results)
	}()

	// fan-in
	output := make([]string, len(input))
	for res := range results {
		output[res.index] = res.hash
	}

	return output
}

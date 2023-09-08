package workflowsupport

import "sync"

func ProcessAsync[T any, U any](inputs []U, block func(u U) (T, error)) []T {
	var results []T
	resultChannel := make(chan T)
	wg := sync.WaitGroup{}

	for _, i := range inputs {
		wg.Add(1)

		go func(input U) {
			defer wg.Done()

			result, err := block(input)
			if err != nil {
				return
			}

			resultChannel <- result
		}(i)
	}

	go func() {
		wg.Wait()
		close(resultChannel)
	}()

	for result := range resultChannel {
		results = append(results, result)
	}

	return results
}

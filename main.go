package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {

	integI := make(chan int)
	integTwo := make(chan int)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer close(integTwo)
		for i := 190; i < 200; i++ {
			integTwo <- i
		}
	}()

	go func() {
		defer close(integI)
		for i := 0; i < 100; i++ {
			integI <- i
		}
	}()

	appCtx, cancel := context.WithCancel(context.Background())
	resChan := fanIn(appCtx, integI, integTwo)
	go func(resChan <-chan int) {
		for res := range resChan {
			fmt.Println(res)
		}
	}(resChan)

	time.Sleep(1 * time.Second)
	cancel()
	// for res := range resChan {
	// 	fmt.Println(res)
	// }

}

func fanIn(ctx context.Context, chanels ...<-chan int) <-chan int {
	combinedFetcher := make(chan int, 110)

	var wg sync.WaitGroup
	wg.Add(len(chanels))

	for _, f := range chanels {
		f := f
		go func() {
			defer wg.Done()
			for {
				select {
				case res, ok := <-f:
					if !ok {
						fmt.Println("Channel closed, exiting goroutine")
						return
					}
					select {
					case combinedFetcher <- res:
						fmt.Println("Sent to combinedFetcher:", res)
					case <-ctx.Done():
						fmt.Println("Context done, exiting goroutine")
						return
					}
				case <-ctx.Done():
					fmt.Println("Context done, exiting goroutine (outer select)")
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(combinedFetcher)
		fmt.Println("CombinedFetcher closed")
	}()

	return combinedFetcher
}

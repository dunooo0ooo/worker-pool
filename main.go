package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

func worker(id int, jobs <-chan string, quit <-chan struct{}) {
	for {
		select {
		case j, ok := <-jobs:
			if !ok {
				fmt.Println("worker", id, "exiting")
				return
			}
			fmt.Println("worker", id, "started job", j)
			time.Sleep(time.Second)
			fmt.Println("worker", id, "finished job", j)
		case <-quit:
			fmt.Println("worker", id, "terminated by manager")
			return
		}
	}
}

func main() {
	jobs := make(chan string, 100)
	addWorker := make(chan struct{})
	removeWorker := make(chan struct{})

	var wg sync.WaitGroup
	var workerCount int
	var quitChannels []chan struct{}

	go func() {
		for {
			select {
			case <-addWorker:
				workerCount++
				quitChan := make(chan struct{})
				quitChannels = append(quitChannels, quitChan)
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					worker(id, jobs, quitChan)
				}(workerCount)
				fmt.Println("Added worker", workerCount)

			case <-removeWorker:
				if workerCount > 0 {
					workerCount--
					close(quitChannels[workerCount])
					quitChannels = quitChannels[:workerCount]
					fmt.Println("Removed worker", workerCount+1)
				}
			}
		}
	}()

	for i := 0; i < 15; i++ {
		addWorker <- struct{}{}
	}

	for j := 1; j <= 5; j++ {
		jobs <- strconv.Itoa(j)
	}
	close(jobs)

	time.Sleep(2 * time.Second)
	addWorker <- struct{}{}
	time.Sleep(1 * time.Second)
	removeWorker <- struct{}{}

	go func() {
		wg.Wait()
	}()

	fmt.Println("All workers finished")
}

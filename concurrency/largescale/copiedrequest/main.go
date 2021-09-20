package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// runtimeの条件（マシン、ネットワーク状況、データストアへのパスなど）で処理のスピードに差が出る可能性がある
// よって、リクエストを複製し、最速で処理し終わったものの結果を使い、残りはdoneチャンネルによって終了するテクニックがある
// でも複製した分だけ無駄なリソースを使うってことだよなー完全にトレードオフ
func main() {
	done := make(chan interface{})
	result := make(chan int)

	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go dowork(done, i, &wg, result)
	}

	firstReturned := <-result
	close(done)
	wg.Wait()

	fmt.Printf("Received an answer from #%v\n", firstReturned)
}

func dowork(done <-chan interface{}, id int, wg *sync.WaitGroup, result chan<- int) {
	started := time.Now()
	defer wg.Done()

	// simulate random load
	simulatedLoadTime := time.Duration(1+rand.Intn(5)) * time.Second
	select {
	case <-done:
	case <-time.After(simulatedLoadTime):
	}

	select {
	case <-done:
	case result <- id:
	}

	took := time.Since(started)
	//Display how long handlers would have taken
	if took < simulatedLoadTime {
		took = simulatedLoadTime
	}
	fmt.Printf("%v took %v\n", id, took)
}

package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

// 時間のかかる処理を広げて（fanout）、その後収縮させる（fanin）。素数の判定に時間がかかるので、その処理をCPU分並列にして、前後の処理はパイプラインでつないでいる。
func main() {
	rand := func() interface{} { return rand.Intn(50000000) }
	done := make(chan interface{})
	defer close(done)

	start := time.Now()

	randIntStream := toInt(done, repeatFn(done, rand))

	numFinders := runtime.NumCPU()
	finders := make([]<-chan int, numFinders)
	for i := 0; i < numFinders; i++ {
		finders[i] = primeFinder(done, randIntStream)
	}

	fmt.Println("Primes:")
	for prime := range take(done, funIn(done, finders...), 10) {
		fmt.Printf("\t%d\n", prime)
	}

	fmt.Printf("Search took: %v\n", time.Since(start))
}

func repeatFn(
	done <-chan interface{},
	fn func() interface{},
) <-chan interface{} {
	valueStream := make(chan interface{})
	go func() {
		defer close(valueStream)
		for {
			select {
			case <-done:
				return
			case valueStream <- fn():
			}
		}
	}()
	return valueStream
}

func toInt(
	done <-chan interface{},
	valueStream <-chan interface{},
) <-chan int {
	intStream := make(chan int)
	go func() {
		defer close(intStream)
		for v := range valueStream {
			select {
			case <-done:
				return
			case intStream <- v.(int):
			}
		}
	}()
	return intStream
}

func repeat(
	done <-chan interface{},
	values ...interface{},
) <-chan interface{} {
	valueStream := make(chan interface{})
	go func() {
		defer close(valueStream)
		for {
			for _, v := range values {
				select {
				case <-done:
					return
				case valueStream <- v:
				}
			}
		}
	}()
	return valueStream
}

func take(
	done <-chan interface{},
	valueStream <-chan int,
	num int,
) <-chan interface{} {
	takeStream := make(chan interface{})
	go func() {
		defer close(takeStream)
		for i := 0; i < num; i++ {
			select {
			case <-done:
				return
			case takeStream <- <-valueStream:
			}
		}
	}()
	return takeStream
}

func primeFinder(
	done <-chan interface{},
	intStream <-chan int,
) <-chan int {
	valueStream := make(chan int)
	go func() {
		defer close(valueStream)
		for {
			select {
			case <-done:
				return
			case val := <-intStream:
				if isPrime(val) {
					valueStream <- val
				}
			}
		}
	}()
	return valueStream
}

func isPrime(num int) bool {
	for i := 2; i < num; i++ {
		if num%i == 0 {
			return false
		}
	}
	return true
}

func funIn(
	done <-chan interface{},
	channels ...<-chan int,
) <-chan int {
	var wg sync.WaitGroup
	multiplexedStream := make(chan int)

	multiplex := func(c <-chan int) {
		defer wg.Done()
		for i := range c {
			select {
			case <-done:
				return
			case multiplexedStream <- i:
			}
		}
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}

	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()

	return multiplexedStream
}

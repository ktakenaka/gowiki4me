package main

import (
	"fmt"
	"time"
)

func main() {
	done := make(chan interface{})

	go func() {
		//　擬似的にタイムアウト処理
		<-time.After(3 * time.Second)
		close(done)
	}()

	valueStream := valueGenerator(done)
	resultStream := reallyLongCalculation(done, valueStream)

	for result := range resultStream {
		if result == nil {
			break
		}
		fmt.Println(result)
	}
}

func valueGenerator(done <-chan interface{}) <-chan interface{} {
	valueStream := make(chan interface{})
	go func() {
		defer close(valueStream)
		for i := 0; i < 5; i++ {
			select {
			case <-done:
				return
			case valueStream <- i:
			}
		}
	}()
	return valueStream
}

func reallyLongCalculation(done <-chan interface{}, valueStream <-chan interface{}) <-chan interface{} {
	resultStream := make(chan interface{})
	go func() {
		defer close(resultStream)
		for {
			select {
			case <-done:
				return
			case val := <-valueStream:
				resultStream <- longCalculation(done, val)
			}
		}
	}()
	return resultStream
}

func longCalculation(done <-chan interface{}, value interface{}) interface{} {
	valueStream := make(chan interface{})
	go func() {
		defer close(valueStream)
		// 長い計算の模倣コード
		<-time.After(1 * time.Second)
		valueStream <- value
	}()
	select {
	case <-done:
		return nil
	case result := <-valueStream:
		return result
	}
}

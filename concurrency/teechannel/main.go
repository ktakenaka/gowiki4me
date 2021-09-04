package main

import "fmt"

// 一つの処理から上がってきた同じ値を複数のチャネルで使いたいときのパターン
// teeで二つのチャネルに入れるロジックを隠蔽している
func main() {
	done := make(chan interface{})

	in := make(chan interface{})
	defer close(in)

	go func() {
		for i := 0; i < 10; i++ {
			in <- i
		}
		close(done)
	}()

	out1, out2 := tee(done, in)
	for val1 := range out1 {
		fmt.Printf("out1: %v, out2: %v\n", val1, <-out2)
	}
}

func tee(
	done <-chan interface{},
	in <-chan interface{},
) (_, _ <-chan interface{}) {
	out1 := make(chan interface{})
	out2 := make(chan interface{})
	go func() {
		defer close(out1)
		defer close(out2)
		for val := range orDone(done, in) {
			var out1, out2 = out1, out2
			for i := 0; i < 2; i++ {
				select {
				case out1 <- val:
					out1 = nil
				case out2 <- val:
					out2 = nil
				}
			}
		}
	}()
	return out1, out2
}

func orDone(done, c <-chan interface{}) <-chan interface{} {
	valStream := make(chan interface{})
	go func() {
		defer close(valStream)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if ok == false {
					return
				}
				select {
				case valStream <- v:
				case <-done:
				}
			}
		}
	}()
	return valStream
}

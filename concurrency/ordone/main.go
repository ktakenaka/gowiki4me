package main

// capsulate done channel handling
func main() {
	done := make(chan interface{})

	myChan := make(chan interface{})
	defer close(myChan)

	go func() {
		for i := 0; i < 10; i++ {
			myChan <- i
		}
		close(done)
	}()

	for val := range orDone(done, myChan) {
		println(val.(int))
	}
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

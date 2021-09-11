package main

import "fmt"

// レキシカル拘束にすることで、規約による拘束よりも強力な拘束をしている
func main() {
	adhoc()
	lexical()
}

func adhoc() {
	// この関数内では、”data"はloopDataからしか変更されないことを規約により拘束している -> adhoc拘束
	// チームが大きくなるにつれて規約は破られやすくなる（dataをloopData以外の場所で更新される可能性が高くなる）
	data := make([]int, 4)

	loopData := func(handleData chan<- int) {
		defer close(handleData)
		for i := range data {
			handleData <- data[i]
		}
	}

	handleData := make(chan int)
	go loopData(handleData)

	for num := range handleData {
		fmt.Println(num)
	}
}

func lexical() {
	//　レキシカルスコープを使って適切なデータと並行処理のプリミティブだけを複数の並行プロセスが使えるように公開する -> レキシカル拘束
	//　誤った処理を書くことを不可能にしている
	chanOwner := func() <-chan int {
		results := make(chan int, 5)
		go func() {
			defer close(results)
			for i := 0; i <= 5; i++ {
				results <- i
			}
		}()
		// 受信専用のチャネルとして返すことで、resultsが外部で変更されることを防いでいる
		return results
	}

	consumer := func(results <-chan int) {
		for result := range results {
			fmt.Printf("Received: %d\n", result)
		}
		fmt.Println("done receiving!")
	}

	results := chanOwner()
	consumer(results)
}

package main

import (
	"fmt"
	"net/http"
)

type Result struct {
	Error    error
	Response *http.Response
}

// goroutineがエラーを生成するときは、正常系のハンドリング経路に（通常の同期関数のように）エラー処理を入れる
// 下の場合は、外部リクエストでエラーの可能性を考慮して、Result構造体にerrorを入れれるようにしている
func main() {
	checkStatus := func(done <-chan interface{}, urls ...string) <-chan Result {
		results := make(chan Result)
		go func() {
			defer close(results)

			for _, url := range urls {
				var result Result
				resp, err := http.Get(url)
				result = Result{Error: err, Response: resp}
				select {
				case <-done:
					return
				case results <- result:
				}
			}
		}()
		return results
	}

	done := make(chan interface{})
	defer close(done)

	errCount := 0
	urls := []string{"a", "https://www.google.com", "b", "c", "d"}
	for result := range checkStatus(done, urls...) {
		if result.Error != nil {
			fmt.Printf("error: %v\n", result.Error)
			errCount++
			if errCount >= 3 {
				fmt.Println("Too many errors, breaking!")
				break
			}
			continue
		}
		fmt.Printf("Response: %v\n", result.Response.Status)
	}
}

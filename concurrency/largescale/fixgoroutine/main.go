package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// steward: goroutineの健全性を監視するロジック管理人。wardにいるgoroutineが不健全になったら再起動する。
// ward: 管理人が監視するgoroutine
func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	done := make(chan interface{})
	defer close(done)

	// 開始するwardと、outputのstreamを先に作り、stewardの監視のもと動かす
	doWork, intStream := doWorkFn(done, 1, 2, -1, 3, 4, 5)
	doWorkWithSteward := newSteward(1*time.Millisecond, doWork)
	doWorkWithSteward(done, 1*time.Hour)

	for intVal := range take(done, intStream, 6) {
		fmt.Printf("Received: %v\n", intVal)
	}
}

type startGoroutineFn func(
	done <-chan interface{},
	pulseInterval time.Duration,
) (heartbeat <-chan interface{}) // stewardが監視するために、wardのheartbeatを返す

// stewardの実装
func newSteward(
	timeout time.Duration,
	startGoroutine startGoroutineFn,
) startGoroutineFn {
	return func(
		done <-chan interface{},
		pulseInterval time.Duration,
	) <-chan interface{} {
		heartbeat := make(chan interface{})

		go func() {
			defer close(heartbeat)

			var wardDone chan interface{}
			var wardHeartbeat <-chan interface{}

			startWard := func() {
				wardDone = make(chan interface{})
				wardHeartbeat = startGoroutine(or(wardDone, done), timeout/2)
			}
			startWard()
			pulse := time.Tick(pulseInterval)

		monitorLoop:
			for {
				timeoutSignal := time.After(timeout)

				for {
					// stewardも外部にheartbeatを返す
					select {
					case <-pulse:
						select {
						case heartbeat <- struct{}{}:
						default:
						}
					case <-wardHeartbeat:
						// wardからheartbeatを受け取ったら継続してモニターする
						continue monitorLoop
					case <-timeoutSignal:
						// wardから応答がなかったら、restartする
						log.Println("steward: ward unhealthy; restarting")
						startWard()
						continue monitorLoop
					case <-done:
						return
					}
				}
			}
		}()
		return heartbeat
	}
}

func or(
	done,
	c <-chan interface{},
) <-chan interface{} {
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

func bridge(
	done <-chan interface{},
	chanStraem <-chan <-chan interface{},
) <-chan interface{} {
	valStream := make(chan interface{})
	go func() {
		defer close(valStream)
		for {
			var stream <-chan interface{}
			select {
			case maybeStream, ok := <-chanStraem:
				if !ok {
					return
				}
				stream = maybeStream
			case <-done:
				return
			}
			for val := range or(done, stream) {
				select {
				case valStream <- val:
				case <-done:
				}
			}
		}
	}()
	return valStream
}

func take(
	done <-chan interface{},
	valueStream <-chan interface{},
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

// Old implementation
// func doWork(done <-chan interface{}, _ time.Duration) <-chan interface{} {
// 	log.Println("Ward: Hello, I'm irresponsible")
// 	go func() {
// 		<-done
// 		log.Println("ward: I am halting.")
// 	}()
// 	return nil
// }

func doWorkFn(
	done <-chan interface{},
	intList ...int,
) (startGoroutineFn, <-chan interface{}) {
	intChanStream := make(chan (<-chan interface{})) // bridge patternのため
	intStream := bridge(done, intChanStream)
	doWork := func(
		done <-chan interface{},
		pulseInterval time.Duration,
	) <-chan interface{} {
		intStream := make(chan interface{})
		heartbeat := make(chan interface{})
		go func() {
			defer close(intStream)
			select {
			case intChanStream <- intStream:
			case <-done:
				return
			}

			pulse := time.Tick(pulseInterval)

			for {
			valueLoop:
				for _, intVal := range intList {
					if intVal < 0 {
						log.Printf("negative value: %v\n", intVal)
						return
					}

					for {
						select {
						case <-pulse:
							select {
							case heartbeat <- struct{}{}:
							default:
							}
						case intStream <- intVal:
							continue valueLoop
						case <-done:
							return
						}
					}
				}
			}
		}()
		return heartbeat
	}
	return doWork, intStream
}

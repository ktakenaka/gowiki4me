package main

import (
	"context"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	defer log.Println("Done.")

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	apiConnection := Open()
	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ReadFile(context.Background())
			if err != nil {
				log.Printf("cannot ReadFile: %v\n", err)
			}
			log.Printf("Readfile")
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ResolveAddress(context.Background())
			if err != nil {
				log.Printf("cannot ResolveAddress: %v\n", err)
			}
			log.Printf("ResolveAddress")
		}()
	}

	wg.Wait()
}

type APIConnection struct {
	// こんなふうに同じ型ものは並べて書けるのか。地味に知らなかった...
	networkLimit,
	diskLimit,
	apiLimit rateLimit
}

func Open() *APIConnection {
	// 外部API、データベースアクセス、ネットワークアクセスなど実際のユースケースではいろいろな場面で流量制限をかけたくなる。
	// さらに、一つ一つの流量制限の対象の中にも、1秒あたりの基準と1分あたりの基準など、尺度を変えていくつか持ちたくなる。
	// 流量制限はクライアントでも、サーバー側でもかけるようにする
	return &APIConnection{
		apiLimit: MultiLimiter(
			rate.NewLimiter(Per(2, time.Second), 2),
			rate.NewLimiter(Per(10, time.Minute), 10),
		),
		diskLimit: MultiLimiter(
			rate.NewLimiter(rate.Limit(1), 1),
		),
		networkLimit: MultiLimiter(
			rate.NewLimiter(Per(3, time.Second), 3),
		),
	}
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	err := MultiLimiter(a.apiLimit, a.diskLimit).Wait(ctx)
	if err != nil {
		return err
	}
	// Pretend we do work here
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	err := MultiLimiter(a.apiLimit, a.networkLimit).Wait(ctx)
	if err != nil {
		return err
	}
	// Pretend we do work here
	return nil
}

type rateLimit interface {
	Wait(context.Context) error
	Limit() rate.Limit
}

func MultiLimiter(limiters ...rateLimit) *multiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit)
	return &multiLimiter{limiters: limiters}
}

type multiLimiter struct {
	limiters []rateLimit
}

func (l *multiLimiter) Wait(ctx context.Context) error {
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *multiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit()
}

func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every((duration / time.Duration(eventCount)))
}

// +build !local

package main

import (
	"sync"
	"time"

	"github.com/brownsys/tracing-framework-go/xtrace/client"
	"golang.org/x/net/context"
)

func main() {
	c := context.Background()

	client.Log("1")
	client.Log("2")
	client.Log("3")

	var wg sync.WaitGroup
	wg.Add(1)
	go func(c context.Context) {
		client.Log("4")
		wg.Done()
	}(c)

	client.Log("5")
	wg.Wait()

	time.Sleep(time.Hour)
}

// +build local

package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/brownsys/tracing-framework-go/local"
	"github.com/brownsys/tracing-framework-go/xtrace/client"
)

func main() {
	err := client.Connect("localhost:5563")
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to X-Trace server: %v\n", err)
		os.Exit(1)
	}

	client.Log("1")
	client.Log("2")
	client.Log("3")

	var wg sync.WaitGroup
	wg.Add(1)
	go func(__f1 func(), __f2 func()) {
		__f1()
		__f2()
	}(local.GetSpawnCallback(), func() {
		client.Log("4")

		wg.Done()
	})

	client.Log("5")
	wg.Wait()

	time.Sleep(time.Hour)
}

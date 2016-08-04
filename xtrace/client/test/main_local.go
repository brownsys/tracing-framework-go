// +build local

package main

import (
	__runtime "runtime"
	"sync"
	"time"

	"github.com/brownsys/tracing-framework-go/local"
	"github.com/brownsys/tracing-framework-go/xtrace/client"
	"golang.org/x/net/context"
)

func main() {
	c := context.Background()
	__runtime.SetLocal(c)

	client.Log("1")
	client.Log("2")
	client.Log("3")

	var wg sync.WaitGroup
	wg.Add(1)
	go func(__f1 func(), __f2 func(c context.
		Context), arg0 context.Context) {
		__f1()
		__f2(arg0)
	}(local.GetSpawnCallback(), func(c context.Context) {
		client.
			Log("4")
		wg.Done()
	}, c)

	client.Log("5")
	wg.Wait()

	time.Sleep(time.Hour)
}

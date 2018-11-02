package scheduler

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"sync"
)

type RetryServiceFunc func(timeout int, wait *sync.WaitGroup, ch chan string)

type LogFunc func(msg string)

func CheckStatus(params startup.BootParams, ServiceUp RetryServiceFunc, log LogFunc) {
	deps := make(chan string, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go ServiceUp(params.BootTimeout, &wg, deps)
	go func(ch chan string) {
		for {
			select {
			case m, ok := <-ch:
				if ok {
					log(m)
				} else {
					return
				}
			}
		}
	}(deps)

	wg.Wait()
}
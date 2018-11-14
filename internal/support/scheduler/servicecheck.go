package scheduler

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"sync"
)

type RetryServiceFunc func(timeout int, wait *sync.WaitGroup, ch chan string)

func CheckStatus(params startup.BootParams, ServiceUp RetryServiceFunc, ch chan string) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go ServiceUp(params.BootTimeout, &wg, ch)
	wg.Wait()
}
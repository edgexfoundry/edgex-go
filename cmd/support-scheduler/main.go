package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func main() {

	start := time.Now()
	var useRegistry bool
	var useProfile string

	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use Registry.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use Registry.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	params := startup.BootParams{UseRegistry: useRegistry, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, scheduler.Retry, logBeforeInit)

	ok := scheduler.Init(useRegistry)
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", clients.SupportSchedulerServiceKey))
		os.Exit(1)
	}

	scheduler.LoggingClient.Info(fmt.Sprintf("Service dependencies resolved...%s %s ", clients.SupportSchedulerServiceKey, edgex.Version))
	scheduler.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", clients.SupportSchedulerServiceKey, edgex.Version))

	// Bootstrap schedulers
	err := scheduler.LoadScheduler()
	if err != nil {
		scheduler.LoggingClient.Error(fmt.Sprintf("Failed to load schedules and events %s", err.Error()))
	}

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(scheduler.Configuration.Service.Timeout), "Request timed out")
	scheduler.LoggingClient.Info(scheduler.Configuration.Service.StartupMsg)

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	url := scheduler.Configuration.Service.Host + ":" + strconv.Itoa(scheduler.Configuration.Service.Port)
	startup.StartHTTPServer(scheduler.LoggingClient, scheduler.Configuration.Service.Timeout, scheduler.LoadRestRoutes(), url, errs)

	// Start the ticker
	scheduler.StartTicker()

	// Time it took to start service
	scheduler.LoggingClient.Info("Service started in: " + time.Since(start).String())
	scheduler.LoggingClient.Info("Listening on port: " + strconv.Itoa(scheduler.Configuration.Service.Port))
	c := <-errs
	scheduler.Destruct()
	scheduler.LoggingClient.Warn(fmt.Sprintf("terminating: %v", c))

	os.Exit(0)
}

func logBeforeInit(err error) {
	if scheduler.LoggingClient == nil {
		scheduler.LoggingClient = logger.NewClient(clients.SupportSchedulerServiceKey, false, "", models.InfoLog)
	}

	scheduler.LoggingClient.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}

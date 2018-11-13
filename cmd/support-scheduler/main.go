package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/gorilla/context"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler"
)

func main() {

	start := time.Now()
	var useConsul bool
	var useProfile string

	flag.BoolVar(&useConsul, "consul", false, "Indicates the service should use consul.")
	flag.BoolVar(&useConsul, "c", false, "Indicates the service should use consul.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	params := startup.BootParams{UseConsul: useConsul, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, scheduler.Retry, logBeforeInit)

	ok := scheduler.Init(useConsul)
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", internal.SupportSchedulerServiceKey))
		os.Exit(1)
	}

	scheduler.LoggingClient.Info(fmt.Sprintf("Service dependencies resolved...%s %s ", internal.SupportSchedulerServiceKey, edgex.Version))
	scheduler.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.SupportSchedulerServiceKey, edgex.Version))

	// Bootstrap schedulers
	err := scheduler.LoadScheduler()
	if err != nil {
		scheduler.LoggingClient.Error(fmt.Sprintf("Failed to load schedules and events %s", err.Error()))
	}

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(scheduler.Configuration.Service.Timeout), "Request timed out")
	scheduler.LoggingClient.Info(scheduler.Configuration.Service.StartupMsg)

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	startHttpServer(errs, scheduler.Configuration.Service.Port)

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
	scheduler.LoggingClient = logger.NewClient(internal.SupportSchedulerServiceKey, false, "", logger.InfoLog)
	scheduler.LoggingClient.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}

func startHttpServer(errChan chan error, port int) {
	go func() {
		correlation.LoggingClient = scheduler.LoggingClient
		r := scheduler.LoadRestRoutes()
		errChan <- http.ListenAndServe(":"+strconv.Itoa(port), context.ClearHandler(r))
	}()
}

package main

import (
	"flag"
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/core/data"

	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler"
	client "github.com/edgexfoundry/edgex-go/pkg/clients/scheduler"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"

)

var loggingClient logger.LoggingClient

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

	//Read Configuration.
	configuration := &scheduler.ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		logBeforeTermination(err)
		return
	}

	//Determine if configuration should be overridden from Consul
	var consulMsg string
	if useConsul {
		consulMsg = "Loading configuration from Consul..."
		err := scheduler.ConnectToConsul(*configuration)
		if err != nil {
			logBeforeTermination(err)
			return //end program since user explicitly told us to use Consul.
		}
	} else {
		consulMsg = "Bypassing Consul configuration..."
	}

	// Setup Logging
	logTarget := setLoggingTarget(*configuration)
	loggingClient = logger.NewClient(internal.SupportSchedulerServiceKey, configuration.EnableRemoteLogging, logTarget)

	loggingClient.Info(consulMsg)
	loggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.SupportSchedulerServiceKey, edgex.Version))


	client.SetConfiguration(configuration.ConsulHost,configuration.ConsulPort)
	var schedulerClient = client.GetSchedulerClient()

	err = scheduler.Init(*configuration, schedulerClient, loggingClient, useConsul)
	if err != nil {
		loggingClient.Error(fmt.Sprintf("call to init() failed: %v", err.Error()))
		return
	}

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration.ServiceTimeout), "Request timed out")
	loggingClient.Info(configuration.AppOpenMsg, "")

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	startHttpServer(errs, configuration.ServicePort)

	// Time it took to start service
	loggingClient.Info("Service started in: "+time.Since(start).String(), "")
	loggingClient.Info("Listening on port: " + strconv.Itoa(configuration.ServicePort))
	c := <-errs
	data.Destruct()
	loggingClient.Warn(fmt.Sprintf("terminating: %v", c))

	// Start the Scheduler Service
	r := scheduler.LoadRestRoutes()
	http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration.ServiceTimeout), "Request timed out")
	loggingClient.Info(configuration.AppOpenMsg, "")


	scheduler.StartTicker()

	// Time it took to start service
	loggingClient.Info("service started in: "+time.Since(start).String(), "")
	loggingClient.Info("listening on port: "+strconv.Itoa(configuration.ServicePort), "")
	loggingClient.Error(http.ListenAndServe(":"+strconv.Itoa(configuration.ServicePort), r).Error())

	scheduler.StopTicker()
}

func logBeforeTermination(err error) {
	loggingClient = logger.NewClient(internal.SupportSchedulerServiceKey, false, "")
	loggingClient.Error(err.Error())
}

func setLoggingTarget(conf scheduler.ConfigurationStruct) string {
	logTarget := conf.LoggingRemoteURL
	if !conf.EnableRemoteLogging {
		return conf.LogFile
	}
	return logTarget
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
		r := data.LoadRestRoutes()
		errChan <- http.ListenAndServe(":"+strconv.Itoa(port), r)
	}()
}
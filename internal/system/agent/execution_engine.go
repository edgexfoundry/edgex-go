package agent

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/logger"
	"os/exec"
)

type ExecuteApp struct {
}

func runExec(service string, operation string) error {

	// Preparing the call to the executor app.
	cmd := exec.Command(Configuration.ExecutorPath, service, operation)

	cmd.Dir = Configuration.ExecutorPath

	_, err := cmd.CombinedOutput()
	if err != nil {
		logs.LoggingClient.Error(fmt.Sprintf("the error occurred in the invocation of executor: %v ", err.Error()))
	} else {
		logs.LoggingClient.Info("invocation of executor succeeded")
	}

	return err
}

func (ec *ExecuteApp) Start(service string) error {

	err := runExec(service, "start")
	if err != nil {
		logs.LoggingClient.Error(fmt.Sprintf("there was an error in starting service %s: %s", service, err))
	} else {
		logs.LoggingClient.Debug(fmt.Sprintf("starting service {%s} succeeded", service))
	}
	return err
}

func (ec *ExecuteApp) Stop(service string) error {

	err := runExec(service, "stop")
	if err != nil {
		logs.LoggingClient.Error(fmt.Sprintf("there was an error in stopping service %s: %s", service, err))
	} else {
		logs.LoggingClient.Debug(fmt.Sprintf("stopping service {%s} succeeded", service))
	}
	return err
}

func (ec *ExecuteApp) Restart(service string) error {

	err := runExec(service, "restart")
	if err != nil {
		logs.LoggingClient.Error(fmt.Sprintf("there was an error in restarting service %s: %s", service, err))
	} else {
		logs.LoggingClient.Debug(fmt.Sprintf("restarting service {%s} succeeded", service))
	}
	return err
}

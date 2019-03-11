package agent

import (
	"fmt"
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
		LoggingClient.Error(fmt.Sprintf("an error occurred in calling executor on service %s where requested operation was %s: %v ", service, operation, err.Error()))
	} else {
		LoggingClient.Info("invocation of executor succeeded")
	}

	return err
}

func (ec *ExecuteApp) Start(service string) error {

	err := runExec(service, "start")
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error in starting service %s: %v", service, err.Error()))
	} else {
		LoggingClient.Debug(fmt.Sprintf("starting service %s succeeded", service))
	}
	return err
}

func (ec *ExecuteApp) Stop(service string) error {

	err := runExec(service, "stop")
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error in stopping service %s: %v", service, err.Error()))
	} else {
		LoggingClient.Debug(fmt.Sprintf("stopping service %s succeeded", service))
	}
	return err
}

func (ec *ExecuteApp) Restart(service string) error {

	err := runExec(service, "restart")
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error in restarting service %s: %v", service, err.Error()))
	} else {
		LoggingClient.Debug(fmt.Sprintf("restarting service %s succeeded", service))
	}
	return err
}

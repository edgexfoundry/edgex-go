package agent

import (
	"context"
	"fmt"
	"os/exec"
)

type ExecuteApp struct {
}

func runExec(service string, operation string) error {

	// Preparing the call to the executor app.
	cmd := exec.Command(Configuration.ExecutorPath, service, operation)

	_, err := cmd.CombinedOutput()
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("an error occurred in calling executor on service %s where requested operation was %s: %v ", service, operation, err.Error()))
	} else {
		LoggingClient.Info("invocation of executor succeeded")
	}

	return err
}

func execMetrics(service string) ([]byte, error) {

	cmd := exec.Command(Configuration.ExecutorPath, service, METRICS)

	out, err := cmd.CombinedOutput()
	if err.Error() != "exit status 1" {
		LoggingClient.Error(fmt.Sprintf("an error occurred in calling executor (to fetch metrics) for service %s: %v ", service, err.Error()))
	} else {
		LoggingClient.Debug("invocation of execMetrics() succeeded")
	}
	return out, err
}

func (ea *ExecuteApp) Start(service string) error {

	err := runExec(service, "start")
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error in starting service %s: %v", service, err.Error()))
	} else {
		LoggingClient.Debug(fmt.Sprintf("starting service %s succeeded", service))
	}
	return err
}

func (ea *ExecuteApp) Stop(service string) error {

	err := runExec(service, "stop")
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error in stopping service %s: %v", service, err.Error()))
	} else {
		LoggingClient.Debug(fmt.Sprintf("stopping service %s succeeded", service))
	}
	return err
}

func (ea *ExecuteApp) Restart(service string) error {

	err := runExec(service, "restart")
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error in restarting service %s: %v", service, err.Error()))
	} else {
		LoggingClient.Debug(fmt.Sprintf("restarting service %s succeeded", service))
	}
	return err
}

func (ea *ExecuteApp) Metrics(ctx context.Context, service string) ([]byte, error) {

	out, err := execMetrics(service)
	if err.Error() != "exit status 1" {
		LoggingClient.Error(fmt.Sprintf("error in fetching metrics %s: %v", service, err.Error()))
	} else {
		LoggingClient.Debug(fmt.Sprintf("fetching metrics %s succeeded", service))
	}
	return out, err
}

package executor

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/system/agent/logger"
)

type ExecuteOs struct {
}

func (oe *ExecuteOs) StopService(service string) error {

	cmd := exec.Command("ps", "-ax")
	out, err := cmd.CombinedOutput()
	if err != nil {
		logs.LoggingClient.Error("StopService() failed", "error message", err.Error())
		return err
	}
	findAndStopProcess(string(out), err, service)

	return nil
}

func findAndStopProcess(output string, err error, process string) error {

	var pid int
	if err != nil {
		logs.LoggingClient.Error("findAndStopProcess() failed", "error message", err.Error())
		return nil
	}

	// We are only interested in that segment of the output (from listing the running processes) which has this pattern:
	// <PID> ttys###    H:MM.SS <process-name>
	// For example the following:
	// 19922 ttys010    0:01.50 edgex-core-metadata
	if strings.Contains(output, process) {
		// Find the PID of the process which we seek to stop.
		for _, line := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {
			if strings.Contains(line, process) {
				tokens := strings.Split(line, " ")
				pid, err = strconv.Atoi(tokens[0])
				if err != nil {
					logs.LoggingClient.Error("error converting PID to integer", "error message", err.Error(), "unparsable token", tokens[0])
				}
				logs.LoggingClient.Debug("found service to be stopped", "PID", tokens[0], "service name", process)
			}
		}

		// Now stop the process using the PID found above.
		// Make a system call.
		proc, err := os.FindProcess(pid)
		if err != nil {
			logs.LoggingClient.Error("os.FindProcess(pid) failed", "error message", err.Error())
		}
		proc.Kill()
	} else {
		// TODO Return suitable response...
		logs.LoggingClient.Debug("service not running", "level", "OS", "service name", process)
	}

	return nil
}

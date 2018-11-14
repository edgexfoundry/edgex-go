package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ExecuteOs implements ServiceStopper to be able to stop native services
// by killing the processes
type ExecuteOs struct {
}

// Stop will stop the service by searching for the PID and killing it directly
func (oe *ExecuteOs) Stop(service string, params []string) error {

	cmd := exec.Command("ps", "-ax")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	// find the process's pid from the process listing returned by ps
	pid, err := parseProcessListing(string(out), service)
	if err != nil {
		return err
	}

	// Now stop the process using the PID found above.
	// Make a system call.
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

// parseProcessListing searches for a process name in the output
// and returns the pid if the processes exists in the output
func parseProcessListing(output string, processName string) (int, error) {
	// We are only interested in that segment of the output (from listing the running processes) which has this pattern:
	// <PID> ttys###    H:MM.SS <process-name>
	// For example the following:
	// 19922 ttys010    0:01.50 edgex-core-metadata
	if strings.Contains(output, processName) {
		// Find the PID of the process which we seek to stop.
		for _, line := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {
			if strings.Contains(line, processName) {
				tokens := strings.Split(line, " ")
				pid, err := strconv.Atoi(tokens[0])
				if err != nil {
					return 0, err
				}
				return pid, nil
			}
		}
	}
	return 0, fmt.Errorf("failed to find the process \"%s\"", processName)
}

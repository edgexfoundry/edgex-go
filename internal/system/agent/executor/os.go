package executor

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type ExecuteOs struct {
}

func (oe *ExecuteOs) StopService(service string) error {

	cmd := exec.Command("ps", "-ax")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// fmt.Sprintf("Call to StopService() failed with %s\n", err)
		return err
	}
	findAndStopProcess(string(out), err, service)

	return nil
}

func findAndStopProcess(output string, err error, process string) error {

	var pid int
	if err != nil {
		log.Printf("Error during findAndStopProcess() as follows: %s" + err.Error())
		// agent.LoggingClient.Error(" Error during findAndStopProcess() as follows: ", err.Error())
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
					log.Printf("Error converting PID to integer: %s" + err.Error())
					// log.Fatalf("Error converting PID to integer: " + err.Error())
				}
				// fmt.Sprintf("Found the data with the PID {%v} of the the micro-service {%v} which we seek to stop!", tokens[0], process)
			}
		}

		// Now stop the process using the PID found above.
		// Make a system call.
		proc, err := os.FindProcess(pid)
		if err != nil {
			log.Printf("Error during os.FindProcess(pid) as follows: %s" + err.Error())
			// agent.LoggingClient.Error(" Error during os.FindProcess(pid) as follows: ", err.Error())
		}
		proc.Kill()
	} else {
		// TODO Return suitable response...
		// fmt.Sprintf("OS-level >> The micro-service {%v} was NOT found in a running state...", process)
	}

	return nil
}

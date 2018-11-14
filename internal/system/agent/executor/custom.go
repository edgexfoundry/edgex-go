package executor

import (
	"os"
	"os/exec"
)

// CustomProgram is a custom implementation of managing services with the SMA
// that is specified through the configuration.toml file for sys-mgmt-agent
type CustomProgram struct {
	Program string
}

// Stop will stop a service calling the custom program with args "program stop SERVICE"
func (cp *CustomProgram) Stop(service string, params []string) error {
	_, err := cp.runProgram(service, "stop")
	return err
}

func (cp *CustomProgram) runProgram(service, op string) ([]byte, error) {
	// make sure the file exists first assuming the program is an absolute path
	file := cp.Program
	if _, err := os.Stat(cp.Program); os.IsNotExist(err) {
		// doesn't exist, try looking on the $PATH
		file, err = exec.LookPath(cp.Program)
		if err != nil {
			// couldn't find the program either as an absolute file
			// or as a file on the $PATH
			return nil, err
		}
	}

	// run the command
	return exec.Command(file, op, service).CombinedOutput()
}

// Start will start a service calling the custom program with args "program start SERVICE"
func (cp *CustomProgram) Start(service string, params []string) error {
	_, err := cp.runProgram(service, "start")
	return err
}

// Restart will restart a service calling the custom program with args "program restart SERVICE"
func (cp *CustomProgram) Restart(service string, params []string) error {
	_, err := cp.runProgram(service, "restart")
	return err
}

// Disable will disable a service calling the custom program with args "program disable SERVICE"
func (cp *CustomProgram) Disable(service string, params []string) error {
	_, err := cp.runProgram(service, "disable")
	return err
}

// Enable will enable a service calling the custom program with args "program enable SERVICE"
func (cp *CustomProgram) Enable(service string, params []string) error {
	_, err := cp.runProgram(service, "enable")
	return err
}

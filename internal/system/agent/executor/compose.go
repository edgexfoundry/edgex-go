package executor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/logger"
)
var services = map[string]string{
internal.SupportNotificationsServiceKey: "notifications",
internal.CoreDataServiceKey: "data",
internal.CoreMetaDataServiceKey: "metadata",
internal.CoreCommandServiceKey: "command",
internal.ExportClientServiceKey: "export-client",
internal.ExportDistroServiceKey: "export-distro",
internal.SupportLoggingServiceKey: "logging",
internal.ConfigSeedServiceKey: "config-seed",
}

func WasDockerContainerComposeStarted(service string) bool {

	var (
		cmdOut []byte
		err    error
	)
	cmdName := "docker"
	cmdArgs := []string{"ps"}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		logs.LoggingClient.Error("error running the docker-compose command", "error message", err.Error())
		os.Exit(1)
	}
	composeOutput := string(cmdOut)
	// Find whether the container to start has started.
	for _, line := range strings.Split(strings.TrimSuffix(composeOutput, "\n"), "\n") {
		if strings.Contains(line, service) {

			if strings.Contains(line, "Up") {
				logs.LoggingClient.Debug("container started", "service name", service, "details", line)
				return true
			} else {
				logs.LoggingClient.Warn("container not started", "service name", service)
				return false
			}
		}
	}
	return false
}

func StartDockerContainerCompose(service string, composeUrl string) error {
	_, knownService := services[service]

	if knownService {
		RunDockerComposeCommand(service, services[service], composeUrl)

		return nil
	} else {
		newError := fmt.Errorf("unknown service: %v", service)
		logs.LoggingClient.Error(newError.Error())

		return newError
	}
}

func RunDockerComposeCommand(service string, dockerComposeService string, composeUrl string) {

	var (
		err    error
		cmdDir string
	)
	cmdName := "docker-compose"

	// Retry fetch of the docker-compose.yml from the GitHub repository.
	err = Do(func(attempt int) (bool, error) {
		var err error
		cmdDir, err = FetchDockerComposeYamlAndPath(composeUrl)
		// Try 5 times
		return attempt < 5, err
	})
	if err != nil {
		logs.LoggingClient.Error("unable to pull the latest compose file from repository" ,"error message", err.Error())
	}

	cmdArgs := []string{"up", "-d", dockerComposeService}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = cmdDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		logs.LoggingClient.Warn("docker-compose up -d failed", "error message", err.Error())
	}
	println(out)

	if ! WasDockerContainerComposeStarted(service) {
		logs.LoggingClient.Warn("container not started", "service name",  service)
	}
}

func FetchDockerComposeYamlAndPath(composeUrl string) (string, error) {

	// [1] Fetch contents of the latest "docker-compose.yml" file from GitHub.
	req, _ := http.NewRequest("GET", composeUrl, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.LoggingClient.Error("GET failed", "error message", err.Error())
	}

	if res.Body != nil {
		defer res.Body.Close()
	}
	body, _ := ioutil.ReadAll(res.Body)

	// [2] Determine the directory (in the deployed filesystem) that we will be writing the fetched contents to.
	cmdName := "curl"
	cmdName = "pwd"

	cmd := exec.Command(cmdName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logs.LoggingClient.Error("exec.Command(cmdName) failed", "error message", err.Error())
	}
	composeOutput := string(out)

	cmdArgs := composeOutput
	path := strings.TrimSuffix(cmdArgs, "\n")

	composeFile := "/docker-compose.yml"
	filename := path + composeFile

	// [3] Get info about the file "filename".
	fileInfo, err := os.Stat(filename)
	if err != nil {
		logs.LoggingClient.Warn("docker-compose.yml does not exist; creating", "error message", err.Error())
	}

	// [4] Determine whether we have already fetched the contents and written them to the deployed filesystem.
	if os.IsNotExist(err) {
		println(fileInfo)
		err = ioutil.WriteFile(filename, []byte(body), 0666)
		if err != nil {
			logs.LoggingClient.Error("already fetched the contents and written them to the deployed file-system", "error message", err.Error())
		}
	}
	return path, err
}

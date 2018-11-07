package executor

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
)

func WasDockerContainerComposeStarted(service string) bool {

	var (
		cmdOut []byte
		err    error
	)
	cmdName := "docker"
	cmdArgs := []string{"ps"}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running the docker-compose command: ", err.Error())
		os.Exit(1)
	}
	composeOutput := string(cmdOut)
	// Find whether the container which we sought to start has started.
	for _, line := range strings.Split(strings.TrimSuffix(composeOutput, "\n"), "\n") {
		if strings.Contains(line, service) {

			if strings.Contains(line, "Up") {
				// fmt.Sprintf("The container for {%v} has started! Some (container) details as follows:\n{%v}", service, line)
				return true
			} else {
				// fmt.Sprintf("The container for {%v} has NOT started!" + service)
				return false
			}
		}
	}
	return false
}

func StartDockerContainerCompose(service string, composeUrl string) error {

	var dockerComposeService string

	switch service {

	case internal.SupportNotificationsServiceKey:
		dockerComposeService = "notifications"
		RunDockerComposeCommand(service, dockerComposeService, composeUrl)
		break

	case internal.CoreDataServiceKey:
		dockerComposeService = "data"
		RunDockerComposeCommand(service, dockerComposeService, composeUrl)
		break

	case internal.CoreMetaDataServiceKey:
		dockerComposeService = "metadata"
		RunDockerComposeCommand(service, dockerComposeService, composeUrl)
		break

	case internal.CoreCommandServiceKey:
		dockerComposeService = "command"
		RunDockerComposeCommand(service, dockerComposeService, composeUrl)
		break

	case internal.ExportClientServiceKey:
		dockerComposeService = "export-client"
		RunDockerComposeCommand(service, dockerComposeService, composeUrl)
		break

	case internal.ExportDistroServiceKey:
		dockerComposeService = "export-distro"
		RunDockerComposeCommand(service, dockerComposeService, composeUrl)
		break

	case internal.SupportLoggingServiceKey:
		dockerComposeService = "logging"
		RunDockerComposeCommand(service, dockerComposeService, composeUrl)
		break

	case internal.ConfigSeedServiceKey:
		dockerComposeService = "config-seed"
		RunDockerComposeCommand(service, dockerComposeService, composeUrl)
		break

	default:
		break
	}
	return nil
}

func RunDockerComposeCommand(service string, dockerComposeService string, composeUrl string) {

	var (
		err    error
		cmdDir string
	)
	cmdName := "docker-compose"

	// Retry the fetching of the docker-compose.yml from the Github edgexfoundry repository.
	err = Do(func(attempt int) (bool, error) {
		var err error
		cmdDir, err = FetchDockerComposeYamlAndPath(composeUrl)
		// Try 5 times
		return attempt < 5, err
	})
	if err != nil {
		log.Printf("Unable to pull the latest compose file from Github edgexfoundry repository: %s", err.Error())
		// agent.LoggingClient.Error("Unable to pull the latest compose file from Github edgexfoundry repository: %s", err.Error())
	}

	cmdArgs := []string{"up", "-d", dockerComposeService}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = cmdDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf(fmt.Sprintf("Call to docker-compose up -d failed with %s\n", err.Error()))
		// agent.LoggingClient.Error("Call to docker-compose up -d failed with %s\n", err.Error())
	}
	println(out)

	if !WasDockerContainerComposeStarted(service) {
		log.Printf(fmt.Sprintf("The container for {%s} was NOT started!" + service))
		// agent.LoggingClient.Warn("The container for {%s} was NOT started!" + service)
	}
}

func FetchDockerComposeYamlAndPath(composeUrl string) (string, error) {

	// [1] Fetch contents of the latest "docker-compose.yml" file from Github.
	req, _ := http.NewRequest("GET", composeUrl, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf(fmt.Sprintf("Call to http.Get() failed with %s\n", err.Error()))
		// agent.LoggingClient.Error("Call to http.Get() failed with %s\n", err.Error())
	}
	body, _ := ioutil.ReadAll(res.Body)

	// [2] Determine the directory (in the deployed file-system) that we will be writing the fetched contents to.
	cmdName := "curl"
	cmdName = "pwd"

	cmd := exec.Command(cmdName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf(fmt.Sprintf("Call to exec.Command(cmdName) failed with %s\n", err.Error()))
		// agent.LoggingClient.Error("Call to exec.Command(cmdName) failed with %s\n", err.Error())
	}
	composeOutput := string(out)

	cmdArgs := composeOutput
	path := strings.TrimSuffix(cmdArgs, "\n")

	composeFile := "/docker-compose.yml"
	filename := path + composeFile

	// [3] Get info about the named file (i.e. under the designator "filename").
	fileInfo, err := os.Stat(filename)
	if err != nil {
		log.Printf(err.Error(), " File docker-compose.yml does not exist, yet!")
		log.Printf(" Therefore, creating docker-compose.yml from scratch...")
		// agent.LoggingClient.Error(" File docker-compose.yml does not exist, yet!", err.Error())
		// agent.LoggingClient.Info(" Therefore, creating docker-compose.yml from scratch...")
	}

	// [4] Determine whether we have _already_ fetched the contents and written them to the deployed file-system.
	if os.IsNotExist(err) {
		println(fileInfo)
		err = ioutil.WriteFile(filename, []byte(body), 0666)
		if err != nil {
			log.Printf(err.Error(), "%s We have already fetched the contents and written them to the deployed file-system!", err.Error())
			// agent.LoggingClient.Error("%s We have already fetched the contents and written them to the deployed file-system!", err.Error())
		}
	}
	return path, err
}

package executor

import (
	"fmt"
	"strings"
	"os"
	"log"
	"os/exec"
	"net/http"
	"io/ioutil"
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
		fmt.Fprintln(os.Stderr, "There was an error running the docker-compose command: ", err)
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

func StartDockerContainerCompose(service string) error {

	var dockerComposeService string

	switch service {

	case internal.SupportNotificationsServiceKey:
		dockerComposeService = "notifications"
		RunDockerComposeCommand(service, dockerComposeService)
		break

	case internal.CoreDataServiceKey:
		dockerComposeService = "data"
		RunDockerComposeCommand(service, dockerComposeService)
		break

	case internal.CoreMetaDataServiceKey:
		dockerComposeService = "metadata"
		RunDockerComposeCommand(service, dockerComposeService)
		break

	case internal.CoreCommandServiceKey:
		dockerComposeService = "command"
		RunDockerComposeCommand(service, dockerComposeService)
		break

	case internal.ExportClientServiceKey:
		dockerComposeService = "export-client"
		RunDockerComposeCommand(service, dockerComposeService)
		break

	case internal.ExportDistroServiceKey:
		dockerComposeService = "export-distro"
		RunDockerComposeCommand(service, dockerComposeService)
		break

	case internal.SupportLoggingServiceKey:
		dockerComposeService = "logging"
		RunDockerComposeCommand(service, dockerComposeService)
		break

	case internal.ConfigSeedServiceKey:
		dockerComposeService = "config-seed"
		RunDockerComposeCommand(service, dockerComposeService)
		break

	default:
		// fmt.Sprintf(">> Unknown service: %v", service)
		break
	}
	return nil
}

func RunDockerComposeCommand(service string, dockerComposeService string) {

	var (
		err    error
		cmdDir string
	)
	cmdName := "docker-compose"

	// Retry the fetching of the docker-compose.yml from the Github edgexfoundry repository.
	// fmt.Sprintf("Pulling latest compose file from Github edgexfoundry repository...")
	err = Do(func(attempt int) (bool, error) {
		var err error
		cmdDir, err = FetchDockerComposeYamlAndPath()
		// Try 5 times
		return attempt < 5, err
	})
	if err != nil {
		log.Fatalln("Unable to pull the latest compose file from Github edgexfoundry repository.:", err)
	}

	cmdArgs := []string{"up", "-d", dockerComposeService}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = cmdDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(fmt.Sprintf("Call to docker-compose up -d failed with %s\n", err))
	}

	// composeOutput := string(out)
	// fmt.Sprintf("For the micro-service {%v}, we got this docker-compose output:\n {%v}", service, composeOutput)
	println(out)

	if ! WasDockerContainerComposeStarted(service) {
		log.Fatal(fmt.Sprintf("The container for {%v} was NOT started!" + service))
	}
}

func FetchDockerComposeYamlAndPath() (string, error) {

	// Specifying the location of the latest "docker-compose.yml" file (on Github).
	composeUrl := "https://raw.githubusercontent.com/edgexfoundry/developer-scripts/master/compose-files/docker-compose-california-0.6.1.yml"
	req, _ := http.NewRequest("GET", composeUrl, nil)
	res, _ := http.DefaultClient.Do(req)

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, _ := ioutil.ReadAll(res.Body)

	// [1] Fetch contents of the latest "docker-compose.yml" file from Github.
	resp, err := http.Get("https://raw.githubusercontent.com/edgexfoundry/developer-scripts/master/compose-files/docker-compose-california-0.6.0.yml")
	if err != nil {
		log.Fatal(fmt.Sprintf("Call to http.Get() failed with %s\n", err))
	}
	defer resp.Body.Close()

	// [2] Determine the directory (in the deployed file-system) that we will be writing the fetched contents to.
	cmdName := "curl"
	cmdName = "pwd"

	cmd := exec.Command(cmdName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(fmt.Sprintf("Call to exec.Command(cmdName) failed with %s\n", err))
	}
	composeOutput := string(out)

	cmdArgs := composeOutput
	path := strings.TrimSuffix(cmdArgs, "\n")
	// fmt.Sprintf("Determined this to be the directory we will be writing to: {%v}", path)

	composeFile := "/docker-compose.yml"
	filename := path + composeFile

	// [3] Get info about the named file (i.e. under the designator "filename").
	fileInfo, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err.Error(), "Filename-related error!")
	}

	// [4] Determine whether we have _already_ fetched the contents and written them to the deployed file-system.
	if os.IsNotExist(err) {
		println(fileInfo)
		// fmt.Sprintf("File {%v} does NOT exist. Therefore, create it.", fileInfo)
		err = ioutil.WriteFile(filename, []byte(body), 0666)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// fmt.Sprintf("File {%v} already exists! Therefore, NOT creating it.", filename)
	}

	return path, err
}

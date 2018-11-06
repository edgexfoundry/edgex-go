package executor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/system/agent"
)

// ExecuteSnap is a struct for managing services inside the snap
type ExecuteSnap struct {
}

// StopService of ExecuteSnap will stop a service in the snap using `snapctl`
func (oe *ExecuteSnap) StopService(service string, params string) error {

	// use $SNAP to get the name of the snap as snapctl needs to use it
	// and this also lets the name of the snap change if needed
	// and ensures that if you're not running inside a snap we don't try to use
	// snapctl which will fail
	snapName := os.Getenv("SNAP")
	if snapName == "" {
		return errors.New("$SNAP not set, not running inside of a snap")
	}

	// make a map of the service names and use it as a set
	// to check for membership
	serviceNameSet := make(map[string]bool)
	for _, supportedService := range []string{
		internal.ConfigSeedServiceKey,
		internal.CoreCommandServiceKey,
		internal.CoreDataServiceKey,
		internal.CoreMetaDataServiceKey,
		internal.ExportClientServiceKey,
		internal.ExportDistroServiceKey,
		internal.SupportLoggingServiceKey,
		internal.SupportNotificationsServiceKey,
		// note that the sys-mgmt-agent is here and snapctl lets us stop
		// ourselves, but this should probably be handled somewhere else in sys-mgmt-agent
		// more gracefully
		internal.SystemManagementAgentServiceKey,
		internal.SupportSchedulerServiceKey,
	} {
		serviceNameSet[supportedService] = true
	}

	if _, found := serviceNameSet[service]; !found {
		return fmt.Errorf("unknown snap service %s", service)
	}

	// trim the prefix, as the service names in the snap are like "core-command"
	// but the name of the services here are "edgex-core-command"
	rootSvcName := strings.TrimPrefix(service, internal.ServiceKeyPrefix)

	// use snapctl to stop the service - note that this won't disable the service
	// so after a reboot the service will come up again
	cmd := exec.Command("snapctl", "stop", snapName+"."+rootSvcName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// assume that the agent will have been initialize and this isn't a nil interface
		// otherwise this will panic
		agent.LoggingClient.Error(fmt.Sprintf("failed to stop snap service: %s", out))
	}
	return err
}

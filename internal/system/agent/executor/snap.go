package executor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
)

// ExecuteSnap implemnts ServiceStopper, ServiceStarter, ServiceRestarter, ServiceEnabler, and ServiceDisabler
// for managing services inside the snap using snapctl stop, snapctl start, etc.
type ExecuteSnap struct {
}

// Stop will stop a service in a snap using snapctl
func (es *ExecuteSnap) Stop(service string, params []string) error {
	svcName, err := getSnapServiceName(service)
	if err != nil {
		return err
	}
	_, err = runSnapctlOp([]string{"stop"}, svcName)
	return err
}

// Start will stop a service in a snap using snapctl
func (es *ExecuteSnap) Start(service string, params []string) error {
	svcName, err := getSnapServiceName(service)
	if err != nil {
		return err
	}
	_, err = runSnapctlOp([]string{"start"}, svcName)
	return err
}

// Restart will stop and then start again the service in a snap using snapctl
func (es *ExecuteSnap) Restart(service string, params []string) error {
	svcName, err := getSnapServiceName(service)
	if err != nil {
		return err
	}
	_, err = runSnapctlOp([]string{"restart"}, svcName)
	return err
}

// Enable will enable and start the service in a snap using snapctl
// so it will startup automatically on boot
// note that snapctl always starts the service when enabling it,
// there's not currently a way to leave a service stopped now,
// but enable it to run on the next and all subsequent boots
func (es *ExecuteSnap) Enable(service string, params []string) error {
	svcName, err := getSnapServiceName(service)
	if err != nil {
		return err
	}
	_, err = runSnapctlOp([]string{"start", "--enable"}, svcName)
	return err
}

// Disable will disable and stop the service using snapctl so it doesn't startup on boot
// note that snapctl always stops the service when disables it,
// there's not currently a way to leave a service running now,
// but disable it from running on the next and all subsequent boots
func (es *ExecuteSnap) Disable(service string, params []string) error {
	svcName, err := getSnapServiceName(service)
	if err != nil {
		return err
	}
	_, err = runSnapctlOp([]string{"stop", "--disable"}, svcName)
	return err
}

func getSnapServiceName(svc string) (string, error) {
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
		internal.SupportSchedulerServiceKey,
	} {
		serviceNameSet[supportedService] = true
	}

	// ensure that the service is known and exists
	if _, found := serviceNameSet[svc]; !found {
		return "", fmt.Errorf("unknown snap service %s", svc)
	}

	// trim the prefix, as the service names in the snap are like "core-command"
	// but the name of the services here are "edgex-core-command"
	rootSvcName := strings.TrimPrefix(svc, internal.ServiceKeyPrefix)

	// config-seed is actually called core-config-seed in the snap
	if rootSvcName == "config-seed" {
		rootSvcName = "core-config-seed"
	}

	return rootSvcName, nil
}

func runSnapctlOp(ops []string, svc string) ([]byte, error) {
	// use $SNAP_NAME to get the name of the snap as snapctl needs to use it
	// and this also lets the name of the snap change if needed
	// and ensures that if you're not running inside a snap we don't try to use
	// snapctl which will fail
	snapName := os.Getenv("SNAP_NAME")
	if snapName == "" {
		return nil, errors.New("$SNAP_NAME not set, not running inside of a snap")
	}

	// use snapctl to stop the service - note that this won't disable the service
	// so after a reboot the service will come up again
	cmd := exec.Command("snapctl", append(ops, snapName+"."+svc)...)
	return cmd.CombinedOutput()
}

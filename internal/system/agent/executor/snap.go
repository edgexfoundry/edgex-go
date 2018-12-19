/*******************************************************************************
 * Copyright 2018 Canonical Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package executor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
)

// ExecuteSnap is a struct for managing services inside the snap
type ExecuteSnap struct {
}

// StopService of ExecuteSnap will stop a service in the snap using `snapctl`
func (oe *ExecuteSnap) StopService(service string) error {

	// use $SNAP_NAME to get the name of the snap as snapctl needs to use it
	// and this also lets the name of the snap change if needed
	// and ensures that if you're not running inside a snap we don't try to use
	// snapctl which will fail
	snapName := os.Getenv("SNAP_NAME")
	if snapName == "" {
		return errors.New("$SNAP_NAME not set, not running inside of a snap")
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
	_, err := cmd.CombinedOutput()
	return err
}

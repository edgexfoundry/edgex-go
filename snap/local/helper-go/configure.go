/*
 * Copyright (C) 2021 Canonical Ltd
 * Copyright (C) 2023 Intel Corporation
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/canonical/edgex-snap-hooks/v3/env"
	"github.com/canonical/edgex-snap-hooks/v3/log"
	opt "github.com/canonical/edgex-snap-hooks/v3/options"
	"github.com/canonical/edgex-snap-hooks/v3/snapctl"
)

// snapService returns the snap service name for the given app as <snap>.<app>
func snapService(app string) string {
	return env.SnapName + "." + app
}

// This function creates the redis config dir under $SNAP_DATA,
// and creates an empty redis.conf file. This allows the command
// line for the service to always specify the config file, and
// allows for redis to run when security is disabled
func clearRedisConf() error {
	path := filepath.Join(env.SnapData, "/redis/conf/redis.conf")
	if err := ioutil.WriteFile(path, nil, 0644); err != nil {
		return err
	}
	return nil
}

func consulAclFileExists() bool {
	path := filepath.Join(env.SnapData, "/consul/config/consul_acl.json")
	_, err := os.Stat(path)
	return err == nil
}

// This function deletes the Consul ACL configuration file. This
// allows Consul to operate in insecure mode.
func rmConsulAclFile() error {
	path := filepath.Join(env.SnapData, "/consul/config/consul_acl.json")
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}

func processSecuritySwitch() error {
	edgexSecurity, err := snapctl.Get("security").Run()
	if err != nil {
		return fmt.Errorf("error reading snap option: %v", err)
	}

	switch edgexSecurity {
	case "":
		// default - security is enabled
	case "true":
		// manually enabling the disabled switching security
		return fmt.Errorf("security is enabled by default. %s",
			"Once disabled, it can only be re-enabled by re-installing this snap.")
	case "false":
		if err := disableSecurityAndStopAll(); err != nil {
			return fmt.Errorf("error disabling security: %v", err)
		}
	default:
		return fmt.Errorf("unexpected value for security: %s", edgexSecurity)
	}
	return nil
}

func disableSecurityAndStopAll() error {
	// If consul_acls.json doesn't exist, then secret-store has already been
	// disabled, so just return
	if !consulAclFileExists() {
		log.Info("Security is already disabled")
		return nil
	}

	log.Info("!!! DISABLING SECURITY  !!!")

	// Stop all
	// The non-sec services will be started again by the autostart processor
	if err := snapctl.Stop("edgexfoundry").Run(); err != nil {
		return fmt.Errorf("error stopping services: %s", err)
	}

	// Disable autostart of security services
	var autostartKeyValues []string
	for _, s := range append(securityServices, securitySetupServices...) {
		autostartKeyValues = append(autostartKeyValues, "apps."+s+".autostart", "false")
	}
	if err := snapctl.Set(autostartKeyValues...).Run(); err != nil {
		return fmt.Errorf("error setting snap option: %v", err)
	}

	// Disable use of Secret Store for EdgeX services
	if err := snapctl.Set("config.edgex-security-secret-store", "false").Run(); err != nil {
		return fmt.Errorf("error setting snap option: %v", err)
	}

	// Clear redis config
	if err := clearRedisConf(); err != nil {
		return err
	}
	// Clear consul ACLs
	if err := rmConsulAclFile(); err != nil {
		return err
	}

	return nil
}

func configure() {
	log.SetComponentName("configure")
	log.Debug("Start")

	err := processSecuritySwitch()
	if err != nil {
		log.Fatalf("Error processing security switch: %v", err)
	}

	// Process snap config options
	err = opt.ProcessConfig(
		coreData,
		coreMetadata,
		coreCommand,
		coreCommonConfigBootstrapper,
		supportNotifications,
		supportScheduler,
		securitySecretStoreSetup,
		securityBootstrapper, // local executable
		securityProxyAuth,
	)
	if err != nil {
		log.Fatalf("Error processing config options: %v", err)
	}

	// Process autostart to schedule the services start/stop
	// The start/stop operations scheduled here will be performed
	// 	once the configure hook exits without any error.
	err = opt.ProcessAutostart(allServices()...)
	if err != nil {
		log.Fatalf("Error processing autostart options: %v", err)
	}

	log.Debug("End")
}

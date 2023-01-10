/*
 * Copyright (C) 2021 Canonical Ltd
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
 * SPDX-License-Identifier: Apache-2.0'
 */

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/canonical/edgex-snap-hooks/v2/env"
	"github.com/canonical/edgex-snap-hooks/v2/log"
	opt "github.com/canonical/edgex-snap-hooks/v2/options"
	"github.com/canonical/edgex-snap-hooks/v2/snapctl"
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

func disableSecurityAndStopAll(securityServices []string) error {
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
	for _, s := range securityServices {
		autostartKeyValues = append(autostartKeyValues, "apps."+s+".autostart", "false")
	}
	if err := snapctl.Set(autostartKeyValues...).Run(); err != nil {
		return fmt.Errorf("error unsetting snap option: %v", err)
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

	// Process snap config options
	err := opt.ProcessConfig(
		coreData,
		coreMetadata,
		coreCommand,
		supportNotifications,
		supportScheduler,
		securitySecretStoreSetup,
		securityBootstrapper, // local executable
		securityProxySetup,
	)
	if err != nil {
		log.Fatalf("Error processing config options: %v", err)
	}

	// Check if security needs to be disabled
	if edgexSecurity, err := snapctl.Get("security").Run(); err != nil {
		log.Fatalf("Error reading snap option: %v", err)
	} else if edgexSecurity == "enabled" {
		log.Fatalf("Security is enabled by default. %s",
			"Once disabled, it can only be re-enabled by re-installing this snap.")
	} else if edgexSecurity == "disabled" {
		if err := disableSecurityAndStopAll(append(securityServices, securitySetupServices...)); err != nil {
			log.Fatalf("Error disabling security: %v", err)
		}
	}

	// Process autostart to schedule the services start/stop
	// The start/stop operations scheduled here will be performed
	// 	once the configure hook exits without any error.
	err = opt.ProcessAutostart(allServices()...)
	if err != nil {
		log.Fatalf("Error processing autostart options: %v", err)
	}

	// Unset autostart for oneshot services so they don't start again
	var oneshotAutostart []string
	for _, s := range securitySetupServices {
		oneshotAutostart = append(oneshotAutostart, "apps."+s+".autostart")
	}
	if err = snapctl.Unset(oneshotAutostart...).Run(); err != nil {
		log.Fatalf("Error unsetting snap option: %v", err)
	}

	// Schedule the startup of the oneshot service to apply secrets config options
	// 	once the dependency services are ready (see snapcraft.yaml)
	if err := snapctl.Start(snapService(secretsConfigProcessor)).Run(); err != nil {
		log.Fatalf("Error starting service: %s", err)
	}

	log.Debug("End")
}

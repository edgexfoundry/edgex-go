/*
 * Copyright (C) 2022 Canonical Ltd
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
	"flag"
	"fmt"
	"os"
	"strings"

	hooks "github.com/canonical/edgex-snap-hooks/v2"
	"github.com/canonical/edgex-snap-hooks/v2/log"
	opt "github.com/canonical/edgex-snap-hooks/v2/options"
	"github.com/canonical/edgex-snap-hooks/v2/snapctl"
)

// Deprecated
// Legacy options starting with "env." are superseded by app options
func applyConfigOptions(service string) error {
	envJSON, err := snapctl.Get(hooks.EnvConfig + "." + service).Run()
	if err != nil {
		return fmt.Errorf("failed to read config options for %s: %v", service, err)
	}

	if envJSON != "" {
		log.Debugf("Applying env options to service: %s: %s", service, strings.ReplaceAll(envJSON, "\n", ""))
		if err := hooks.HandleEdgeXConfig(service, envJSON, nil); err != nil {
			return err
		}
	}
	return nil
}

// options is called by the main function to configure options
func options() {
	flagset := flag.NewFlagSet("options", flag.ExitOnError)
	service := flagset.String("service", "", "Handle config options of a single service only")
	flagset.Parse(os.Args[2:])

	log.SetComponentName("options")

	log.Info("Configuring options for " + *service)

	// process the EdgeX >=2.2 snap options
	if err := opt.ProcessAppCustomOptions(*service); err != nil {
		log.Fatalf("Could not process custom options: %v", err)
	}

	// process the legacy snap options
	if err := applyConfigOptions(*service); err != nil {
		log.Fatalf("Error handling config options for %s: %v", *service, err)
	}
}

func processAppOptions(deferStartup bool) error {
	log.Info("Processing config options")
	err := opt.ProcessConfig(
		coreData,
		coreMetadata,
		coreCommand,
		supportNotifications,
		supportScheduler,
		appServiceConfigurable,
		securitySecretStoreSetup,
		securityBootstrapper, // local executable
		securityProxySetup,
		systemManagementAgent,
	)
	if err != nil {
		return fmt.Errorf("could not process config options: %v", err)
	}

	// After installation, the configure hook initiates the deferred startup of services,
	// 	processes snap options and exits. The actual services startup happens only
	// 	after the configure hook exits.
	//
	// The following options should not be processed within the configure hook during
	//	the initial installation (install-mode=defer-startup). They should be processed
	//	only on follow-up calls to the configure hook (i.e. when snap set/unset is called)
	if !deferStartup {
		err = opt.ProcessAppCustomOptions(
			secretsConfig, // also processed in security-proxy-post-setup.sh
		)
		if err != nil {
			return fmt.Errorf("could not process custom options: %v", err)
		}
	}

	return nil
}

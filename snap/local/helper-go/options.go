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

	hooks "github.com/canonical/edgex-snap-hooks/v2"
	app_options "github.com/canonical/edgex-snap-hooks/v2/options"
)

func applyConfigOptions(service string) error {
	envJSON, err := cli.Config(hooks.EnvConfig + "." + service)
	if err != nil {
		return fmt.Errorf("failed to read config options for %s: %v", service, err)
	}

	if envJSON != "" {
		hooks.Debug(fmt.Sprintf("edgexfoundry:configure-options: service: %s envJSON: %s", service, envJSON))
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

	fmt.Println("Configuring options for service: " + *service)

	debug, err := cli.Config("debug")
	if err != nil {
		fmt.Println(fmt.Sprintf("edgexfoundry:configure-options: can't read value of 'debug': %v", err))
		os.Exit(1)
	}

	if err = hooks.Init(debug == "true", "edgexfoundry"); err != nil {
		fmt.Println(fmt.Sprintf("edgexfoundry:configure-options: initialization failure: %v", err))
		os.Exit(1)
	}

	hooks.Info("edgexfoundry:configure-options: handling config options for a single service: " + *service)

	// process the EdgeX >=2.2 snap options
	err = app_options.ProcessAppCustomOptions(*service)
	if err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:configure-options: could not process custom options: %v", err))
		os.Exit(1)
	}

	// process the legacy snap options
	if err := applyConfigOptions(*service); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:configure-options: error handling config options for %s: %v", *service, err))
		os.Exit(1)
	}
}

func processAppOptions() {
	err := app_options.ProcessAppConfig(
		"core-data",
		"core-metadata",
		"core-command",
		"support-notifications",
		"support-scheduler",
		"app-service-configurable",
		"security-secretstore-setup",
		"security-bootstrapper", // local executable
		"security-proxy-setup",
		"sys-mgmt-agent",
	)
	if err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:configure could not process config options: %v", err))
		os.Exit(1)
	}

	// After installation, the configure hook initiates the deferred startup of services,
	// 	processes snap options and exits. The actual services startup happens only
	// 	after the configure hook exits.
	//
	// The following options should not be processed within the configure hook during
	//	the initial installation (install-mode=defer-startup). They should be processed
	//	only on follow-up calls to the configure hook (i.e. when snap set/unset is called)
	installMode, err := hooks.NewSnapCtl().Config("install-mode") // this set in the install hook
	if err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:configure failed to read 'install-mode': %s", err))
		os.Exit(1)
	}
	if installMode != "defer-startup" {
		err = app_options.ProcessAppCustomOptions(
			"secrets-config", // also processed in security-proxy-post-setup.sh
		)
		if err != nil {
			hooks.Error(fmt.Sprintf("edgexfoundry:configure: could not process custom options: %v", err))
			os.Exit(1)
		}
	}
}

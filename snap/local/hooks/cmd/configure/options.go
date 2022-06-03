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

	if err := applyConfigOptions(*service); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:configure-options: error handling config options for %s: %v", *service, err))
		os.Exit(1)
	}
}

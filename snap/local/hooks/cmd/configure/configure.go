// -*- Mode: Go; indent-tabs-mode: t -*-

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
	"os"

	hooks "github.com/canonical/edgex-snap-hooks/v2"
)

var cli *hooks.CtlCli = hooks.NewSnapCtl()

const ( // iota is reset to 0
	devVirtService = iota
	kuiperService
	secProxyService
	secStoreService
	otherService
)

const (
	ON  = "on"
	OFF = "off"
)

// handleSingleService starts or stops a service based on
// the given state (ON|OFF). It also ensures that the top
// level service configuration option is set accordingly.
func handleSingleService(name, state string) error {

	switch state {
	case OFF:
		hooks.Debug("edgexfoundry:configure: state: off")
		if err := cli.Stop(name, true); err != nil {
			return err
		}
		if err := cli.SetConfig(name, OFF); err != nil {
			return err
		}
	case ON:
		hooks.Debug("edgexfoundry:configure: state: on")
		if err := cli.Start(name, true); err != nil {
			return err
		}
		if err := cli.SetConfig(name, ON); err != nil {
			return err
		}
	case "":
		hooks.Debug("edgexfoundry:configure: state: ''")
	default:
		return fmt.Errorf("edgexfoundry:configure: invalid state %s for service: %s", state, name)
	}

	return nil
}

func serviceType(name string) int {
	switch name {
	case hooks.ServiceDevVirt:
		return devVirtService
	case hooks.ServiceKuiper:
		return kuiperService
	case hooks.ServiceProxy:
		return secProxyService
	case hooks.ServiceSecStore:
		return secStoreService
	default:
		return otherService
	}
}

// handleAllServices iterates through all of the services in the
// edgexfoundry snap and:
//
// - queries the config option associated with the service state (on|off)
// - queries the environment configuration for the service (env.<service-name>)
//   - if env configuration for the service exists, use it to write a
//     service-specific .env file to the service config dir in $SNAP_DATA
// - start/stop any tightly couple services (e.g. if the secret-store
//   is disabled, the proxy also has to come down) if required
// - start/stop the service itself if required
//
// NOTE - at this time, this function does *not* restart a service based
// on env configuration changes. If changes are made after a service has
// been started, the service must be restarted manually.
//
func handleAllServices() error {

	// grab and log the current service configuration
	for _, s := range hooks.Services {
		var envJSON string

		status, err := cli.Config(s)
		if err != nil {
			return err
		}

		hooks.Debug(fmt.Sprintf("edgexfoundry:configure: service: %s status: %s", s, status))

		serviceCfg := hooks.EnvConfig + "." + s
		envJSON, err = cli.Config(serviceCfg)
		if err != nil {
			return fmt.Errorf("edgexfoundry:configure failed to read service %s configuration - %v", s, err)
		}

		if envJSON != "" {
			hooks.Debug(fmt.Sprintf("edgexfoundry:configure: service: %s envJSON: %s", s, envJSON))
			if err := hooks.HandleEdgeXConfig(s, envJSON, nil); err != nil {
				return err
			}
		}

		sType := serviceType(s)

		switch sType {
		case devVirtService:
			hooks.Debug("edgexfoundry:configure: device-virtual")
			// device-virtual is built with device-sdk-go which waits
			// for core-data and core-metadata to come online, so if we are
			// enabling a device service, we should also enable those services
			if status == ON {
				if err = handleSingleService("core-data", ON); err != nil {
					return err
				}
				if err = handleSingleService("core-metadata", ON); err != nil {
					return err
				}
			}

			// handle the service too
			if err = handleSingleService(s, status); err != nil {
				return nil
			}
		case kuiperService:
			hooks.Debug("edgexfoundry:configure: kuiper")

			// if we are turning kuiper on, make sure
			// app-service-configurable is on too
			if err = handleSingleService("app-service-configurable", status); err != nil {
				return err
			}
			if err = handleSingleService("kuiper", status); err != nil {
				return err
			}
		case secProxyService:
			hooks.Debug("edgexfoundry:configure: proxy")
			// FIXME - this logic always overrwrites the existing value. This
			// should be fixed.

			// the security-proxy consists of the following base services
			// - kong
			// - postgres (because kong requires it)
			if err = handleSingleService("postgres", status); err != nil {
				return err
			}
			if err = handleSingleService("kong-daemon", status); err != nil {
				return err
			}
			if err = handleSingleService("security-proxy-setup", status); err != nil {
				return err
			}

			// additionally, the security-proxy needs to use the following
			// services:
			// - vault (because security-proxy-setup will access/store secrets in vault)
			// - security-secretstore-setup
			// so if we are turning the security-api-gateway on, then turn
			// those services on too
			if status == ON {
				if err = handleSingleService("vault", ON); err != nil {
					return err
				}
				if err = handleSingleService("security-secretstore-setup", ON); err != nil {
					return err
				}
			}
		case secStoreService:
			hooks.Debug("edgexfoundry:configure: secretstore")
			// the security-api-gateway consists of the following services:
			// - vault
			// - security-secretstore-setup
			// and since the security-api-gateway needs to be able to use
			// security-secret-store, we also need to turn off those services
			// if this one is disabled
			if status == OFF {
				if err = handleSingleService("postgres", OFF); err != nil {
					return err
				}
				if err = handleSingleService("kong-daemon", OFF); err != nil {
					return err
				}
				if err = handleSingleService("security-proxy-setup", OFF); err != nil {
					return err
				}

			}
			if err = handleSingleService("vault", status); err != nil {
				return err
			}
			if err = handleSingleService("security-secretstore-setup", status); err != nil {
				return err
			}
		default:
			hooks.Debug("edgexfoundry:configure: other service")
			// default case for all other services just enable/disable the service using
			// snapd/systemd
			// if the service is meant to be off, then disable it
			if err = handleSingleService(s, status); err != nil {
				return nil
			}
		}
	}

	return nil
}

func main() {
	var debug = false
	var err error

	status, err := cli.Config("debug")
	if err != nil {
		fmt.Println(fmt.Sprintf("edgexfoundry:configure: can't read value of 'debug': %v", err))
		os.Exit(1)
	}
	if status == "true" {
		debug = true
	}

	if err = hooks.Init(debug, "edgexfoundry"); err != nil {
		fmt.Println(fmt.Sprintf("edgexfoundry:configure: initialization failure: %v", err))
		os.Exit(1)

	}

	if err = handleAllServices(); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:configure: error handling services: %v", err))
		os.Exit(1)
	}
}

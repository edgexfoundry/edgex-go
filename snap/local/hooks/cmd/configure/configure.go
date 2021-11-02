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
	kuiperService = iota
	secProxyService
	secStoreService
	otherService
)

const (
	INSTALL = "install"
	ON      = "on"
	OFF     = "off"
)

func getKuiperServices() []string {
	return []string{hooks.ServiceAppCfg, hooks.ServiceKuiper}
}

func getProxyServices() []string {
	return []string{"postgres", "kong-daemon", "security-proxy-setup"}
}

// handleSingleService starts or stops a service based on
// the given state (ON|OFF). It also ensures that the top
// level service configuration option is set accordingly.
func handleSingleService(name, state string, reset bool) error {

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
		if reset {
			if err := cli.UnsetConfig(name); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("edgexfoundry:configure: invalid state %s for service: %s", state, name)
	}

	return nil
}

func handleServices(serviceList []string, state string, reset bool) error {
	for _, s := range serviceList {
		if err := handleSingleService(s, state, reset); err != nil {
			return err
		}
	}
	return nil
}

func serviceType(name string) int {
	switch name {
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

func buildStartCmd(startServices []string, newServices []string) []string {
	for _, s := range newServices {
		s = hooks.SnapName + "." + s
		startServices = append(startServices, s)
	}
	return startServices
}

// handleAllServices iterates through all of the services in the
// edgexfoundry snap and:
//
// - queries the config option associated with the service state (on|off|install|'')
// - queries the environment configuration for the service (env.<service-name>)
//   - if env configuration for the service exists, use it to write a
//     service-specific .env file to the service config dir in $SNAP_DATA
// - if install is true, just build a list of services to start, don't
//   actually start/stop any services
// - otherwise:
//   - start/stop any tightly coupled services (e.g. if the secret-store
//     is disabled, the proxy also has to come down) if required
//   - start/stop the service itself if required
//
// NOTE - at this time, this function does *not* restart a service based
// on env configuration changes. If changes are made after a service has
// been started, the service must be restarted manually.
//
func handleAllServices(install bool) (error, []string) {
	var startServices []string
	var serviceList []string
	secretStoreActive := true

	// grab and log the current service configuration
	for _, s := range hooks.Services {
		var envJSON string

		status, err := cli.Config(s)
		if err != nil {
			return err, nil
		}

		hooks.Info(fmt.Sprintf("edgexfoundry:configure: service: %s status: %s", s, status))

		serviceCfg := hooks.EnvConfig + "." + s
		envJSON, err = cli.Config(serviceCfg)
		if err != nil {
			err = fmt.Errorf("edgexfoundry:configure failed to read service %s configuration - %v", s, err)
			return err, nil
		}

		if envJSON != "" {
			hooks.Debug(fmt.Sprintf("edgexfoundry:configure: service: %s envJSON: %s", s, envJSON))
			if err := hooks.HandleEdgeXConfig(s, envJSON, nil); err != nil {
				return err, nil
			}
		}

		// SecBootstrapper is a valid service for configuration
		// purposes, however it isn't individually controlable
		// using on|off, so once configuration has been handled
		// skip to the next service.
		if s == hooks.ServiceSecBootstrapper {
			continue
		}

		sType := serviceType(s)

		switch sType {
		case kuiperService:
			hooks.Debug("edgexfoundry:configure: kuiper")

			switch status {
			case INSTALL:
				if err := cli.UnsetConfig("kuiper"); err != nil {
					return err, nil
				}
				fallthrough
			case ON:
				serviceList = getKuiperServices()
				if install {
					startServices = append(startServices, serviceList...)
					hooks.Info(fmt.Sprintf("edgexfoundry:configure startServices: %v", startServices))
					continue
				}
			case OFF:
				if install {
					continue
				}

				serviceList = getKuiperServices()
			default:
				// Note - this is the default status of all services if no
				// configuration has been specified; no-op
				continue
			}

		case secProxyService:
			hooks.Info("edgexfoundry:configure: proxy")

			switch status {
			case INSTALL:
				if err := cli.UnsetConfig("security-proxy"); err != nil {
					return err, nil
				}
				fallthrough
			case ON:
				// NOTE: the original bash based implementation would
				// additionally handle the secret-store dependency.
				// Due to the added complexity, this initial implementation
				// does not automatically handle enabling the secret-store
				// if/when the proxy is dynamically enabled.
				if !secretStoreActive {
					err = fmt.Errorf("edgexfoundry:configure security-proxy=on not allowed;" +
						"secret-store=off")
					return err, nil
				}

				serviceList = getProxyServices()
				if install {
					startServices = append(startServices, serviceList...)
					hooks.Info(fmt.Sprintf("edgexfoundry:configure startServices: %v", startServices))
					continue
				}
			case OFF:
				if install {
					continue
				}

				serviceList = getProxyServices()
			default:
				// Note - this is the default status of all services if no
				// configuration has been specified; no-op
				continue
			}

		case secStoreService:
			hooks.Info("edgexfoundry:configure: secretstore")

			switch status {
			case INSTALL:
				if err := cli.UnsetConfig("security-secret-store"); err != nil {
					return err, nil
				}
				fallthrough
			case ON:
				serviceList = []string{"vault", "security-secretstore-setup",
					"security-consul-bootstrapper", "security-bootstrapper-redis"}
				hooks.Info(fmt.Sprintf("edgexfoundry:configure serviceList: %v", serviceList))
				if install {
					startServices = append(startServices, serviceList...)
					hooks.Info(fmt.Sprintf("edgexfoundry:configure startServices: %v", startServices))
					continue
				}
			case OFF:
				secretStoreActive = false

				if install {
					continue
				}

				serviceList = []string{"postgres", "kong-daemon", "security-proxy-setup",
					"vault", "security-secretstore-setup", "security-consul-bootstrapper",
					"security-bootstrapper-redis"}

				// TODO: the original Hanoi implementation did NOT handle restarting the
				// rest of the framework, nor does this implementation.
				//
				// If/when we can properly disable secret-store usage *after* install, we
				// need to handle the following:
				//
				// - stop all services that use the secret store
				// - stop consul & redis
				// - clear redis password (rm $SNAP_DATA/conf/redis.conf)
				//   touch same file
				// - rm the consul ACL conf file (and maybe other consul files)
				// - start all the services just disabled
			default:
				// Note - this is the default status of all services if no
				// configuration has been specified; no-op
				continue
			}

		default:
			hooks.Info("edgexfoundry:configure: other service")
			// default case for all other services

			switch status {
			case INSTALL:
				// FIXME: make this cli.UnsetConfig
				if err := cli.UnsetConfig(s); err != nil {
					return err, nil
				}
				fallthrough
			case ON:
				serviceList = []string{s}

				if install {
					startServices = append(startServices, s)
					hooks.Info(fmt.Sprintf("edgexfoundry:configure startServices: %v", startServices))
					continue
				}
			case OFF:
				if install {
					continue
				}

				serviceList = append(serviceList, s)
			default:
				// Note - this is the default status of all services if no
				// configuration has been specified; no-op
				continue
			}
		}

		hooks.Info(fmt.Sprintf("edgexfoundry:configure calling handleServices: %v", serviceList))
		if err = handleServices(serviceList, status, false); err != nil {
			return err, nil
		}
		// clear serviceList
		serviceList = nil
	}

	return nil, startServices
}

func main() {
	var debug = false
	var err error
	var installMode = false
	var startServices []string

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

	install, err := cli.Config("install")
	if err != nil {
		fmt.Println(fmt.Sprintf("edgexfoundry:configure: reading 'install': %v", err))
		os.Exit(1)
	}

	if install == "true" {
		installMode = true
	}

	// if installMode is true, then this function just returns a list of services
	// to start, otherwise it actually will start/stop services based on current
	// configuration status of each service
	if err, startServices = handleAllServices(installMode); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:configure: error handling services: %v", err))
		os.Exit(1)
	}

	// start all the services and clear install flag
	if install == "true" {
		hooks.Info(fmt.Sprintf("edgexfoundry:configure install=true; starting disabled services"))

		if err = cli.StartMultiple(true, startServices...); err != nil {
			hooks.Error(fmt.Sprintf("edgexfoundry:configure failure starting/enabling services: %v", err))
			os.Exit(1)
		}

		if err = cli.SetConfig("install", "false"); err != nil {
			hooks.Error(fmt.Sprintf("edgexfoundry:install setting 'install=false'; %v", err))
			os.Exit(1)
		}
	}
}

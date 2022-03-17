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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	hooks "github.com/canonical/edgex-snap-hooks/v2"
)

const ( // iota is reset to 0
	kuiperService = iota
	secProxyService
	secStoreService
	otherService
)

const (
	ON    = "on"
	OFF   = "off"
	UNSET = ""
)

// getKuiperServices returns the list of services used for
// EdgeX rules-engine processing.
func getKuiperServices() []string {
	return []string{hooks.ServiceAppCfg, hooks.ServiceKuiper}
}

// getProxyServices returns the list of services which implement
// the API Gateway. Note this list *excludes* Consul and the
// Secret Store services.
func getProxyServices() []string {
	return []string{"postgres", "kong-daemon", "security-proxy-setup"}
}

// getSecretStoreServices returns the list of services which implement
// the Secret Store and related dependencies (i.e. the services that
// secure redis and consul which are tightly bound to the secret store
// being enabled).
func getSecretStoreServices() []string {
	return []string{"security-secretstore-setup", "vault",
		"security-consul-bootstrapper", "security-bootstrapper-redis"}
}

// getEdgeXRefServices returns the list of EdgeX reference services in the snap
// (excludes all non-EdgeX runtime dependencies, and security-*-setup jobs).
func getEdgeXRefServices() []string {
	return []string{"core-data", "core-metadata", "core-command",
		"device-virtual", "support-notifications",
		"support-scheduler", "sys-mgmt-agent"}
}

// getRequiredServices returns the minimum list of required
// snap services for a working EdgeX instance.
func getRequiredServices() []string {
	return []string{"consul", "redis", "core-metadata"}
}

// getCoreDefaultServices returns the list of core services
// that are started by default (in addition to the required
// services)
func getCoreDefaultServices() []string {
	return []string{"core-command", "core-data"}
}

// getOptServices returns the list of optional EdgeX services
// (i.e. disabled by default).
//
// Note:
// - sys-mgmt-agent isn't included because as of Ireland
//   it's considered deprecated.
// - kuiper isn't included because it's not yet possible
//   to provide kuiper configuration via content interface
func getOptServices() []string {
	return []string{"support-notifications", "support-scheduler", "device-virtual"}
}

func isDisableAllowed(s string) error {
	for _, v := range getRequiredServices() {
		if s == v {
			return fmt.Errorf("edgexfoundry:configure: can't disable required service: %s", s)
		}
	}
	return nil
}

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
	default:
		return fmt.Errorf("edgexfoundry:configure: invalid state %s for service: %s", state, name)
	}

	return nil
}

func handleServices(serviceList []string, state string) error {
	for _, s := range serviceList {
		if err := handleSingleService(s, state); err != nil {
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

// This function creates the redis config dir under $SNAP_DATA,
// and creates an empty redis.conf file. This allows the command
// line for the service to always specify the config file, and
// allows for redis when the config option security-secret-store
// is "on" or "off".
func clearRedisConf() error {
	path := filepath.Join(hooks.SnapData, "/redis/conf/redis.conf")
	if err := ioutil.WriteFile(path, nil, 0644); err != nil {
		return err
	}
	return nil
}

func consulAclFileExists() bool {
	path := filepath.Join(hooks.SnapData, "/consul/config/consul_acl.json")
	_, err := os.Stat(path)
	return err == nil
}

// This function deletes the Consul ACL configuration file. This
// allows Consul to operate in insecure mode.
func rmConsulAclFile() error {
	path := filepath.Join(hooks.SnapData, "/consul/config/consul_acl.json")
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}

func disableSecretStoreAndRestart() error {
	hooks.Info(fmt.Sprintf("edgexfoundry:configure: disabling secret store"))

	// if consul_acls.json doesn't exist, then secret-store has already been
	// disabled, so just return
	if !consulAclFileExists() {
		hooks.Info(fmt.Sprintf("edgexfoundry:configure: secret store is already disabled"))
		return nil
	}

	// stop & disable proxy services
	for _, s := range getProxyServices() {
		if err := handleSingleService(s, OFF); err != nil {
			return err
		}
	}

	// stop & disable secret store services
	for _, s := range getSecretStoreServices() {
		if err := handleSingleService(s, OFF); err != nil {
			return err
		}
	}

	// stop EdgeX services
	// TODO: can't use handleServices because that would result in the
	// snap config option for each service to be needlessly set to "off"
	// then back to "on"; re-factor handleServices/handleSingleService
	for _, s := range getEdgeXRefServices() {
		if err := cli.Stop(s, false); err != nil {
			return err
		}
	}

	// stop Kuiper-related services
	// TODO - kuiper will be stopped, but not restarted because
	// additional re-configuration may be needed.
	for _, s := range getKuiperServices() {
		if err := cli.Stop(s, false); err != nil {
			return err
		}
	}

	// stop redis
	if err := cli.Stop("redis", false); err != nil {
		return err
	}

	// stop consul
	if err := cli.Stop("consul", false); err != nil {
		return err
	}

	// - clear redis password
	if err := clearRedisConf(); err != nil {
		return err
	}

	// - clear consul ACLs
	if err := rmConsulAclFile(); err != nil {
		return err
	}

	// - start required services
	for _, s := range getRequiredServices() {
		if err := cli.Start(s, false); err != nil {
			return err
		}
	}

	// Now check config status of the optional EdgeX
	// services and restart where necessary
	for _, s := range getEdgeXRefServices() {
		status, err := cli.Config(s)
		if err != nil {
			return err
		}

		// walk thru remaining edgex services
		// if status is ON, start
		// if status isn't set, if the service is
		// part of the enabledServices (i.e. services
		// always started), then also start it
		if status == ON || (status == "" && strings.HasPrefix(s, "core-")) {
			if err := cli.Start(s, false); err != nil {
				return err
			}
		}
	}
	return nil
}

// handleAllServices iterates through all of the services in the
// edgexfoundry snap and:
//
// - queries the config option associated with the service state (on|off|'')
// - queries the environment configuration for the service (env.<service-name>)
//   - if env configuration for the service exists, use it to write a
//     service-specific .env file to the service config dir in $SNAP_DATA
// - if deferStartup == true, continue to the next service
// - otherwise handle runtime state changes
//   - start/stop any tightly coupled services (e.g. if the secret-store
//     is disabled, the proxy also has to come down) if required
//   - start/stop the service itself if required
//
//
// NOTE - at this time, this function does *not* restart a service based
// on env configuration changes. If changes are made after a service has
// been started, the service must be restarted manually.
//
func handleAllServices(deferStartup bool) error {
	secretStoreActive := true

	// grab and log the current service configuration
	for _, s := range hooks.Services {
		var serviceList []string

		status, err := cli.Config(s)
		if err != nil {
			return err
		}

		hooks.Debug(fmt.Sprintf("edgexfoundry:configure: service: %s status: %s", s, status))

		err = applyConfigOptions(s)
		if err != nil {
			return fmt.Errorf("failed to apply config options for %s: %v", s, err)
		}

		// if deferStartup is set, don't start/stop services
		if deferStartup {
			continue
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
			case ON, OFF:
				serviceList = getKuiperServices()
			case UNSET:
				// this is the default status of all services if no
				// configuration has been specified; no-op
				continue
			default:
				return fmt.Errorf("edgexfoundry:configure: invalid value for kuiper: %s", status)
			}

		case secProxyService:
			hooks.Debug("edgexfoundry:configure: proxy")

			switch status {
			case ON:
				// NOTE: the original bash based implementation would
				// additionally handle the secret-store dependency.
				// Due to the added complexity, this initial implementation
				// does not automatically handle enabling the secret-store
				// if/when the proxy is dynamically enabled.
				if !secretStoreActive {
					err = fmt.Errorf("edgexfoundry:configure security-proxy=on not allowed;" +
						"secret-store=off")
					return err
				}

				fallthrough
			case OFF:
				serviceList = getProxyServices()
			case UNSET:
				// this is the default status of all services if no
				// configuration has been specified; no-op
				continue
			default:
				return fmt.Errorf("edgexfoundry:configure: invalid value for security-proxy: %s", status)
			}

		case secStoreService:
			hooks.Debug("edgexfoundry:configure: secretstore")

			switch status {
			case ON:
				return fmt.Errorf("edgexfoundry:configure security-secret-store=on not allowed")
			case OFF:
				// TODO - this var is used by the secProxyCase to ensure that the
				// secret store is active when the proxy is being enabled at runtime.
				// This relies on the fact that the secret store comes before the proxy
				// in hooks.Services. To make this less fragile, the proxy case should
				// check the status of the secret store directly.
				secretStoreActive = false

				if err = disableSecretStoreAndRestart(); err != nil {
					return err
				}
				continue
			case UNSET:
				// this is the default status of all services if no
				// configuration has been specified; no-op
				continue
			default:
				return fmt.Errorf("edgexfoundry:configure: invalid value for security-secret-store: %s", status)
			}

		default:
			hooks.Debug("edgexfoundry:configure: other service")
			// default case for all other services

			switch status {
			case ON:
				serviceList = []string{s}
			case OFF:
				if err := isDisableAllowed(s); err != nil {
					return err
				}
				serviceList = []string{s}
			case UNSET:
				// this is the default status of all services if no
				// configuration has been specified; no-op
				continue
			default:
				return fmt.Errorf("edgexfoundry:configure: invalid value for %s: %s", s, status)
			}
		}

		hooks.Debug(fmt.Sprintf("edgexfoundry:configure calling handleServices: %v", serviceList))
		if err = handleServices(serviceList, status); err != nil {
			return err
		}
	}

	return nil
}

func checkCoreConfig(services []string) ([]string, error) {
	// walk thru the list of default services
	for _, s := range getCoreDefaultServices() {
		status, err := cli.Config(s)
		if err != nil {
			return nil, err
		}

		switch status {
		case OFF:
			break
		case ON, UNSET:
			services = append(services, s)
		default:
			err = fmt.Errorf("edgexfoundry:configure: invalid value: %s for %s", status, s)
			return nil, err
		}
	}
	return services, nil
}

func checkOptConfig(services []string) ([]string, error) {
	// walk thru the list of default services
	for _, s := range getOptServices() {
		status, err := cli.Config(s)
		if err != nil {
			return nil, err
		}

		switch status {
		case OFF, UNSET:
			break
		case ON:
			services = append(services, s)
		default:
			err = fmt.Errorf("edgexfoundry:configure: invalid value: %s for %s", status, s)
			return nil, err
		}
	}
	return services, nil
}

func checkSecurityConfig(services []string) ([]string, error) {

	status, err := cli.Config("security-secret-store")
	if err != nil {
		return nil, err
	}

	switch status {
	case OFF:
		// if security-secret-store is off, no proxy either...
		return services, nil
	case UNSET:
		// default behavior
		services = append(services, getSecretStoreServices()...)
	default:
		err = fmt.Errorf("edgexfoundry:configure: invalid setting for security-secret-store: %s", status)
		return nil, err
	}

	// check secret-proxy
	status, err = cli.Config("security-proxy")
	if err != nil {
		return nil, err
	}

	switch status {
	case OFF:
		break
	case UNSET:
		// default behavior
		services = append(services, getProxyServices()...)
	default:
		err = fmt.Errorf("edgexfoundry:configure: invalid setting for security-proxy: %s", status)
		return nil, err
	}
	return services, nil
}

func configure() {
	var debug = false
	var err error
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
	hooks.Info("edgexfoundry:configure: started")

	val, err := cli.Config("install-mode")
	if err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:configure: reading 'install-mode': %v", err))
		os.Exit(1)
	}

	deferStartup := (val == "defer-startup")
	hooks.Info(fmt.Sprintf("edgexfoundry:configure: deferStartup=%v", deferStartup))

	// handle per service configuration and enable/disable services
	if err = handleAllServices(deferStartup); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:configure: error handling services: %v", err))
		os.Exit(1)
	}

	// Handle deferred startup of services disabled in the install hook.
	//
	// NOTE - there's code duplication between this startup logic and
	// the function handleAllServices(). While it might be possible to
	// merge the two, since delayed startup is itself a workaround to
	// an underlying snapd limitation (namely that services are started
	// before the config hook runs), leaving the duplication means less
	// re-factoring if/when snapd adds a new hook.
	if deferStartup {
		hooks.Info(fmt.Sprintf("edgexfoundry:configure install-mode=defer-startup; starting disabled services"))

		// add required services
		startServices = append(startServices, getRequiredServices()...)

		// check security configuration
		startServices, err = checkSecurityConfig(startServices)
		if err != nil {
			hooks.Error(fmt.Sprintf("edgexfoundry:configure: security config error; %v", err))
			os.Exit(1)
		}

		// TODO: don't support kuiper until it's possible to share
		// kuiper & app-services-configurable (rules-engine) config
		// via content interface

		// check core services
		startServices, err = checkCoreConfig(startServices)
		if err != nil {
			hooks.Error(fmt.Sprintf("edgexfoundry:configure: core config error; %v", err))
			os.Exit(1)
		}

		// check optional services
		startServices, err = checkOptConfig(startServices)
		if err != nil {
			hooks.Error(fmt.Sprintf("edgexfoundry:configure: optional config error; %v", err))
			os.Exit(1)
		}

		// NOTE - the services will be started after the configure hook finishes
		if err = cli.StartMultiple(true, startServices...); err != nil {
			hooks.Error(fmt.Sprintf("edgexfoundry:configure failure starting/enabling services: %v", err))
			os.Exit(1)
		}

		if err = cli.UnsetConfig("install-mode"); err != nil {
			hooks.Error(fmt.Sprintf("edgexfoundry:install un-setting 'install'; %v", err))
			os.Exit(1)
		}
	}
}

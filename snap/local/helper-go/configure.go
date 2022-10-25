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

	"github.com/canonical/edgex-snap-hooks/v2/env"
	"github.com/canonical/edgex-snap-hooks/v2/log"
	"github.com/canonical/edgex-snap-hooks/v2/snapctl"
)

const (
	ON    = "on"
	OFF   = "off"
	UNSET = ""
)

var (
	rulesEngineServices = []string{
		appServiceConfigurable,
		eKuiper,
	}
	proxyServices = []string{
		kong,
		postgres,
		securityProxySetup,
	}
	secretStoreServices = []string{
		vault,
		securitySecretStoreSetup,
		securityBootstrapperConsul,
		securityBootstrapperRedis,
	}
	referenceServices = []string{
		coreData,
		coreMetadata,
		coreCommand,
		supportNotifications,
		supportScheduler,
		systemManagementAgent,
	}
	requiredServices = []string{
		consul,
		redis,
		coreMetadata,
	}
	coreDefaultServices = []string{
		coreCommand,
		coreData,
	}
	optionalServices = []string{
		supportNotifications,
		supportScheduler,
		eKuiper,
		appServiceConfigurable,
		systemManagementAgent,
	}
)

// snapService returns the snap service name for the given app as <snap>.<app>
func snapService(app string) string {
	return env.SnapName + "." + app
}

func isDisableAllowed(s string) error {
	for _, v := range requiredServices {
		if s == v {
			return fmt.Errorf("can't disable required service: %s", s)
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
		log.Debugf("%s state: off", name)
		if err := snapctl.Stop(snapService(name)).Disable().Run(); err != nil {
			return err
		}
		if err := snapctl.Set(name, OFF).Run(); err != nil {
			return err
		}
	case ON:
		log.Debugf("%s state: on", name)
		if err := snapctl.Start(snapService(name)).Enable().Run(); err != nil {
			return err
		}
		if err := snapctl.Set(name, ON).Run(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid state %s for service: %s", state, name)
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

// This function creates the redis config dir under $SNAP_DATA,
// and creates an empty redis.conf file. This allows the command
// line for the service to always specify the config file, and
// allows for redis when the config option security-secret-store
// is "on" or "off".
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

func disableSecretStoreAndRestart() error {
	log.Info("disabling secret store")

	// if consul_acls.json doesn't exist, then secret-store has already been
	// disabled, so just return
	if !consulAclFileExists() {
		log.Info("secret store is already disabled")
		return nil
	}

	// stop & disable proxy services
	for _, s := range proxyServices {
		if err := handleSingleService(s, OFF); err != nil {
			return err
		}
	}

	// stop & disable secret store services
	for _, s := range secretStoreServices {
		if err := handleSingleService(s, OFF); err != nil {
			return err
		}
	}

	// stop EdgeX services
	// TODO: can't use handleServices because that would result in the
	// snap config option for each service to be needlessly set to "off"
	// then back to "on"; re-factor handleServices/handleSingleService
	for _, s := range referenceServices {
		if err := snapctl.Stop(snapService(s)).Run(); err != nil {
			return err
		}
	}

	// stop Kuiper-related services
	// TODO - kuiper will be stopped, but not restarted because
	// additional re-configuration may be needed.
	for _, s := range rulesEngineServices {
		if err := snapctl.Stop(snapService(s)).Run(); err != nil {
			return err
		}
	}

	// stop redis
	if err := snapctl.Stop(snapService(redis)).Run(); err != nil {
		return err
	}

	// stop consul
	if err := snapctl.Stop(snapService(consul)).Run(); err != nil {
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
	for _, s := range requiredServices {
		if err := snapctl.Start(snapService(s)).Run(); err != nil {
			return err
		}
	}

	// Now check config status of the optional EdgeX
	// services and restart where necessary
	for _, s := range referenceServices {
		status, err := snapctl.Get(s).Run()
		if err != nil {
			return err
		}

		// walk thru remaining edgex services
		// if status is ON, start
		// if status isn't set, if the service is
		// part of the enabledServices (i.e. services
		// always started), then also start it
		if status == ON || (status == "" && strings.HasPrefix(s, "core-")) {
			if err := snapctl.Start(snapService(s)).Run(); err != nil {
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

	var services = []string{
		// core services
		consul,
		redis,
		coreData,
		coreMetadata,
		coreCommand,
		// support services
		supportNotifications,
		supportScheduler,
		eKuiper,
		// app-services
		appServiceConfigurable,
		// security services
		securitySecretStore,
		securityProxy,
		// management
		systemManagementAgent,
		// others
		securityBootstrapper, // oneshot service
	}

	// grab and log the current service configuration
	for _, s := range services {
		var serviceList []string

		status, err := snapctl.Get(s).Run()
		if err != nil {
			return err
		}

		log.Debugf("Handling service: %s, status: '%s'", s, status)

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
		if s == securityBootstrapper {
			continue
		}

		// sType := serviceType(s)

		switch s {
		case eKuiper:
			switch status {
			case ON, OFF:
				serviceList = rulesEngineServices
			case UNSET:
				// this is the default status of all services if no
				// configuration has been specified; no-op
				continue
			default:
				return fmt.Errorf("invalid value for kuiper: %s", status)
			}

		case securityProxy:
			switch status {
			case ON:
				// NOTE: the original bash based implementation would
				// additionally handle the secret-store dependency.
				// Due to the added complexity, this initial implementation
				// does not automatically handle enabling the secret-store
				// if/when the proxy is dynamically enabled.
				if !secretStoreActive {
					return fmt.Errorf("security-proxy=on not allowed when secret-store is off")
				}

				fallthrough
			case OFF:
				serviceList = proxyServices
			case UNSET:
				// this is the default status of all services if no
				// configuration has been specified; no-op
				continue
			default:
				return fmt.Errorf("invalid value for security-proxy: %s", status)
			}

		case securitySecretStore:
			switch status {
			case ON:
				return fmt.Errorf("security-secret-store=on not allowed")
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
				return fmt.Errorf("invalid value for security-secret-store: %s", status)
			}

		default: // default case for all other services
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
				return fmt.Errorf("invalid value for %s: %s", s, status)
			}
		}

		log.Debugf("calling handleServices: %v", serviceList)
		if err = handleServices(serviceList, status); err != nil {
			return err
		}
	}

	return nil
}

func checkCoreConfig(services []string) ([]string, error) {
	// walk thru the list of default services
	for _, s := range coreDefaultServices {
		status, err := snapctl.Get(s).Run()
		if err != nil {
			return nil, err
		}

		switch status {
		case OFF:
			break
		case ON, UNSET:
			services = append(services, s)
		default:
			err = fmt.Errorf("invalid value: %s for %s", status, s)
			return nil, err
		}
	}
	return services, nil
}

func checkOptConfig(services []string) ([]string, error) {
	// walk thru the list of default services
	for _, s := range optionalServices {
		status, err := snapctl.Get(s).Run()
		if err != nil {
			return nil, err
		}

		switch status {
		case OFF, UNSET:
			break
		case ON:
			services = append(services, s)
		default:
			err = fmt.Errorf("invalid value: %s for %s", status, s)
			return nil, err
		}
	}
	return services, nil
}

func checkSecurityConfig(services []string) ([]string, error) {

	status, err := snapctl.Get("security-secret-store").Run()
	if err != nil {
		return nil, err
	}

	switch status {
	case OFF:
		// if security-secret-store is off, no proxy either...
		return services, nil
	case UNSET:
		// default behavior
		services = append(services, secretStoreServices...)
	default:
		err = fmt.Errorf("invalid setting for security-secret-store: %s", status)
		return nil, err
	}

	// check secret-proxy
	status, err = snapctl.Get("security-proxy").Run()
	if err != nil {
		return nil, err
	}

	switch status {
	case OFF:
		break
	case UNSET:
		// default behavior
		services = append(services, proxyServices...)
	default:
		err = fmt.Errorf("invalid setting for security-proxy: %s", status)
		return nil, err
	}
	return services, nil
}

func configure() {
	log.SetComponentName("configure")
	log.Debug("Start")

	installMode, err := snapctl.Get("install-mode").Run()
	if err != nil {
		log.Fatalf("error reading 'install-mode': %v", err)
	}
	deferStartup := (installMode == "defer-startup")
	log.Infof("Defer startup: %t", deferStartup)

	// process the EdgeX >=2.2 app options
	if err := processAppOptions(deferStartup); err != nil {
		log.Fatalf("error processing app options: %v", err)
	}

	// handle per service configuration and enable/disable services
	if err = handleAllServices(deferStartup); err != nil {
		log.Fatalf("error handling services: %v", err)
	}

	// Handle deferred startup of services disabled in the install hook.
	//
	// NOTE - there's code duplication between this startup logic and
	// the function handleAllServices(). While it might be possible to
	// merge the two, since delayed startup is itself a workaround to
	// an underlying snapd limitation (namely that services are started
	// before the config hook runs), leaving the duplication means less
	// re-factoring if/when snapd adds a new hook.
	var startServices []string
	if deferStartup {
		log.Info("Defer startup: starting disabled services")

		// add required services
		startServices = append(startServices, requiredServices...)

		// check security configuration
		startServices, err = checkSecurityConfig(startServices)
		if err != nil {
			log.Fatalf("security service config error: %v", err)
		}

		// TODO: don't support kuiper until it's possible to share
		// kuiper & app-services-configurable (rules-engine) config
		// via content interface

		// check core services
		startServices, err = checkCoreConfig(startServices)
		if err != nil {
			log.Fatalf("core service config error: %v", err)
		}

		// check optional services
		startServices, err = checkOptConfig(startServices)
		if err != nil {
			log.Fatalf("optional service config error: %v", err)
		}

		for i, s := range startServices {
			startServices[i] = snapService(s)
		}
		// NOTE: the services will be scheduled to start by snapd after the configure hook exits
		if err = snapctl.Start(startServices...).Enable().Run(); err != nil {
			log.Fatalf("error starting/enabling services: %v", err)
		}

		if err = snapctl.Unset("install-mode").Run(); err != nil {
			log.Fatalf("error un-setting 'install'; %v", err)
		}
	}

	log.Debug("End")
}

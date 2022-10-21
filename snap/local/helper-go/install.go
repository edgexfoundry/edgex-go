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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	hooks "github.com/canonical/edgex-snap-hooks/v2"
	"github.com/canonical/edgex-snap-hooks/v2/env"
	"github.com/canonical/edgex-snap-hooks/v2/log"
	"github.com/canonical/edgex-snap-hooks/v2/snapctl"
)

const secretStoreAddTokensCfg = "env.security-secret-store.add-secretstore-tokens"
const secretStoreAddKnownSecretsCfg = "env.security-secret-store.add-known-secrets"
const consulAddRegistryACLRolesCfg = "env.security-bootstrapper.add-registry-acl-roles"

// device-rest and device-virtual are both on the /cmd/security-file-token-provider/res/token-config.json file,
// so they should not need to be set here as well
var secretStoreTokens = []string{
	"app-functional-tests",
	"app-rules-engine",
	"app-http-export",
	"app-mqtt-export",
	"app-external-mqtt-trigger",
	"app-push-to-core",
	"app-rfid-llrp-inventory",
	"application-service",
	"device-camera",
	"device-mqtt",
	"device-modbus",
	"device-coap",
	"device-snmp",
	"device-gpio",
	"device-bacnet",
	"device-grove",
	"device-uart",
	"device-rfid-llrp",
	"device-usb-camera",
	"device-onvif-camera",
	"edgex-ekuiper",
}

var secretStoreKnownSecrets = []string{
	"redisdb[app-functional-tests]",
	"redisdb[app-rules-engine]",
	"redisdb[app-http-export]",
	"redisdb[app-mqtt-export]",
	"redisdb[app-external-mqtt-trigger]",
	"redisdb[app-push-to-core]",
	"redisdb[app-rfid-llrp-inventory]",
	"redisdb[application-service]",
	"redisdb[device-rest]",
	"redisdb[device-virtual]",
	"redisdb[device-camera]",
	"redisdb[device-mqtt]",
	"redisdb[device-modbus]",
	"redisdb[device-coap]",
	"redisdb[device-snmp]",
	"redisdb[device-gpio]",
	"redisdb[device-bacnet]",
	"redisdb[device-grove]",
	"redisdb[device-uart]",
	"redisdb[device-rfid-llrp]",
	"redisdb[device-usb-camera]",
	"redisdb[device-onvif-camera]",
	"redisdb[edgex-ekuiper]",
}

var (
	snapConf     = env.Snap + "/config"
	snapDataConf = env.SnapData + "/config"
)

// installConfFiles copies service configuration.toml files from $SNAP to $SNAP_DATA
func installConfFiles() error {
	var err error

	// services w/configuration that needs to be copied
	// to $SNAP_DATA
	var servicesWithConfig = []string{
		securityBootstrapper,
		securityBootstrapperRedis,
		securityFileTokenProvider,
		securityProxySetup,
		securitySecretStoreSetup,
		coreCommand,
		coreData,
		coreMetadata,
		supportNotifications,
		supportScheduler,
		systemManagementAgent,
		appServiceConfigurable,
	}

	for _, v := range servicesWithConfig {
		destDir := snapDataConf + "/"
		srcDir := snapConf + "/"

		// handle exceptions (i.e. config in non-std dirs)
		if v == securityBootstrapperRedis {
			destDir = destDir + "security-bootstrapper/res-bootstrap-redis"
			srcDir = srcDir + "security-bootstrapper/res-bootstrap-redis"
		} else if v == appServiceConfigurable {
			destDir = destDir + v + "/res/rules-engine"
			srcDir = srcDir + "/res/rules-engine"
		} else {
			destDir = destDir + v + "/res"
			srcDir = srcDir + v + "/res"
		}

		if err = os.MkdirAll(destDir, 0755); err != nil {
			return err
		}

		srcPath := srcDir + "/configuration.toml"
		destPath := destDir + "/configuration.toml"
		err = hooks.CopyFile(srcPath, destPath)
		if err != nil {
			return err
		}

		// copy additional files
		if v == coreMetadata {
			uomSrcPath := srcDir + "/uom.toml"
			uomDestPath := destDir + "/uom.toml"
			err = hooks.CopyFile(uomSrcPath, uomDestPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// installKuiper execs a shell script to install Kuiper's file into $SNAP_DATA
func installKuiper() error {
	// install files using edgex-ekuiper install hook
	filePath := env.Snap + "/snap.edgex-ekuiper/hooks/install"

	cmdSetupKuiper := exec.Cmd{
		Path:   filePath,
		Env:    append(os.Environ(), "KUIPER_BASE_KEY="+env.SnapData+"/kuiper"),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	err := cmdSetupKuiper.Run()
	if err != nil {
		return err
	}

	return nil
}

// installSecretStore: Steps 5, 8, 6, 11
func installSecretStore() error {
	var err error

	// Set the default values of
	//  ADD_KNOWN_SECRETS
	//	ADD_SECRETSTORE_TOKENS
	//	ADD_REGISTRY_ACL_ROLES
	// We do not have access to the snap configuration in the install hook,
	// so this just sets the values to the default list of services
	if err = snapctl.Set(secretStoreAddTokensCfg, strings.Join(secretStoreTokens, ",")).Run(); err != nil {
		return err
	}

	if err = snapctl.Set(secretStoreAddKnownSecretsCfg, strings.Join(secretStoreKnownSecrets, ",")).Run(); err != nil {
		return err
	}

	if err = os.MkdirAll(env.SnapData+"/secrets", 0700); err != nil {
		return err
	}

	path := "/security-file-token-provider/res/token-config.json"
	if err = hooks.CopyFile(snapConf+path, snapDataConf+path); err != nil {
		return err
	}

	// install the template config yaml file for securing Kong's admin
	// APIs in security-secretstore-setup service
	path = "/security-secretstore-setup/res/kong-admin-config.template.yml"
	err = hooks.CopyFile(snapConf+path, snapDataConf+path)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(snapDataConf+"/security-secret-store", 0755); err != nil {
		return err
	}

	path = "/security-secret-store/vault-config.hcl"
	destPath := snapDataConf + path
	if err = hooks.CopyFile(snapConf+path, destPath); err != nil {
		return err
	}

	if err = os.Chmod(destPath, 0644); err != nil {
		return err
	}

	return nil
}

// installConsul: step 7
func installConsul() error {
	var err error

	// Set the default value of ADD_REGISTRY_ACL_ROLES
	// using the same list of services as used in ADD_KNOWN_SECRETS
	// We do not have access to the snap configuration in the install hook,
	// so this just sets the values to the default list of services
	if err = snapctl.Set(consulAddRegistryACLRolesCfg, strings.Join(secretStoreTokens, ",")).Run(); err != nil {
		return err
	}

	if err = os.MkdirAll(env.SnapData+"/consul/config", 0755); err != nil {
		return err
	}

	if err = os.MkdirAll(env.SnapData+"/consul/data", 0755); err != nil {
		return err
	}

	return nil
}

// TODO: this function actually causes postgres to start in order
// to setup the security for postgres, thus we may need to move
// the install/setup logic for the proxy to the configure hook.
func setupPostgres() error {

	setupScriptPath, err := exec.LookPath("install-setup-postgres.sh")
	if err != nil {
		return err
	}

	cmdSetupPostgres := exec.Cmd{
		Path:   setupScriptPath,
		Args:   []string{setupScriptPath},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	if err = cmdSetupPostgres.Run(); err != nil {
		return err
	}

	return nil
}

// installProxy handles initialization of the API Gateway.
func installProxy() error {
	var err error

	if err = os.MkdirAll(env.SnapCommon+"/logs", 0755); err != nil {
		return err
	}

	if err = os.MkdirAll(snapDataConf+"/security-proxy-setup", 0755); err != nil {
		return err
	}

	// ensure prefix uses the 'current' symlink in it's path, otherwise refreshes to a
	// new snap revision will break
	snapDataCurr := strings.Replace(env.SnapData, env.SnapRev, "current", 1)
	rStrings := map[string]string{
		"#prefix = /usr/local/kong/":  "prefix = " + snapDataCurr + "/kong",
		"#nginx_user = nobody nobody": "nginx_user = root root",
	}

	path := "/security-proxy-setup/kong.conf"
	if err = hooks.CopyFileReplace(snapConf+path, snapDataConf+path, rStrings); err != nil {
		return err
	}

	if err = setupPostgres(); err != nil {
		return err
	}

	return nil
}

// This function creates the redis config dir under $SNAP_DATA,
// and creates an empty redis.conf file. This allows the command
// line for the service to always specify the config file, and
// allows for redis when the config option security-secret-store
// is "on" or "off".
func installRedis() error {
	fileName := filepath.Join(env.SnapData, "/redis/conf/redis.conf")
	if _, err := os.Stat(filepath.Join(env.SnapData, "redis")); err != nil {
		// dir doesn't exist
		if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(fileName, nil, 0644); err != nil {
			return err
		}
	}
	return nil
}

func install() {
	log.SetComponentName("install")

	var err error

	if err = installConfFiles(); err != nil {
		log.Fatalf("Error installing config files: %v", err)
	}

	if err = installKuiper(); err != nil {
		log.Fatalf("Error installing eKuiper: %v", err)
	}

	if err = installSecretStore(); err != nil {
		log.Fatalf("Error installing secret store: %v", err)
	}

	if err = installConsul(); err != nil {
		log.Fatalf("Error installing consul: %v", err)
	}

	if err = installProxy(); err != nil {
		log.Fatalf("Error installing proxy: %v", err)
	}

	if err = installRedis(); err != nil {
		log.Fatalf("Error installing redis: %v", err)
	}

	// Stop and disable all services as they will be
	// re-enabled in the configure hook if install-mode=defer-startup and
	// they have their state set to "on".
	var services []string
	if serviceMap, err := snapctl.Services().Run(); err != nil {
		log.Fatalf("Error getting list of services: %v", err)
	} else {
		for k := range serviceMap {
			services = append(services, k)
		}
	}

	log.Infof("Disabling all services to defer startup to after configuration: %v", services)
	if err = snapctl.Stop(services...).Disable().Run(); err != nil {
		log.Fatalf("Error disabling services: %v", err)
	}

	if err = snapctl.Set("install-mode", "defer-startup").Run(); err != nil {
		log.Fatalf("Error setting 'install-mode'; %v", err)
	}
}

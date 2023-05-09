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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	hooks "github.com/canonical/edgex-snap-hooks/v3"
	"github.com/canonical/edgex-snap-hooks/v3/env"
	"github.com/canonical/edgex-snap-hooks/v3/log"
	"github.com/canonical/edgex-snap-hooks/v3/snapctl"
)

// Default value of EDGEX_ADD_SECRETSTORE_TOKENS and EDGEX_ADD_REGISTRY_ACL_ROLES
// The device-rest and device-virtual are already set in /cmd/security-file-token-provider/res/token-config.json
var secretStoreTokens = []string{
	"app-functional-tests",
	"app-rules-engine",
	"app-http-export",
	"app-mqtt-export",
	"app-external-mqtt-trigger",
	"app-push-to-core",
	"app-rfid-llrp-inventory",
	"application-service",
	"device-mqtt",
	"device-modbus",
	"device-coap",
	"device-snmp",
	"device-gpio",
	"device-bacnet",
	"device-rfid-llrp",
	"device-usb-camera",
	"device-onvif-camera",
	"edgex-ekuiper",
}

// Default value of EDGEX_ADD_KNOWN_SECRETS
var secretStoreKnownSecrets = []string{
	"redisdb[device-rest]",
	"redisdb[device-virtual]",
	"redisdb[app-functional-tests]",
	"redisdb[app-rules-engine]",
	"redisdb[app-http-export]",
	"redisdb[app-mqtt-export]",
	"redisdb[app-external-mqtt-trigger]",
	"redisdb[app-push-to-core]",
	"redisdb[app-rfid-llrp-inventory]",
	"redisdb[application-service]",
	"redisdb[device-mqtt]",
	"redisdb[device-modbus]",
	"redisdb[device-coap]",
	"redisdb[device-snmp]",
	"redisdb[device-gpio]",
	"redisdb[device-bacnet]",
	"redisdb[device-rfid-llrp]",
	"redisdb[device-usb-camera]",
	"redisdb[device-onvif-camera]",
	"redisdb[edgex-ekuiper]",
}

var (
	snapConf     = env.Snap + "/config"
	snapDataConf = env.SnapData + "/config"
)

// installConfFiles copies service configuration files from $SNAP to $SNAP_DATA
func installConfFiles() error {
	var err error

	// services w/configuration that needs to be copied
	// to $SNAP_DATA
	var servicesWithConfig = []string{
		securitySecretsConfig,
		securityBootstrapper,
		securityBootstrapperRedis,
		securityFileTokenProvider,
		securityProxyAuth,
		securitySecretStoreSetup,
		coreCommonConfigBootstrapper,
		coreCommand,
		coreData,
		coreMetadata,
		supportNotifications,
		supportScheduler,
	}

	for _, v := range servicesWithConfig {
		destDir := snapDataConf + "/"
		srcDir := snapConf + "/"

		// handle exceptions (i.e. config in non-std dirs)
		if v == securityBootstrapperRedis {
			destDir = destDir + "security-bootstrapper/res-bootstrap-redis"
			srcDir = srcDir + "security-bootstrapper/res-bootstrap-redis"
		} else {
			destDir = destDir + v + "/res"
			srcDir = srcDir + v + "/res"
		}

		err = hooks.CopyDir(srcDir, destDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func installSecretStore() error {
	var err error

	// Set the default value of EDGEX_ADD_SECRETSTORE_TOKENS via snap option
	if err = snapctl.Set("apps.security-secretstore-setup.config.edgex-add-secretstore-tokens",
		strings.Join(secretStoreTokens, ",")).Run(); err != nil {
		return err
	}

	// Set the default value of EDGEX_ADD_KNOWN_SECRETS via snap option
	if err = snapctl.Set("apps.security-secretstore-setup.config.edgex-add-known-secrets",
		strings.Join(secretStoreKnownSecrets, ",")).Run(); err != nil {
		return err
	}

	if err = os.MkdirAll(env.SnapData+"/secrets", 0700); err != nil {
		return err
	}

	path := "/security-file-token-provider/res/token-config.json"
	if err = hooks.CopyFile(snapConf+path, snapDataConf+path); err != nil {
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

func installConsul() error {
	var err error

	// Set the default value of EDGEX_ADD_REGISTRY_ACL_ROLES via snap option
	// using the same list of services as used in EDGEX_ADD_KNOWN_SECRETS
	if err = snapctl.Set("apps.security-bootstrapper.config.edgex-add-registry-acl-roles",
		strings.Join(secretStoreTokens, ",")).Run(); err != nil {
		return err
	}

	if err = os.MkdirAll(env.SnapData+"/consul/data", 0755); err != nil {
		return err
	}

	if err = hooks.CopyDir(snapConf+"/consul", env.SnapData+"/consul/config"); err != nil {
		return err
	}

	return nil
}

// installProxy handles initialization of the API Gateway.
func installProxy() error {
	var err error

	if err = os.MkdirAll(env.SnapCommon+"/nginx/logs", 0755); err != nil {
		return err
	}

	if err = hooks.CopyDir(snapConf+"/nginx", env.SnapData+"/nginx"); err != nil {
		return err
	}

	return nil
}

// This function creates the redis config dir under $SNAP_DATA,
// and creates an empty redis.conf file. This allows the command
// line for the service to always specify the config file, and
// allows running redis with or without security config
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

}

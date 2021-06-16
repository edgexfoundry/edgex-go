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
	"os/exec"
	"strings"

	hooks "github.com/canonical/edgex-snap-hooks/v2"
)

var cli *hooks.CtlCli = hooks.NewSnapCtl()

const proxyRouteCfg = "env.security-proxy.add-proxy-route"
const secretStoreAddTokensCfg = "env.security-secret-store.add-secretstore-tokens"
const secretStoreAddKnownSecretsCfg = "env.security-secret-store.add-known-secrets"
const consulAddRegistryACLRolesCfg = "env.security-bootstrapper.add-registry-acl-roles"

// device-rest and device-virtual are both on the /cmd/security-file-token-provider/res/token-config.json file,
// so they should not need to be set here as well
const secretStoreTokens = "app-http-export,app-mqtt-export,device-camera,device-mqtt,device-modbus,device-coap,device-snmp"

const secretStoreKnownSecrets = "" +
	"redisdb[app-rules-engine]," +
	"redisdb[app-http-export]," +
	"redisdb[app-mqtt-export]," +
	"redisdb[device-rest]," +
	"redisdb[device-virtual]," +
	"redisdb[device-camera]," +
	"redisdb[device-mqtt]," +
	"redisdb[device-modbus]," +
	"redisdb[device-coap]," +
	"redisdb[device-snmp]"

var services = []string{
	"security-bootstrapper",
	"security-bootstrap-redis",
	"security-file-token-provider",
	"security-proxy-setup",
	"security-secretstore-setup",
	"core-command",
	"core-data",
	"core-metadata",
	"support-notifications",
	"support-scheduler",
	"sys-mgmt-agent",
	"device-virtual",
	"app-service-configurable",
}

// installConfFiles copies service configuration.toml files from $SNAP to $SNAP_DATA
func installConfFiles() error {
	var err error

	// TODO: should we continue to do this, since config overrides are
	// the preferred way to make configuration changes now?
	for _, v := range services {
		destDir := hooks.SnapDataConf + "/"
		srcDir := hooks.SnapConf + "/"

		// handle exceptions (i.e. config in non-std dirs)
		if v == "security-bootstrap-redis" {
			destDir = destDir + "security-bootstrapper/res-bootstrap-redis"
			srcDir = srcDir + "security-bootstrapper/res-bootstrap-redis"
		} else if v == "app-service-configurable" {
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
	}

	return nil
}

// installDevices installs device-virtual's static device files to $SNAP_DATA
func installDevices() error {
	var err error

	destDir := hooks.SnapDataConf + "/device-virtual/res/devices"
	destPath := destDir + "/devices.toml"
	srcPath := hooks.SnapConf + "/device-virtual/res/devices/devices.toml"

	if err = os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	if err = hooks.CopyFile(srcPath, destPath); err != nil {
		return err
	}

	return nil
}

// installDevProfiles installs device-virtual's device profiles to $SNAP_DATA
func installDevProfiles() error {
	var err error
	profiles := []string{"bool", "float", "int", "uint", "binary"}

	srcDir := hooks.SnapConf + "/device-virtual/res/profiles"
	destDir := hooks.SnapDataConf + "/device-virtual/res/profiles"
	if err = os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	for _, v := range profiles {
		fileName := "/device.virtual." + v + ".yaml"
		srcPath := srcDir + fileName
		destPath := destDir + fileName

		if err = hooks.CopyFile(srcPath, destPath); err != nil {
			return err
		}
	}

	return nil
}

// installKuiper execs a shell script to install Kuiper's file into $SNAP_DATA
func installKuiper() error {
	setupScriptPath, err := exec.LookPath("install-setup-kuiper.sh")
	if err != nil {
		return err
	}

	cmdSetupKuiper := exec.Cmd{
		Path:   setupScriptPath,
		Args:   []string{setupScriptPath},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	err = cmdSetupKuiper.Run()
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
	if err = cli.SetConfig(secretStoreAddTokensCfg, secretStoreTokens); err != nil {
		return err
	}

	if err = cli.SetConfig(secretStoreAddKnownSecretsCfg, secretStoreKnownSecrets); err != nil {
		return err
	}

	if err = os.MkdirAll(hooks.SnapData+"/secrets", 0700); err != nil {
		return err
	}

	path := "/security-file-token-provider/res/token-config.json"
	if err = hooks.CopyFile(hooks.SnapConf+path, hooks.SnapDataConf+path); err != nil {
		return err
	}

	// install the template config yaml file for securing Kong's admin
	// APIs in security-secretstore-setup service
	path = "/security-secretstore-setup/res/kong-admin-config.template.yml"
	err = hooks.CopyFile(hooks.SnapConf+path, hooks.SnapDataConf+path)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(hooks.SnapDataConf+"/security-secret-store", 0755); err != nil {
		return err
	}

	path = "/security-secret-store/vault-config.hcl"
	destPath := hooks.SnapDataConf + path
	if err = hooks.CopyFile(hooks.SnapConf+path, destPath); err != nil {
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
	if err = cli.SetConfig(consulAddRegistryACLRolesCfg, secretStoreTokens); err != nil {
		return err
	}

	if err = os.MkdirAll(hooks.SnapData+"/consul/config", 0755); err != nil {
		return err
	}

	if err = os.MkdirAll(hooks.SnapData+"/consul/data", 0755); err != nil {
		return err
	}

	return nil
}

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

	if err = os.MkdirAll(hooks.SnapCommon+"/logs", 0755); err != nil {
		return err
	}

	// the configuration is not available in the install hook, so we can only set it here
	// to the default value (consul).
	proxyRoute := "consul.http://localhost:8500"
	if err = cli.SetConfig(proxyRouteCfg, proxyRoute); err != nil {
		return err
	}

	if err = os.MkdirAll(hooks.SnapDataConf+"/security-proxy-setup", 0755); err != nil {
		return err
	}

	// ensure prefix uses the 'current' symlink in it's path, otherwise refreshes to a
	// new snap revision will break
	snapDataCurr := strings.Replace(hooks.SnapData, hooks.SnapRev, "current", 1)
	rStrings := map[string]string{
		"#prefix = /usr/local/kong/":  "prefix = " + snapDataCurr + "/kong",
		"#nginx_user = nobody nobody": "nginx_user = root root",
	}

	path := "/security-proxy-setup/kong.conf"
	if err = hooks.CopyFileReplace(hooks.SnapConf+path, hooks.SnapDataConf+path, rStrings); err != nil {
		return err
	}

	if err = setupPostgres(); err != nil {
		return err
	}

	return nil
}

func installRedis() error {
	if err := os.MkdirAll(hooks.SnapData+"/redis", 0755); err != nil {
		return err
	}

	return nil
}

func main() {
	var debug = false
	var err error

	status, err := cli.Config("debug")
	if err != nil {
		fmt.Println(fmt.Sprintf("edgexfoundry:install: can't read value of 'debug': %v", err))
		os.Exit(1)
	}
	if status == "true" {
		debug = true
	}

	if err = hooks.Init(debug, "edgexfoundry"); err != nil {
		fmt.Println(fmt.Sprintf("edgexfoundry:install: initialization failure: %v", err))
		os.Exit(1)

	}

	if err = installConfFiles(); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:install: %v", err))
		os.Exit(1)
	}

	if err = installDevices(); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:install: %v", err))
		os.Exit(1)
	}

	if err = installDevProfiles(); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:install: %v", err))
		os.Exit(1)
	}

	if err = installKuiper(); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:install: %v", err))
		os.Exit(1)
	}

	if err = installSecretStore(); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:install: %v", err))
		os.Exit(1)
	}

	if err = installConsul(); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:install: %v", err))
		os.Exit(1)
	}

	if err = installProxy(); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:install %v", err))
		os.Exit(1)
	}

	if err = installRedis(); err != nil {
		hooks.Error(fmt.Sprintf("edgexfoundry:install %v", err))
		os.Exit(1)
	}

	// just like the configure hook, this code needs to iterate over every service
	// and setup .env files for each if required...
	for _, v := range hooks.Services {
		serviceCfg := hooks.EnvConfig + "." + v
		envJSON, err := cli.Config(serviceCfg)
		if err != nil {
			hooks.Error(fmt.Sprintf("edgexfoundry:install failed to read service %s configuration - %v", v, err))
			os.Exit(1)
		}

		if envJSON != "" {
			hooks.Debug(fmt.Sprintf("edgexfoundry:install: service envJSON: %s", envJSON))
			if err := hooks.HandleEdgeXConfig(v, envJSON, nil); err != nil {
				hooks.Error(fmt.Sprintf("edgexfoundry:install failed to process service %s configuration - %v", v, err))
				os.Exit(1)
			}
		}
	}
}

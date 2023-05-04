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

// installConfFiles copies service configuration.yaml files from $SNAP to $SNAP_DATA
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

// installSecretStore: Steps 5, 8, 6, 11
func installSecretStore() error {
	var err error

	// Set the default values of
	//  EDGEX_ADD_KNOWN_SECRETS
	//	EDGEX_ADD_SECRETSTORE_TOKENS
	//	EDGEX_ADD_REGISTRY_ACL_ROLES
	// We do not have access to the snap configuration in the install hook,
	// so this just sets the values to the default list of services
	if err = snapctl.Set("apps.security-secretstore-setup.config.edgex-add-secretstore-tokens",
		strings.Join(secretStoreTokens, ",")).Run(); err != nil {
		return err
	}

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

// installConsul: step 7
func installConsul() error {
	var err error

	// Set the default value of EDGEX_ADD_REGISTRY_ACL_ROLES
	// using the same list of services as used in EDGEX_ADD_KNOWN_SECRETS
	// We do not have access to the snap configuration in the install hook,
	// so this just sets the values to the default list of services
	if err = snapctl.Set("apps.security-bootstrapper.config.edgex-add-registry-acl-roles",
		strings.Join(secretStoreTokens, ",")).Run(); err != nil {
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

// installProxy handles initialization of the API Gateway.
func installProxy() error {
	var err error

	if err = os.MkdirAll(env.SnapCommon+"/nginx/logs", 0755); err != nil {
		return err
	}

	// tmpdirs := []string{"body", "fastcgi", "proxy", "scgi"}

	// for _, d := range tmpdirs {
	// 	if err = os.MkdirAll(fmt.Sprintf("%s/nginx_tmp_%s", env.SnapCommon, d), 0750); err != nil {
	// 		return err
	// 	}
	// }

	if err = os.MkdirAll(filepath.Join(env.SnapData, "/nginx", "conf.d"), 0750); err != nil {
		return err
	}

	// ensure prefix uses the 'current' symlink in it's path, otherwise refreshes to a
	// new snap revision will break
	//snapDataCurr := strings.Replace(env.SnapData, env.SnapRev, "current", 1)
	rStrings := map[string]string{
		/*
			"include /etc/nginx/conf.d/*.conf;":                    "include " + snapDataCurr + "/nginx/*.conf;",
			"include /etc/nginx/sites-enabled/*;":                  "include " + snapDataCurr + "/nginx/sites-enabled/*;",
			"include /etc/nginx/modules-enabled/*.conf;":           "include " + snapDataCurr + "/nginx/modules-enabled/*.conf;",
			"client_body_temp_path /var/lib/nginx/body;":           "client_body_temp_path " + env.SnapCommon + "/nginx_tmp_body;",
			"fastcgi_temp_path /var/lib/nginx/fastcgi;":            "fastcgi_temp_path " + env.SnapCommon + "/nginx_tmp_fastcgi;",
			"proxy_temp_path /var/lib/nginx/proxy;":                "proxy_temp_path " + env.SnapCommon + "/nginx_tmp_proxy;",
			"scgi_temp_path /var/lib/nginx/scgi;":                  "scgi_temp_path " + env.SnapCommon + "/nginx_tmp_scgi;",
			"uwsgi_temp_path /var/lib/nginx/uwsgi;":                "uwsgi_temp_path " + env.SnapCommon + "/nginx_tmp_uwsgi;",
			"include /etc/nginx/mime.types;":                       "include " + env.Snap + "/etc/nginx/mime.types;",
			"access_log /var/log/nginx/access.log;":                "access_log " + env.SnapCommon + "/logs/nginx_access.log;",
			"error_log /var/log/nginx/error.log;":                  "error_log " + env.SnapCommon + "/logs/nginx_error.log;",
			"ssl_certificate \"/etc/ssl/nginx/nginx.crt\";":        "ssl_certificate \"" + snapDataCurr + "/nginx/nginx.crt\";",
			"ssl_certificate_key \"/etc/ssl/nginx/nginx.key\";":    "ssl_certificate_key \"" + snapDataCurr + "/nginx/nginx.key\";",
			"include /etc/nginx/conf.d/edgex-custom-rewrites.inc;": "include " + snapDataCurr + "/nginx/edgex-custom-rewrites.inc;",
		*/
	}

	path := "nginx.conf"
	if err = hooks.CopyFileReplace(filepath.Join(snapConf, "nginx", path), filepath.Join(env.SnapData, "nginx", path), rStrings); err != nil {
		return err
	}

	path = "edgex-default.conf"
	if err = hooks.CopyFileReplace(filepath.Join(snapConf, "nginx", path), filepath.Join(env.SnapData, "nginx/conf.d", path), rStrings); err != nil {
		return err
	}

	path = "edgex-custom-rewrites.inc"
	if err = hooks.CopyFileReplace(filepath.Join(snapConf, "nginx", path), filepath.Join(env.SnapData, "nginx/conf.d", path), rStrings); err != nil {
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

	// Enable autostart so that services start by default after seeding configuration
	// Set the option for each app instead of globally (i.e. autostart=true), so
	// that the option can be selectively changed for oneshot services after starting
	// them once!
	var autostartKeyValues []string
	for _, s := range allServices() {
		autostartKeyValues = append(autostartKeyValues, "apps."+s+".autostart", "true")
	}
	if err = snapctl.Set(autostartKeyValues...).Run(); err != nil {
		log.Fatalf("Error setting snap option: %v", err)
	}
}

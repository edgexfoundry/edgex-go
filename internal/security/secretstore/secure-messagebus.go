/*******************************************************************************
 * Copyright 2021 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package secretstore

import (
	"errors"
	"fmt"
	"os"
	"text/template"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
)

const (
	eKuiperEdgeXSourceTemplate = `
application_conf:
  port: 5571
  protocol: tcp
  server: localhost
  topic: application
default:
  optional:
    Username: {{.User}}
    Password: {{.Password}}
  port: 6379
  protocol: redis
  server: localhost
  connectionSelector: edgex.redisMsgBus
  topic: rules-events
  type: redis
mqtt_conf:
  optional:
    ClientId: client1
  port: 1883
  protocol: tcp
  server: localhost
  topic: events
  type: mqtt
`

	eKuiperConnectionsTemplate = `
edgex:
  redisMsgBus: #connection key
    protocol: redis
    server: localhost
    port: 6379
    type: redis
    optional:
      Username: {{.User}}
      Password: {{.Password}}
`
	// Can't use constants from go-mod-messaging since that will create ZMQ dependency, which we do not want!
	redisSecureMessageBusType = "redis"
	mqttSecureMessageBusType  = "mqtt"
	noneSecureMessageBusType  = "none"
	blankSecureMessageBusType = ""
)

func ConfigureSecureMessageBus(secureMessageBus config.SecureMessageBusInfo, redis5Pair UserPasswordPair, lc logger.LoggingClient) error {
	switch secureMessageBus.Type {
	// Currently, only support Secure MessageBus when using the Redis implementation.
	case redisSecureMessageBusType:
		// eKuiper now has two configuration files (EdgeX Sources and Connections)

		err := configureKuiperForSecureMessageBus(redis5Pair, "EdgeX Source", eKuiperEdgeXSourceTemplate, secureMessageBus.KuiperConfigPath, lc)
		if err != nil {
			return err
		}

		err = configureKuiperForSecureMessageBus(redis5Pair, "Connections", eKuiperConnectionsTemplate, secureMessageBus.KuiperConnectionsPath, lc)
		if err != nil {
			return err
		}

	// TODO: Add support for secure MQTT MessageBus
	case mqttSecureMessageBusType:
		return fmt.Errorf("secure MQTT MessageBus not yet supported")

	case noneSecureMessageBusType, blankSecureMessageBusType:
		return nil

	default:
		return fmt.Errorf("invalid Secure MessageBus Type of '%s'", secureMessageBus.Type)
	}

	return nil
}

func configureKuiperForSecureMessageBus(credentials UserPasswordPair, fileType string, fileTemplate string, path string, lc logger.LoggingClient) error {
	// This capability depends on the eKuiper file existing, which depends on the version of eKuiper installed.
	// If the file doesn't exist, then the eKuiper version installed doesn't use it, so skip the injection.
	_, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		lc.Infof("eKuiper file %s doesn't exist, skipping Secure MessageBus credentials injection", path)
		return nil
	}

	tmpl, err := template.New("eKuiper").Parse(fileTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse eKuiper %s template: %w", fileType, err)
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open/create eKuiper %s file %s: %w", fileType, path, err)
	}

	defer func() {
		_ = file.Close()
	}()

	err = tmpl.Execute(file, credentials)
	if err != nil {
		return fmt.Errorf("failed to write eKuiper  %s file %s: %w", fileType, path, err)
	}

	lc.Infof("Wrote eKuiper %s at %s with Secure MessageBus credentials", fileType, path)

	return nil
}

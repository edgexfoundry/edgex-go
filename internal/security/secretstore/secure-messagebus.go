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
	"fmt"
	"os"
	"text/template"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	kuiperConfigTemplate = `
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
  server: edgex-redis
  serviceServer: http://localhost:48080
  topic: rules-events
  type: redis
mqtt_conf:
  optional:
    ClientId: client1
  port: 1883
  protocol: tcp
  server: 127.0.0.1
  topic: events
  type: mqtt
`
	// Can't use constants from go-mod-messaging since that will create ZMQ dependency, which we do not want!
	redisSecureMessageBusType = "redis"
	mqttSecureMessageBusType  = "mqtt"
	noneSecureMessageBusType  = "none"
	blankSecureMessageBusType = ""
)

func ConfigureSecureMessageBus(secureMessageBus config.SecureMessageBusInfo, redis5Pair UserPasswordPair, lc logger.LoggingClient) error {
	switch secureMessageBus.Type {
	// Currently only support Secure MessageBus when using the Redis implementation
	case redisSecureMessageBusType:
		err := configureKuiperForSecureMessageBus(redis5Pair, secureMessageBus.KuiperConfigPath, lc)
		if err != nil {
			return err
		}

	// TODO: Add support for secure MQTT MessageBus
	case mqttSecureMessageBusType:
		return fmt.Errorf("Secure MQTT MessageBus not yet supported")

	case noneSecureMessageBusType, blankSecureMessageBusType:
		return nil

	default:
		return fmt.Errorf("Invalid Secure MessageBus Type of '%s'", secureMessageBus.Type)
	}

	return nil
}

func configureKuiperForSecureMessageBus(credentials UserPasswordPair, configPath string, lc logger.LoggingClient) error {
	tmpl, err := template.New("kuiper").Parse(kuiperConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse Kuiper Edgex config template: %w", err)
	}

	file, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open/create Kuiper Edgex config file %s: %w", configPath, err)
	}

	defer func() {
		_ = file.Close()
	}()

	err = tmpl.Execute(file, credentials)
	if err != nil {
		return fmt.Errorf("failed to write Kuiper Edgex config file %s: %w", configPath, err)
	}

	lc.Infof("Wrote Kuiper config at %s with secure MessageBus credentials", configPath)

	return nil
}

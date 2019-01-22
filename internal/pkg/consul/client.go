/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
 *******************************************************************************/

package consulclient

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/mitchellh/consulstructure"
	"strconv"
	"github.com/edgexfoundry/edgex-go/internal/pkg/registry/types"
)

// Configuration struct for consul - used to initialize the service
type ConsulConfig struct {
	ConsulAddress  string
	ConsulPort     int
	ServiceName    string
	ServiceAddress string
	ServicePort    int
	CheckAddress   string
	CheckInterval  string
}

var consul *consulapi.Client = nil // Call consulInit to initialize this variable

func NewConsulConfig(reg config.RegistryInfo, svc config.ServiceInfo, key string) ConsulConfig {
	c := ConsulConfig{
		ServiceName:    key,
		ServicePort:    svc.Port,
		ServiceAddress: svc.Host,
		CheckAddress:   svc.HealthCheck(),
		CheckInterval:  svc.CheckInterval,
		ConsulAddress:  reg.Host,
		ConsulPort:     reg.Port,
	}
	return c
}

func NewConsulDecoder(reg config.RegistryInfo) *consulstructure.Decoder {
	cfg := &consulapi.Config{}
	cfg.Address = reg.Host + ":" + strconv.Itoa(reg.Port)
	d := &consulstructure.Decoder{
		Consul: cfg,
	}
	return d
}

// Initialize consul by connecting to the agent and registering the service/check
func ConsulInit(config ConsulConfig) error {
	var err error // Declare error to be used throughout function

	// Connect to the Consul Agent
	defaultConfig := &consulapi.Config{}
	defaultConfig.Address = config.ConsulAddress + ":" + strconv.Itoa(config.ConsulPort)
	consul, err = consulapi.NewClient(defaultConfig)
	if err != nil {
		return err
	}

	// Register the Service
	err = consul.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
		Name:    config.ServiceName,
		Address: config.ServiceAddress,
		Port:    config.ServicePort,
	})
	if err != nil {
		return err
	}

	// Register the Health Check
	err = consul.Agent().CheckRegister(&consulapi.AgentCheckRegistration{
		Name:      "Health Check: " + config.ServiceName,
		Notes:     "Check the health of the API",
		ServiceID: config.ServiceName,
		AgentServiceCheck: consulapi.AgentServiceCheck{
			HTTP:     config.CheckAddress,
			Interval: config.CheckInterval,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func GetServiceEndpoint(serviceKey string) (types.ServiceEndpoint, error) {
	services, err := consul.Agent().Services()
	if err != nil {
		return types.ServiceEndpoint{}, err
	}

	endpoint := types.ServiceEndpoint{}
	for key, service := range services {
		if key == serviceKey {
			endpoint.Port = service.Port
			endpoint.Key = key
			endpoint.Address = service.Address
		}
	}
	return endpoint, nil
}

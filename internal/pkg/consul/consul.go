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
	"errors"
	"reflect"
	"strconv"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
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

type ServiceEndpoint struct {
	Key     string
	Address string
	Port    int
}

var consul *consulapi.Client = nil // Call consulInit to initialize this variable

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

func GetServiceEndpoint(serviceKey string) (ServiceEndpoint, error) {
	services, err := consul.Agent().Services()
	if err != nil {
		return ServiceEndpoint{}, err
	}

	endpoint := ServiceEndpoint{}
	for key, service := range services {
		if key == serviceKey {
			endpoint.Port = service.Port
			endpoint.Key = key
			endpoint.Address = service.Address
		}
	}
	return endpoint, nil
}

// Look at the key/value pairs to update configuration
func CheckKeyValuePairs(configurationStruct interface{}, applicationName string, profiles []string) error {
	// Consul wasn't initialized
	if consul == nil {
		err := errors.New("Consul wasn't initialized, can't check key/value pairs")
		return err
	}

	kv := consul.KV()

	// Reflection to get the field names (These will be part of the key names)
	configValue := reflect.ValueOf(configurationStruct)
	// Loop through the fields
	for i := 0; i < configValue.Elem().NumField(); i++ {
		fieldName := configValue.Elem().Type().Field(i).Name
		fieldValue := configValue.Elem().Field(i)
		keyPath := "config/" + applicationName + ";" + strings.Join(profiles, ";") + "/" + fieldName
		var byteValue []byte // Byte array that will be passed to Consul

		// Switch off of the value type
		switch fieldValue.Kind() {
		case reflect.Bool:
			byteValue = []byte(strconv.FormatBool(fieldValue.Bool()))

			// Check if the key is already there
			pair, _, err := kv.Get(keyPath, nil)
			if err != nil {
				return err
			}
			// Pair doesn't exist, create it
			if pair == nil {
				pair = &consulapi.KVPair{
					Key:   keyPath,
					Value: byteValue,
				}
				_, err = kv.Put(pair, nil)
				if err != nil {
					return err
				}
			} else { // Pair does exist, get the new value
				pair, _, err = kv.Get(keyPath, nil)
				if err != nil {
					return err
				}

				newValue, err := strconv.ParseBool(string(pair.Value))
				if err != nil {
					return err
				}

				fieldValue.SetBool(newValue) // Set the new value
			}
			break
		case reflect.String:
			byteValue = []byte(fieldValue.String())

			// Check if the key is already there
			pair, _, err := kv.Get(keyPath, nil)
			if err != nil {
				return err
			}
			// Pair doesn't exist, create it
			if pair == nil {
				pair = &consulapi.KVPair{
					Key:   keyPath,
					Value: byteValue,
				}
				_, err = kv.Put(pair, nil)
				if err != nil {
					return err
				}
			} else { // Pair does exist, get the new value
				pair, _, err = kv.Get(keyPath, nil)
				if err != nil {
					return err
				}

				newValue := string(pair.Value)

				fieldValue.SetString(newValue) // Set the new value
			}
			break
		case reflect.Int:
			byteValue = []byte(strconv.FormatInt(fieldValue.Int(), 10))

			// Check if the key is already there
			pair, _, err := kv.Get(keyPath, nil)
			if err != nil {
				return err
			}
			// Pair doesn't exist, create it
			if pair == nil {
				pair = &consulapi.KVPair{
					Key:   keyPath,
					Value: byteValue,
				}
				_, err = kv.Put(pair, nil)
				if err != nil {
					return err
				}
			} else { // Pair does exist, get the new value
				pair, _, err = kv.Get(keyPath, nil)
				if err != nil {
					return err
				}

				newValue, err := strconv.ParseInt(string(pair.Value), 10, 64)
				if err != nil {
					return err
				}

				fieldValue.SetInt(newValue) // Set the new value
			}
			break
		default:
			err := errors.New("Can't get the type of field: " + keyPath)
			return err
		}
	}

	return nil
}

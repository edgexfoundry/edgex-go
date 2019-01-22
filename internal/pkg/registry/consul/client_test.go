//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package consul

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/hashicorp/consul/api"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/edgexfoundry/edgex-go/internal/pkg/registry/types"
)

const (
	serviceName        = "consulUnitTest"
	serviceHost        = "localhost"
	defaultServicePort = 8000
	consulBasePath     = internal.ConfigRegistryStem + serviceName + "/"
)

// change values to localhost and 8500 if you need to run tests against real Consul service running locally
var (
	testHost = ""
	port     = 0
)

type MyConfig struct {
	Logging  config.LoggingInfo
	Service  types.ServiceEndpoint
	Port     int
	Host     string
	LogLevel string
}

func TestMain(m *testing.M) {

	var testMockServer *httptest.Server
	if testHost == "" || port != 8500 {
		mockConsul := NewMockConsul()
		testMockServer = mockConsul.Start()

		URL, _ := url.Parse(testMockServer.URL)
		testHost = URL.Hostname()
		port, _ = strconv.Atoi(URL.Port())
	}

	exitCode := m.Run()
	if testMockServer != nil {
		defer testMockServer.Close()
	}
	os.Exit(exitCode)
}

func TestRegistryRunning(t *testing.T) {
	client := makeConsulClient(t, defaultServicePort, true)
	if !client.IsRegistryRunning() {
		t.Fatal("Consul not running")
	}
}

func TestRegisterWithPingCallback(t *testing.T) {
	doneChan := make(chan bool)
	expectedHealthCheckPath := "api/v1/ping"
	receivedPing := false

	// Setup a server to simulate the service for the health check callback
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.URL.Path, expectedHealthCheckPath) {

			switch request.Method {
			case "GET":
				fmt.Println("Received ping")
				receivedPing = true

				writer.Header().Set("Content-Type", "text/plain")
				writer.Write([]byte("pong"))

				doneChan <- true
			}
		}
	}))
	defer server.Close()

	fmt.Println("Click this link if running actual Consul in docker container which unable to route back to real localhost:  " + server.URL)

	// Figure out which port the simulated service is running on.
	serverUrl, _ := url.Parse(server.URL)
	serverPort, _ := strconv.Atoi(serverUrl.Port())

	client := makeConsulClient(t, serverPort, true)
	// Make sure service is not already registered.
	client.consulClient.Agent().ServiceDeregister(client.ServiceName)
	client.consulClient.Agent().CheckDeregister(client.ServiceName)

	// Try to clean-up after test
	defer func(client *consulClient) {
		client.consulClient.Agent().ServiceDeregister(client.ServiceName)
		client.consulClient.Agent().CheckDeregister(client.ServiceName)
	}(client)

	// Register the service endpoint and health check callback
	err := client.Register()
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	go func() {
		time.Sleep(10 * time.Second)
		doneChan <- false
	}()

	<-doneChan
	assert.True(t, receivedPing, "Never received health check ping")
}

func TestGetServiceEndpoint(t *testing.T) {
	expectedNotFoundEndpoint := types.ServiceEndpoint{}
	expectedFoundEndpoint := types.ServiceEndpoint{
		Key:     serviceName,
		Address: serviceHost,
		Port:    defaultServicePort,
	}

	client := makeConsulClient(t, defaultServicePort, true)
	// Make sure service is not already registered.
	client.consulClient.Agent().ServiceDeregister(client.ServiceName)
	client.consulClient.Agent().CheckDeregister(client.ServiceName)

	// Try to clean-up after test
	defer func(client *consulClient) {
		client.consulClient.Agent().ServiceDeregister(client.ServiceName)
		client.consulClient.Agent().CheckDeregister(client.ServiceName)
	}(client)

	// Test for endpoint not found
	actualEndpoint, err := client.GetServiceEndpoint(client.ServiceName)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, expectedNotFoundEndpoint, actualEndpoint, "Test for endpoint not found result not as expected") {
		t.Fatal()
	}

	// Register the service endpoint
	err = client.Register()
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	// Test endpoint found
	actualEndpoint, err = client.GetServiceEndpoint(client.ServiceName)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, expectedFoundEndpoint, actualEndpoint, "Test for endpoint found result not as expected") {
		t.Fatal()
	}
}

func TestIsServiceAvailable(t *testing.T) {

	expected := false

	client := makeConsulClient(t, defaultServicePort, true)
	// Make sure service is not already registered.
	client.consulClient.Agent().ServiceDeregister(client.ServiceName)
	client.consulClient.Agent().CheckDeregister(client.ServiceName)

	// Try to clean-up after test
	defer func(client *consulClient) {
		client.consulClient.Agent().ServiceDeregister(client.ServiceName)
		client.consulClient.Agent().CheckDeregister(client.ServiceName)
	}(client)

	// Test before registering
	actual := client.IsServiceAvailable(client.ServiceName)
	if !assert.Equal(t, expected, actual, "IsServiceAvailable result not as expected") {
		t.Fatal()
	}

	// Register the service endpoint
	err := client.Register()
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	// Test before registering
	expected = true
	actual = client.IsServiceAvailable(client.ServiceName)
	if !assert.Equal(t, expected, actual, "IsServiceAvailable result not as expected") {
		t.Fatal()
	}
}

func TestRegisterNoServiceInfoError(t *testing.T) {
	// Don't set the service info so check for info results in error
	client := makeConsulClient(t, defaultServicePort, false)

	err := client.Register()
	if !assert.Error(t, err, "Expected error due to no service info") {
		t.Fatal()
	}
}

func TestConfigurationValueExists(t *testing.T) {
	key := "Foo"
	value := []byte("bar")
	fullKey := consulBasePath + key

	client := makeConsulClient(t, defaultServicePort, true)
	expected := false

	// Make sure the target key/value doesn't already exists
	_, err := client.consulClient.KV().Delete(fullKey, nil)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	actual, err := client.ConfigurationValueExists(key)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, expected, actual) {
		t.Fatal()
	}

	keyPair := api.KVPair{
		Key:   fullKey,
		Value: value,
	}

	_, err = client.consulClient.KV().Put(&keyPair, nil)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	expected = true
	actual, err = client.ConfigurationValueExists(key)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, expected, actual) {
		t.Fatal()
	}
}

func TestGetConfigurationValue(t *testing.T) {
	key := "Foo"
	expected := []byte("bar")
	fullKey := consulBasePath + key
	client := makeConsulClient(t, defaultServicePort, true)

	// Make sure the target key/value exists
	keyPair := api.KVPair{
		Key:   fullKey,
		Value: expected,
	}

	_, err := client.consulClient.KV().Put(&keyPair, nil)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	actual, err := client.GetConfigurationValue(key)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, expected, actual) {
		t.Fatal()
	}
}

func TestPutConfigurationValue(t *testing.T) {
	key := "Foo"
	expected := []byte("bar")
	expectedFullKey := consulBasePath + key
	client := makeConsulClient(t, defaultServicePort, true)

	//clean up the the key, if it exists
	client.consulClient.KV().Delete(expectedFullKey, nil)

	err := client.PutConfigurationValue(key, expected)
	assert.NoError(t, err)

	keyValue, _, err := client.consulClient.KV().Get(expectedFullKey, nil)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NotNil(t, keyValue, "%s value not found", expectedFullKey) {
		t.Fatal()
	}

	actual := keyValue.Value

	assert.Equal(t, expected, actual)

}

func TestGetConfiguration(t *testing.T) {
	expected := MyConfig{
		Logging: config.LoggingInfo{
			EnableRemote: true,
			File:         "NONE",
		},
		Service: types.ServiceEndpoint{
			Key:     "Dummy",
			Address: "10.6.7.8",
			Port:    8080,
		},
		Port:     8000,
		Host:     "localhost",
		LogLevel: "debug",
	}

	client := makeConsulClient(t, defaultServicePort, true)

	client.PutConfigurationValue("Logging/EnableRemote", []byte(strconv.FormatBool(expected.Logging.EnableRemote)))
	client.PutConfigurationValue("Logging/File", []byte(expected.Logging.File))
	client.PutConfigurationValue("Service/Key", []byte(expected.Service.Key))
	client.PutConfigurationValue("Service/Address", []byte(expected.Service.Address))
	client.PutConfigurationValue("Service/Port", []byte(strconv.Itoa(expected.Service.Port)))
	client.PutConfigurationValue("Port", []byte(strconv.Itoa(expected.Port)))
	client.PutConfigurationValue("Host", []byte(expected.Host))
	client.PutConfigurationValue("LogLevel", []byte(expected.LogLevel))

	result, err := client.GetConfiguration(&MyConfig{})
	configuration := result.(*MyConfig)

	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NotNil(t, configuration) {
		t.Fatal()
	}

	assert.Equal(t, expected.Logging.EnableRemote, configuration.Logging.EnableRemote, "Logging.EnableRemote not as expected")
	assert.Equal(t, expected.Logging.File, configuration.Logging.File, "Logging.File not as expected")
	assert.Equal(t, expected.Service.Port, configuration.Service.Port, "Service.Port not as expected")
	assert.Equal(t, expected.Service.Address, configuration.Service.Address, "Service.Address not as expected")
	assert.Equal(t, expected.Service.Key, configuration.Service.Key, "Service.Key not as expected")
	assert.Equal(t, expected.Port, configuration.Port, "Port not as expected")
	assert.Equal(t, expected.Host, configuration.Host, "Host not as expected")
	assert.Equal(t, expected.LogLevel, configuration.LogLevel, "LogLevel not as expected")
}

func TestPutConfigurationNoPreviousValues(t *testing.T) {
	client := makeConsulClient(t, defaultServicePort, true)

	// Make sure the tree of values doesn't exist.
	client.consulClient.KV().DeleteTree(consulBasePath, nil)

	defer func() {
		// Clean up
		client.consulClient.KV().DeleteTree(consulBasePath, nil)
	}()

	configMap := createKeyValueMap()
	configuration, err := toml.TreeFromMap(configMap)
	if err != nil {
		log.Fatalf("unable to create TOML Tree from map: %v", err)
	}
	err = client.PutConfiguration(configuration, false)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	keyValues := convertInterfaceToConsulPairs("", configMap)
	for _, keyValue := range keyValues {
		expected := string(keyValue.Value)
		value, err := client.GetConfigurationValue(keyValue.Key)
		if !assert.NoError(t, err) {
			t.Fatal()
		}
		actual := string(value)
		if !assert.Equal(t, expected, actual, "Values for %s are not equal", keyValue.Key) {
			t.Fatal()
		}
	}
}

func TestPutConfigurationWithoutOverWrite(t *testing.T) {
	client := makeConsulClient(t, defaultServicePort, true)

	// Make sure the tree of values doesn't exist.
	client.consulClient.KV().DeleteTree(consulBasePath, nil)

	defer func() {
		// Clean up
		client.consulClient.KV().DeleteTree(consulBasePath, nil)
	}()

	configMap := createKeyValueMap()

	configuration, _ := toml.TreeFromMap(configMap)
	err := client.PutConfiguration(configuration, false)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	//Update map with new value and try to overwrite it
	configMap["int"] = 2
	configMap["int64"] = 164
	configMap["float64"] = 2.4
	configMap["string"] = "bye"
	configMap["bool"] = false

	// Try to put new values with overwrite = false
	configuration, _ = toml.TreeFromMap(configMap)
	err = client.PutConfiguration(configuration, false)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	keyValues := convertInterfaceToConsulPairs("", configMap)
	for _, keyValue := range keyValues {
		expected := string(keyValue.Value)
		value, err := client.GetConfigurationValue(keyValue.Key)
		if !assert.NoError(t, err) {
			t.Fatal()
		}
		actual := string(value)
		if !assert.NotEqual(t, expected, actual, "Values for %s are equal, expected not equal", keyValue.Key) {
			t.Fatal()
		}
	}
}

func TestPutConfigurationOverWrite(t *testing.T) {
	client := makeConsulClient(t, defaultServicePort, true)

	// Make sure the tree of values doesn't exist.
	client.consulClient.KV().DeleteTree(consulBasePath, nil)
	// Clean up after unit test
	defer func() {
		client.consulClient.KV().DeleteTree(consulBasePath, nil)
	}()

	configMap := createKeyValueMap()

	configuration, _ := toml.TreeFromMap(configMap)
	err := client.PutConfiguration(configuration, false)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	//Update map with new value and try to overwrite it
	configMap["int"] = 2
	configMap["float64"] = 2.4
	configMap["string"] = "bye"
	configMap["bool"] = false

	// Try to put new values with overwrite = True
	configuration, _ = toml.TreeFromMap(configMap)
	err = client.PutConfiguration(configuration, true)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	keyValues := convertInterfaceToConsulPairs("", configMap)
	for _, keyValue := range keyValues {
		expected := string(keyValue.Value)
		value, err := client.GetConfigurationValue(keyValue.Key)
		if !assert.NoError(t, err) {
			t.Fatal()
		}
		actual := string(value)
		if !assert.Equal(t, expected, actual, "Values for %s are not equal", keyValue.Key) {
			t.Fatal()
		}
	}
}

func TestWatchForChanges(t *testing.T) {
	expectedConfig := MyConfig{
		Logging: config.LoggingInfo{
			EnableRemote: true,
			File:         "NONE",
		},
		Service: types.ServiceEndpoint{
			Key:     "Dummy",
			Address: "10.6.7.8",
			Port:    8080,
		},
		Port:     8000,
		Host:     "localhost",
		LogLevel: "debug",
	}

	expectedChange := "random"

	client := makeConsulClient(t, defaultServicePort, false)

	// Make sure the tree of values doesn't exist.
	client.consulClient.KV().DeleteTree(consulBasePath, nil)
	// Clean up after unit test
	defer func() {
		client.consulClient.KV().DeleteTree(consulBasePath, nil)
	}()

	client.PutConfigurationValue("Logging/EnableRemote", []byte(strconv.FormatBool(expectedConfig.Logging.EnableRemote)))
	client.PutConfigurationValue("Logging/File", []byte(expectedConfig.Logging.File))
	client.PutConfigurationValue("Service/Key", []byte(expectedConfig.Service.Key))
	client.PutConfigurationValue("Service/Address", []byte(expectedConfig.Service.Address))
	client.PutConfigurationValue("Service/Port", []byte(strconv.Itoa(expectedConfig.Service.Port)))
	client.PutConfigurationValue("Port", []byte(strconv.Itoa(expectedConfig.Port)))
	client.PutConfigurationValue("Host", []byte(expectedConfig.Host))
	client.PutConfigurationValue("LogLevel", []byte(expectedConfig.LogLevel))

	updateChannel := make(chan interface{})
	errorChannel := make(chan error)

	client.WatchForChanges(updateChannel, errorChannel, &config.LoggingInfo{}, "Logging")

	pass := 1
	for {
		select {
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout waiting on configuration changes")

		case changes := <-updateChannel:
			assert.NotNil(t, changes)
			logInfo := changes.(*config.LoggingInfo)

			// first pass is for Consul Decoder always sending data once watch has been setup. It hasn't actually changed
			if pass == 1 {
				if !assert.Equal(t, logInfo.File, expectedConfig.Logging.File) {
					t.Fatal()
				}

				// Make a change to logging
				client.PutConfigurationValue("Logging/File", []byte(expectedChange))

				pass--
				continue
			}

			// Now the data should have changed
			assert.Equal(t, logInfo.File, expectedChange)
			return

		case waitError := <-errorChannel:
			t.Fatalf("received WatchForChanges error: %v", waitError)
		}
	}
}

func makeConsulClient(t *testing.T, servicePort int, setServiceInfo bool) *consulClient {
	registryInfo := config.RegistryInfo{
		Host: testHost,
		Port: port,
	}

	var serviceInfo *config.ServiceInfo = nil

	if setServiceInfo {
		serviceInfo = &config.ServiceInfo{
			CheckInterval: "1s",
			ClientMonitor: 1000,
			Host:          serviceHost,
			Port:          servicePort,
			Protocol:      "http",
		}
	}

	client, err := NewConsulClient(registryInfo, serviceInfo, serviceName)
	if assert.NoError(t, err) == false {
		t.Fatal()
	}

	return client
}

func createKeyValueMap() map[string]interface{} {
	configMap := make(map[string]interface{})

	configMap["int"] = int(1)
	configMap["int64"] = int64(64)
	configMap["float64"] = float64(1.4)
	configMap["string"] = "hello"
	configMap["bool"] = true

	return configMap
}

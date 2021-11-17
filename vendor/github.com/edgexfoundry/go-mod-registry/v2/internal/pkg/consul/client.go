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
	"net/http"
	"strings"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/edgexfoundry/go-mod-registry/v2/pkg/types"
)

const (
	consulStatusPath  = "/v1/status/leader"
	serviceStatusPass = "passing"
	aclError          = "Unexpected response code: 403"
)

type consulClient struct {
	config              *types.Config
	consulUrl           string
	consulClient        *consulapi.Client
	consulConfig        *consulapi.Config
	serviceKey          string
	serviceAddress      string
	servicePort         int
	healthCheckRoute    string
	healthCheckInterval string
	registeredChecks    []string
	getAccessToken      types.GetAccessTokenCallback
}

// Create new Consul Client. Service details are optional, not needed just for configuration, but required if registering
func NewConsulClient(registryConfig types.Config) (*consulClient, error) {

	client := consulClient{
		config:         &registryConfig,
		serviceKey:     registryConfig.ServiceKey,
		consulUrl:      registryConfig.GetRegistryUrl(),
		getAccessToken: registryConfig.GetAccessToken,
	}

	// ServiceHost will be empty when client isn't registering the service
	if registryConfig.ServiceHost != "" {
		client.servicePort = registryConfig.ServicePort
		client.serviceAddress = registryConfig.ServiceHost
		client.healthCheckRoute = registryConfig.CheckRoute
		client.healthCheckInterval = registryConfig.CheckInterval
	}

	var err error

	client.consulConfig = consulapi.DefaultConfig()
	client.consulConfig.Token = registryConfig.AccessToken
	client.consulConfig.Address = client.consulUrl
	client.consulClient, err = consulapi.NewClient(client.consulConfig)
	if err != nil {
		return nil, fmt.Errorf("unable for create new Consul Client for %s: %v", client.consulUrl, err)
	}

	return &client, nil
}

// Simply checks if Consul is up and running at the configured URL
func (client *consulClient) IsAlive() bool {
	netClient := http.Client{Timeout: time.Second * 10}

	// This REST endpoint doesn't require Access Token, so no need to handle Auth Error.
	resp, err := netClient.Get(client.consulUrl + consulStatusPath)
	if err != nil {
		return false
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return true
	}

	return false
}

// Registers the current service with Consul for discover and health check
func (client *consulClient) Register() error {
	if client.serviceKey == "" || client.serviceAddress == "" || client.servicePort == 0 ||
		client.healthCheckRoute == "" || client.healthCheckInterval == "" {
		return fmt.Errorf("unable to register service with consul: Service information not set")
	}

	registration := &consulapi.AgentServiceRegistration{
		Name:    client.serviceKey,
		Address: client.serviceAddress,
		Port:    client.servicePort,
	}

	// Register for service discovery
	err := client.consulClient.Agent().ServiceRegister(registration)

	retry, err := client.reloadAccessTokenOnAuthError(err)
	if retry {
		// Try again with new Access Token
		err = client.consulClient.Agent().ServiceRegister(registration)
	}

	if err != nil {
		return err
	}

	// Register for Health Check
	name := "Health Check: " + client.serviceKey
	notes := "Check the health of the API"
	err = client.RegisterCheck(client.serviceKey, name, notes, client.healthCheckRoute, client.healthCheckInterval)

	if err != nil {
		return err
	}

	return nil
}

// Register check with consul
func (client *consulClient) RegisterCheck(id string, name string, notes string, route string, interval string) error {
	registration := &consulapi.AgentCheckRegistration{
		ID:        id,
		Name:      name,
		Notes:     notes,
		ServiceID: client.serviceKey,
		AgentServiceCheck: consulapi.AgentServiceCheck{
			HTTP:     client.config.GetExpandedRoute(route),
			Interval: interval,
		},
	}

	err := client.consulClient.Agent().CheckRegister(registration)

	retry, err := client.reloadAccessTokenOnAuthError(err)
	if retry {
		// Try again with new Access Token
		err = client.consulClient.Agent().CheckRegister(registration)
	}

	if err != nil {
		client.registeredChecks = append(client.registeredChecks, id)
	}

	return err
}

func (client *consulClient) UnregisterCheck(checkId string) error {
	err := client.consulClient.Agent().CheckDeregister(checkId)

	retry, err := client.reloadAccessTokenOnAuthError(err)
	if retry {
		// Try again with new Access Token
		err = client.consulClient.Agent().CheckDeregister(checkId)
	}

	if err != nil {
		return fmt.Errorf("unable to de-register service health check with consul: %v", err)
	}
	return nil
}

func (client *consulClient) Unregister() error {
	err := client.consulClient.Agent().ServiceDeregister(client.serviceKey)

	retry, err := client.reloadAccessTokenOnAuthError(err)
	if retry {
		// Try again with new Access Token
		err = client.consulClient.Agent().ServiceDeregister(client.serviceKey)
	}

	if err != nil {
		return fmt.Errorf("unable to de-register service with consul: %v", err)
	}

	for _, checkId := range client.registeredChecks {
		if err := client.UnregisterCheck(checkId); err != nil {
			return err
		}
	}

	return nil
}

// GetServiceEndpoint retrieves the port, service ID and host of a known endpoint from Consul.
// If this operation is successful and a known endpoint is found, it is returned. Otherwise, an error is returned.
func (client *consulClient) GetServiceEndpoint(serviceID string) (types.ServiceEndpoint, error) {
	services, err := client.consulClient.Agent().Services()

	retry, err := client.reloadAccessTokenOnAuthError(err)
	if retry {
		// Try again with new Access Token
		services, err = client.consulClient.Agent().Services()
	}

	if err != nil {
		return types.ServiceEndpoint{}, err
	}

	endpoint := types.ServiceEndpoint{}
	if service, ok := services[serviceID]; ok {
		endpoint.Port = service.Port
		endpoint.ServiceId = serviceID
		endpoint.Host = service.Address
	} else {
		return types.ServiceEndpoint{}, fmt.Errorf("no matching service endpoint found")
	}

	return endpoint, nil
}

//GetAllServiceEndpoints retrieves all registered endpoints from Consul.
func (client *consulClient) GetAllServiceEndpoints() ([]types.ServiceEndpoint, error) {
	services, err := client.consulClient.Agent().Services()

	retry, err := client.reloadAccessTokenOnAuthError(err)
	if retry {
		// Try again with new Access Token
		services, err = client.consulClient.Agent().Services()
	}

	if err != nil {
		return nil, err
	}

	var endpoints []types.ServiceEndpoint
	for _, service := range services {
		svc := types.ServiceEndpoint{}
		svc.Port = service.Port
		svc.ServiceId = service.ID
		svc.Host = service.Address
		endpoints = append(endpoints, svc)
	}

	return endpoints, nil
}

// Checks with Consul if the target service is registered and healthy
func (client *consulClient) IsServiceAvailable(serviceKey string) (bool, error) {
	services, err := client.consulClient.Agent().Services()

	retry, err := client.reloadAccessTokenOnAuthError(err)
	if retry {
		// Try again with new Access Token
		services, err = client.consulClient.Agent().Services()
	}

	if err != nil {
		return false, fmt.Errorf("unable to check if service %s is available: %v", serviceKey, err)
	}

	if _, ok := services[serviceKey]; !ok {
		return false, fmt.Errorf("%s service is not registered. Might not have started... ", serviceKey)
	}

	healthCheck, _, err := client.consulClient.Health().Checks(serviceKey, nil)
	if err != nil {
		return false, fmt.Errorf("unable to check health of service %s: %v", serviceKey, err)
	}

	if len(healthCheck) == 0 {
		return false, fmt.Errorf("no health checks for service %s: %v", serviceKey, err)
	}

	status := healthCheck.AggregatedStatus()
	if status != serviceStatusPass {
		return false, fmt.Errorf(" %s service not healthy...", serviceKey)
	}

	return true, nil
}

func (client *consulClient) reloadAccessTokenOnAuthError(err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	if strings.Contains(err.Error(), aclError) && client.getAccessToken != nil {
		newToken, err := client.getAccessToken()
		if err != nil {
			err = fmt.Errorf("failed to renew access token: %s", err.Error())
			return false, err
		}

		client.consulConfig.Token = newToken

		// Have to recreate the consul client with the new Access Token
		client.consulClient, err = consulapi.NewClient(client.consulConfig)
		if err != nil {
			return false, fmt.Errorf("unable for create new Consul Client for %s: %v", client.consulUrl, err)
		}

		return true, nil
	}

	return false, err
}

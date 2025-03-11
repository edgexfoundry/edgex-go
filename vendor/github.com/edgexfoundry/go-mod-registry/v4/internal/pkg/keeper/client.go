//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	httpClient "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/go-mod-registry/v4/pkg/types"
)

type keeperClient struct {
	config              *types.Config
	keeperUrl           string
	serviceKey          string
	serviceHost         string
	servicePort         int
	healthCheckRoute    string
	healthCheckInterval string

	commonClient   interfaces.CommonClient
	registryClient interfaces.RegistryClient
}

// NewKeeperClient creates new Keeper Client. Service details are optional, not needed just for configuration, but required if registering
func NewKeeperClient(registryConfig types.Config) (*keeperClient, error) {
	client := keeperClient{
		config:     &registryConfig,
		serviceKey: registryConfig.ServiceKey,
		keeperUrl:  registryConfig.GetRegistryUrl(),
	}

	// ServiceHost will be empty when client isn't registering the service
	if registryConfig.ServiceHost != "" {
		client.servicePort = registryConfig.ServicePort
		client.serviceHost = registryConfig.ServiceHost
		client.healthCheckRoute = registryConfig.CheckRoute
		client.healthCheckInterval = registryConfig.CheckInterval
	}

	// Create the common and registry http clients for invoking APIs from Keeper
	client.commonClient = httpClient.NewCommonClient(client.keeperUrl, registryConfig.AuthInjector)
	client.registryClient = httpClient.NewRegistryClient(client.keeperUrl, registryConfig.AuthInjector, registryConfig.EnableNameFieldEscape)

	return &client, nil
}

// IsAlive simply checks if Keeper is up and running at the configured URL
func (k *keeperClient) IsAlive() bool {
	if _, err := k.commonClient.Ping(context.Background()); err != nil {
		return false
	}
	return true
}

// Register registers the current service with Keeper for discovery and health check
func (k *keeperClient) Register() error {
	if k.serviceKey == "" || k.serviceHost == "" || k.servicePort == 0 ||
		k.healthCheckRoute == "" || k.healthCheckInterval == "" {
		return fmt.Errorf("unable to register service with keeper: Service information not set")
	}

	registrationReq := requests.AddRegistrationRequest{
		BaseRequest: dtoCommon.BaseRequest{
			Versionable: dtoCommon.Versionable{ApiVersion: common.ApiVersion},
		},
		Registration: dtos.Registration{
			ServiceId: k.serviceKey,
			Host:      k.serviceHost,
			Port:      k.servicePort,
			HealthCheck: dtos.HealthCheck{
				Interval: k.healthCheckInterval,
				Path:     k.healthCheckRoute,
				Type:     "http",
			},
		},
	}

	// check if the service registry exists first
	resp, err := k.registryClient.RegistrationByServiceId(context.Background(), k.serviceKey)
	if err != nil && err.Code() != http.StatusNotFound {
		return fmt.Errorf("failed to check the %s service registry status: %v", k.serviceKey, err)
	}

	// call the UpdateRegister to update the registry if the service already exists
	// otherwise, call Register to create a new registry
	if resp.StatusCode == http.StatusOK {
		err := k.registryClient.UpdateRegister(context.Background(), registrationReq)
		if err != nil {
			return fmt.Errorf("failed to update the %s service registry: %v", k.serviceKey, err)
		}
	} else {
		err := k.registryClient.Register(context.Background(), registrationReq)
		if err != nil {
			return fmt.Errorf("failed to register the %s service: %v", k.serviceKey, err)
		}
	}

	return nil
}

// RegisterCheck registers a health check with Keeper
func (k *keeperClient) RegisterCheck(id string, name string, notes string, url string, interval string) error {
	// keeper combines service discovery and health check into one single register request
	return nil
}

func (k *keeperClient) UnregisterCheck(id string) error {
	// keeper combines service discovery and health check into one single register request
	return nil
}

// Unregister de-registers the current service from Keeper
func (k *keeperClient) Unregister() error {
	registrationReq := requests.AddRegistrationRequest{
		BaseRequest: dtoCommon.BaseRequest{
			Versionable: dtoCommon.Versionable{ApiVersion: common.ApiVersion},
		},
		Registration: dtos.Registration{
			ServiceId: k.serviceKey,
			Host:      k.serviceHost,
			Port:      k.servicePort,
			HealthCheck: dtos.HealthCheck{
				Interval: k.healthCheckInterval,
				Path:     k.healthCheckRoute,
				Type:     "http",
			},
			Status: models.Halt,
		},
	}

	err := k.registryClient.UpdateRegister(context.Background(), registrationReq)
	if err != nil {
		return fmt.Errorf("failed to de-register %s: %v", k.serviceKey, err)
	}

	return nil
}

// GetServiceEndpoint retrieves the port, service ID and host of a known endpoint from Keeper.
// If this operation is successful and a known endpoint is found, it is returned. Otherwise, an error is returned.
func (k *keeperClient) GetServiceEndpoint(serviceKey string) (types.ServiceEndpoint, error) {
	resp, err := k.registryClient.RegistrationByServiceId(context.Background(), serviceKey)
	if err != nil {
		return types.ServiceEndpoint{}, fmt.Errorf("failed to get service %s endpoint: %v", serviceKey, err)
	}

	endpoint := types.ServiceEndpoint{
		ServiceId: serviceKey,
		Host:      resp.Registration.Host,
		Port:      resp.Registration.Port,
	}

	return endpoint, nil
}

// GetAllServiceEndpoints retrieves all registered endpoints from Keeper.
func (k *keeperClient) GetAllServiceEndpoints() ([]types.ServiceEndpoint, error) {
	// filter out registrations with status is HALT which have been deregistered
	resp, err := k.registryClient.AllRegistry(context.Background(), false)
	if err != nil {
		return nil, fmt.Errorf("failed to get all service endpoints: %v", err)
	}

	endpoints := make([]types.ServiceEndpoint, len(resp.Registrations))
	for idx, r := range resp.Registrations {
		endpoint := types.ServiceEndpoint{
			ServiceId: r.ServiceId,
			Host:      r.Host,
			Port:      r.Port,
		}
		endpoints[idx] = endpoint
	}

	return endpoints, nil
}

// IsServiceAvailable checks with Keeper if the target service is registered and healthy
func (k *keeperClient) IsServiceAvailable(serviceKey string) (bool, error) {
	resp, err := k.registryClient.RegistrationByServiceId(context.Background(), serviceKey)
	if err != nil && err.Code() != http.StatusNotFound {
		return false, fmt.Errorf("failed to get %s service registry: %v", serviceKey, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		if strings.EqualFold(resp.Registration.Status, models.Halt) {
			return false, fmt.Errorf(" %s service has been unregistered", serviceKey)
		}
		if !strings.EqualFold(resp.Registration.Status, "up") {
			return false, fmt.Errorf(" %s service not healthy...", serviceKey)
		}

		return true, nil
	case http.StatusNotFound:
		return false, fmt.Errorf("%s service is not registered. Might not have started... ", serviceKey)
	default:
		return false, fmt.Errorf("failed to check service availability: %s", resp.Message)
	}
}

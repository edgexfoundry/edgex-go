/*******************************************************************************
 * Copyright 2022 Intel Inc.
 * Copyright (C) 2023 Intel Corporation
 * Copyright (C) 2024-2025 IOTech Ltd
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

package handlers

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients"
	httpClients "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	clientsMessaging "github.com/edgexfoundry/go-mod-messaging/v4/clients"
	"github.com/edgexfoundry/go-mod-registry/v4/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/v4/registry"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/zerotrust"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

// ClientsBootstrap contains data to boostrap the configured clients
type ClientsBootstrap struct {
	registry registry.Client
}

// NewClientsBootstrap is a factory method that returns the initialized "ClientsBootstrap" receiver struct.
func NewClientsBootstrap() *ClientsBootstrap {
	return &ClientsBootstrap{}
}

// BootstrapHandler fulfills the BootstrapHandler contract.
// It creates instances of each of the EdgeX clients that are in the service's configuration and place them in the DIC.
// If the registry is enabled it will be used to get the URL for client otherwise it will use configuration for the url.
// This handler will fail if an unknown client is specified.
func (cb *ClientsBootstrap) BootstrapHandler(
	_ context.Context,
	_ *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	lc := container.LoggingClientFrom(dic.Get)
	cfg := container.ConfigurationFrom(dic.Get)
	cb.registry = container.RegistryFrom(dic.Get)

	if cfg.GetBootstrap().Clients != nil {
		for serviceKey, serviceInfo := range *cfg.GetBootstrap().Clients {
			var urlFunc clients.ClientBaseUrlFunc

			sp := container.SecretProviderExtFrom(dic.Get)
			jwtSecretProvider := secret.NewJWTSecretProvider(sp)
			if serviceInfo.SecurityOptions[config.SecurityModeKey] == zerotrust.ZeroTrustMode {
				sp.EnableZeroTrust()
			}
			if rt, transpErr := zerotrust.HttpTransportFromClient(sp, serviceInfo, lc); transpErr != nil {
				lc.Errorf("could not obtain an http client for use with zero trust provider: %v", transpErr)
				return false
			} else {
				sp.SetHttpTransport(rt) //only need to set the transport when using SecretProviderExt
				sp.SetFallbackDialer(&net.Dialer{})
			}

			if !serviceInfo.UseMessageBus {
				mode := container.DevRemoteModeFrom(dic.Get)
				if cb.registry == nil || mode.InDevMode || mode.InRemoteMode {
					lc.Infof("Using REST for '%s' clients @ %s", serviceKey, serviceInfo.Url())
					urlFunc = clients.GetDefaultClientBaseUrlFunc(serviceInfo.Url())
				} else {
					lc.Infof("Using ClientBaseUrlFunc for '%s' clients", serviceKey)
					urlFunc = cb.clientUrlFunc(serviceKey, lc)
				}
			}

			switch serviceKey {
			case common.CoreDataServiceKey:
				dic.Update(di.ServiceConstructorMap{
					container.EventClientName: func(get di.Get) interface{} {
						return httpClients.NewEventClientWithUrlCallback(urlFunc, jwtSecretProvider, cfg.GetBootstrap().Service.EnableNameFieldEscape)
					},
					container.ReadingClientName: func(get di.Get) interface{} {
						return httpClients.NewReadingClientWithUrlCallback(urlFunc, jwtSecretProvider, cfg.GetBootstrap().Service.EnableNameFieldEscape)
					},
				})
			case common.CoreMetaDataServiceKey:
				dic.Update(di.ServiceConstructorMap{
					container.DeviceClientName: func(get di.Get) interface{} {
						return httpClients.NewDeviceClientWithUrlCallback(urlFunc, jwtSecretProvider, cfg.GetBootstrap().Service.EnableNameFieldEscape)
					},
					container.DeviceServiceClientName: func(get di.Get) interface{} {
						return httpClients.NewDeviceServiceClientWithUrlCallback(urlFunc, jwtSecretProvider, cfg.GetBootstrap().Service.EnableNameFieldEscape)
					},
					container.DeviceProfileClientName: func(get di.Get) interface{} {
						return httpClients.NewDeviceProfileClientWithUrlCallback(urlFunc, jwtSecretProvider, cfg.GetBootstrap().Service.EnableNameFieldEscape)
					},
					container.ProvisionWatcherClientName: func(get di.Get) interface{} {
						return httpClients.NewProvisionWatcherClientWithUrlCallback(urlFunc, jwtSecretProvider, cfg.GetBootstrap().Service.EnableNameFieldEscape)
					},
				})

			case common.CoreCommandServiceKey:
				var client interfaces.CommandClient

				if serviceInfo.UseMessageBus {
					// TODO: Move following outside loop when multiple messaging based clients exist
					messageClient := container.MessagingClientFrom(dic.Get)
					if messageClient == nil {
						lc.Errorf("Unable to create Command client using MessageBus: %s", "MessageBus Client was not created")
						return false
					}

					if len(cfg.GetBootstrap().Service.RequestTimeout) == 0 {
						lc.Error("Service.RequestTimeout found empty in service's configuration, missing common config? Use -cp or -cc flags for common config")
						return false
					}

					// TODO: Move following outside loop when multiple messaging based clients exist
					timeout, err := time.ParseDuration(cfg.GetBootstrap().Service.RequestTimeout)
					if err != nil {
						lc.Errorf("Unable to parse Service.RequestTimeout as a time duration: %v", err)
						return false
					}

					baseTopic := cfg.GetBootstrap().MessageBus.GetBaseTopicPrefix()
					if cfg.GetBootstrap().Service.EnableNameFieldEscape {
						client = clientsMessaging.NewCommandClientWithNameFieldEscape(messageClient, baseTopic, timeout)
					} else {
						client = clientsMessaging.NewCommandClient(messageClient, baseTopic, timeout)
					}

					lc.Infof("Using messaging for '%s' clients", serviceKey)
				} else {
					client = httpClients.NewCommandClientWithUrlCallback(urlFunc, jwtSecretProvider, cfg.GetBootstrap().Service.EnableNameFieldEscape)
				}

				dic.Update(di.ServiceConstructorMap{
					container.CommandClientName: func(get di.Get) interface{} {
						return client
					},
				})

			case common.SupportNotificationsServiceKey:
				dic.Update(di.ServiceConstructorMap{
					container.NotificationClientName: func(get di.Get) interface{} {
						return httpClients.NewNotificationClientWithUrlCallback(urlFunc, jwtSecretProvider, cfg.GetBootstrap().Service.EnableNameFieldEscape)
					},
					container.SubscriptionClientName: func(get di.Get) interface{} {
						return httpClients.NewSubscriptionClientWithUrlCallback(urlFunc, jwtSecretProvider, cfg.GetBootstrap().Service.EnableNameFieldEscape)
					},
				})

			case common.SupportSchedulerServiceKey:
				dic.Update(di.ServiceConstructorMap{
					container.ScheduleJobClientName: func(get di.Get) interface{} {
						return httpClients.NewScheduleJobClientWithUrlCallback(urlFunc, jwtSecretProvider, cfg.GetBootstrap().Service.EnableNameFieldEscape)
					},
					container.ScheduleActionRecordClientName: func(get di.Get) interface{} {
						return httpClients.NewScheduleActionRecordClientWithUrlCallback(urlFunc, jwtSecretProvider, cfg.GetBootstrap().Service.EnableNameFieldEscape)
					},
				})

			case common.SecurityProxyAuthServiceKey:
				dic.Update(di.ServiceConstructorMap{
					container.SecurityProxyAuthClientName: func(get di.Get) interface{} {
						return httpClients.NewAuthClientWithUrlCallback(urlFunc, jwtSecretProvider)
					},
				})

			default:

			}
		}
	}
	return true
}

func (cb *ClientsBootstrap) clientUrlFunc(serviceKey string, lc logger.LoggingClient) clients.ClientBaseUrlFunc {
	return func() (string, error) {
		var err error
		var endpoint types.ServiceEndpoint

		endpoint, err = cb.registry.GetServiceEndpoint(serviceKey)
		if err != nil {
			return "", fmt.Errorf("unable to Get service endpoint for '%s': %s", serviceKey, err.Error())
		}

		url := fmt.Sprintf("http://%s:%v", endpoint.Host, endpoint.Port)

		lc.Infof("Using registry for URL for '%s': %s", serviceKey, url)

		return url, nil
	}
}

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
package command

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/edgex-go/core/clients/metadata"
	"github.com/edgexfoundry/edgex-go/core/clients/types"
	"github.com/edgexfoundry/edgex-go/internal"
	consulclient "github.com/edgexfoundry/edgex-go/support/consul-client"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
)

var loggingClient logger.LoggingClient
var mdc metadata.DeviceClient
var cc metadata.CommandClient

func ConnectToConsul(conf ConfigurationStruct) error {

	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    internal.CoreCommandServiceKey,
		ServicePort:    conf.ServicePort,
		ServiceAddress: conf.ServiceAddress,
		CheckAddress:   conf.ConsulCheckAddress,
		CheckInterval:  conf.CheckInterval,
		ConsulAddress:  conf.ConsulHost,
		ConsulPort:     conf.ConsulPort,
	})

	if err != nil {
		return fmt.Errorf("connection to Consul could not be made: %v", err.Error())
	} else {
		// Update configuration data from Consul
		if err := consulclient.CheckKeyValuePairs(&conf, internal.CoreCommandServiceKey, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
			return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

func Init(conf ConfigurationStruct, l logger.LoggingClient, useConsul bool) {
	loggingClient = l
	configuration = conf
	//TODO: The above two are set due to global scope throughout the package. How can this be eliminated / refactored?

	// Create metadata clients
	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        conf.MetaDevicePath,
		UseRegistry: useConsul,
		Url:         conf.MetaDeviceURL}

	mdc = metadata.NewDeviceClient(params, types.Endpoint{})
	params.Path = conf.MetaCommandPath
	params.Url = conf.MetaCommandURL
	cc = metadata.NewCommandClient(params, types.Endpoint{})
}

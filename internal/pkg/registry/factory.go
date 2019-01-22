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

package registry

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/registry/consul"
)

var Client RegistryClient

func NewRegistryClient(registryInfo config.RegistryInfo, serviceInfo *config.ServiceInfo, serviceKey string) (RegistryClient, error) {
	switch registryInfo.Type {
	case "consul":
		var err error
		Client, err = consul.NewConsulClient(registryInfo, serviceInfo, serviceKey)
		return Client, err
	default:
		return nil, fmt.Errorf("Unknown registry type '%s' requested", registryInfo.Type)
	}
}

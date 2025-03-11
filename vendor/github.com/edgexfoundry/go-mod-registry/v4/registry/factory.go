//
// Copyright (c) 2019 Intel Corporation
// Copyright (C) 2024 IOTech Ltd
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

	"github.com/edgexfoundry/go-mod-registry/v4/internal/pkg/keeper"
	"github.com/edgexfoundry/go-mod-registry/v4/pkg/types"
)

func NewRegistryClient(registryConfig types.Config) (Client, error) {

	if registryConfig.Host == "" || registryConfig.Port == 0 {
		return nil, fmt.Errorf("unable to create RegistryClient: registry host and/or port or serviceKey not set")
	}

	switch registryConfig.Type {
	case "keeper":
		registryClient, err := keeper.NewKeeperClient(registryConfig)
		return registryClient, err
	default:
		return nil, fmt.Errorf("unknown registry type '%s' requested", registryConfig.Type)
	}
}

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

package configuration

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-configuration/v2/internal/pkg/consul"
	"github.com/edgexfoundry/go-mod-configuration/v2/pkg/types"
)

func NewConfigurationClient(config types.ServiceConfig) (Client, error) {

	if config.Host == "" || config.Port == 0 {
		return nil, fmt.Errorf("unable to create Configuration Client: Configuration service host and/or port or serviceKey not set")
	}

	switch config.Type {
	case "consul":
		var err error
		client, err := consul.NewConsulClient(config)
		return client, err
	default:
		return nil, fmt.Errorf("unknown configuration client type '%s' requested", config.Type)
	}
}

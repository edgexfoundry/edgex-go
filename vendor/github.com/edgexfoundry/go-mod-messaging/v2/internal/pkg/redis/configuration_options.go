/********************************************************************************
 *  Copyright 2020 Dell Inc.
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

package redis

import (
	"github.com/edgexfoundry/go-mod-messaging/v2/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

// OptionalClientConfiguration contains additional configuration properties which can be provided via the
// MessageBus.Optional's field.
type OptionalClientConfiguration struct {
	Password string
}

// NewClientConfiguration creates a OptionalClientConfiguration based on the configuration properties provided.
func NewClientConfiguration(config types.MessageBusConfig) (OptionalClientConfiguration, error) {
	redisConfig := OptionalClientConfiguration{}
	err := pkg.Load(config.Optional, &redisConfig)
	if err != nil {
		return OptionalClientConfiguration{}, err
	}

	return redisConfig, nil
}

/*******************************************************************************
 * Copyright 2022 Intel Corporation
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
 *
 *******************************************************************************/

package bootstrapper

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/mosquitto"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/flags"
)

const (
	Mosquitto = "mosquitto"
	Redis     = "redis"
)

func ConfigureSecureMessageBus(brokerType string, ctx context.Context, cancel context.CancelFunc, f flags.Common) error {
	switch brokerType {
	case Mosquitto:
		mosquitto.Configure(ctx, cancel, f)
		return nil
	case Redis:
		//no op as Redis message bus is handled in configureRedis
		return nil
	default:
		return fmt.Errorf("Broker Type Not Supported: %s", brokerType)
	}
}

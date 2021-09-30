/*******************************************************************************
 * Copyright 2021 Intel Corp.
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

package messaging

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapMessaging "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

// BootstrapHandler fulfills the BootstrapHandler contract.  if enabled, tt creates and initializes the Messaging client
// and adds it to the DIC
func BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageQueue

	messageBusInfo.AuthMode = strings.ToLower(strings.TrimSpace(messageBusInfo.AuthMode))
	if len(messageBusInfo.AuthMode) > 0 && messageBusInfo.AuthMode != bootstrapMessaging.AuthModeNone {
		if err := bootstrapMessaging.SetOptionsAuthData(&messageBusInfo, lc, dic); err != nil {
			lc.Error(err.Error())
			return false
		}
	}

	msgClient, err := messaging.NewMessageClient(
		types.MessageBusConfig{
			PublishHost: types.HostInfo{
				Host:     messageBusInfo.Host,
				Port:     messageBusInfo.Port,
				Protocol: messageBusInfo.Protocol,
			},
			SubscribeHost: types.HostInfo{
				Host:     messageBusInfo.Host,
				Port:     messageBusInfo.Port,
				Protocol: messageBusInfo.Protocol,
			},
			Type:     messageBusInfo.Type,
			Optional: messageBusInfo.Optional,
		})

	if err != nil {
		lc.Errorf("Failed to create MessageClient: %v", err)
		return false
	}

	for startupTimer.HasNotElapsed() {
		select {
		case <-ctx.Done():
			return false
		default:
			err = msgClient.Connect()
			if err != nil {
				lc.Warnf("Unable to connect MessageBus: %w", err)
				startupTimer.SleepForInterval()
				continue
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ctx.Done()
				if msgClient != nil {
					_ = msgClient.Disconnect()
				}
				lc.Infof("Disconnected from MessageBus")
			}()

			dic.Update(di.ServiceConstructorMap{
				container.MessagingClientName: func(get di.Get) interface{} {
					return msgClient
				},
			})

			lc.Info(fmt.Sprintf(
				"Connected to %s Message Bus @ %s://%s:%d publishing on '%s' prefix topic with AuthMode='%s'",
				messageBusInfo.Type,
				messageBusInfo.Protocol,
				messageBusInfo.Host,
				messageBusInfo.Port,
				messageBusInfo.PublishTopicPrefix,
				messageBusInfo.AuthMode))

			return true
		}
	}

	lc.Error("Connecting to MessageBus time out")
	return false
}

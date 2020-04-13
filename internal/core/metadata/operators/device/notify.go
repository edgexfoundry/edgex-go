/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package device

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	metaConfig "github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

type Notifier interface {
	// Execute performs the main logic of the Notifier implementation.
	// This function is meant to be called from a goroutine, thus log eny errors and do not return them.
	Execute()
}

type deviceNotifier struct {
	ctx          context.Context
	database     DeviceServiceLoader
	events       chan DeviceEvent
	logger       logger.LoggingClient
	notifyClient notifications.NotificationsClient
	notifyConfig metaConfig.NotificationInfo
	requester    Requester
}

func NewNotifier(evt chan DeviceEvent, nc notifications.NotificationsClient, cfg metaConfig.NotificationInfo,
	db DeviceServiceLoader, requester Requester, logger logger.LoggingClient, ctx context.Context) Notifier {
	return deviceNotifier{
		ctx:          ctx,
		database:     db,
		events:       evt,
		logger:       logger,
		notifyClient: nc,
		notifyConfig: cfg,
		requester:    requester,
	}
}

// Remember that this method is being invoked via a goroutine. The following logic is all async to the caller.
func (op deviceNotifier) Execute() {
	select {
	case msg := <-op.events:
		if msg.Error != nil {
			op.logger.Error(fmt.Sprintf("dropping event due to error: %s", msg.Error.Error()))
			return // Something happened during the upstream operation. Do nothing.
		}
		deviceId := msg.DeviceId
		deviceName := msg.DeviceName
		httpMethod := msg.HttpMethod

		// Perform logic previously found in rest_device.go:: func notifyDeviceAssociates()
		op.postNotification(deviceName, httpMethod)

		// Callback for device service
		service, err := op.database.GetDeviceServiceById(msg.ServiceId)
		if err != nil {
			op.logger.Error(err.Error())
			return
		}

		if err := op.callback(service, deviceId, httpMethod, models.DEVICE); err != nil {
			op.logger.Error(err.Error())
		}
	}
}

func (op deviceNotifier) postNotification(name string, action string) {
	// Only post notification if the configuration is set
	if op.notifyConfig.PostDeviceChanges {
		// Make the notification
		notification := notifications.Notification{
			Slug:        op.notifyConfig.Slug + strconv.FormatInt(db.MakeTimestamp(), 10),
			Content:     op.notifyConfig.Content + name + "-" + action,
			Category:    notifications.SW_HEALTH,
			Description: op.notifyConfig.Description,
			Labels:      []string{op.notifyConfig.Label},
			Sender:      op.notifyConfig.Sender,
			Severity:    notifications.NORMAL,
		}

		_ = op.notifyClient.SendNotification(op.ctx, notification)
	}
}

// Make the callback for the device service
func (op deviceNotifier) callback(
	service models.DeviceService,
	id string,
	action string,
	actionType models.ActionType) error {

	url := service.Addressable.GetCallbackURL()
	if len(url) > 0 {
		body, err := json.Marshal(models.CallbackAlert{ActionType: actionType, Id: id})
		if err != nil {
			return err
		}
		req, err := http.NewRequest(action, url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Add(clients.ContentType, clients.ContentTypeJSON)
		go op.requester.Execute(req)
	} else {
		op.logger.Error("callback::no addressable for " + service.Name)
	}
	return nil
}

//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/core/command/application"
)

// validateRequestTopic validates the request topic by checking the existence of device and device service,
// returns the internal device request topic to which the command request will be sent.
func validateRequestTopic(prefix string, deviceName string, commandName string, method string, dic *di.Container) (string, error) {
	// retrieve device information through Metadata DeviceClient
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	if dc == nil {
		return "", errors.New("nil Device Client")
	}
	deviceResponse, err := dc.DeviceByName(context.Background(), deviceName)
	if err != nil {
		return "", fmt.Errorf("failed to get Device by name %s: %v", deviceName, err)
	}

	// retrieve device service information through Metadata DeviceClient
	dsc := bootstrapContainer.DeviceServiceClientFrom(dic.Get)
	if dsc == nil {
		return "", errors.New("nil DeviceService Client")
	}
	deviceServiceResponse, err := dsc.DeviceServiceByName(context.Background(), deviceResponse.Device.ServiceName)
	if err != nil {
		return "", fmt.Errorf("failed to get DeviceService by name %s: %v", deviceResponse.Device.ServiceName, err)
	}

	// expected internal command request topic scheme: #/<device-service>/<device>/<command-name>/<method>
	return strings.Join([]string{prefix, deviceServiceResponse.Service.Name, deviceName, commandName, method}, "/"), nil

}

// getCommandQueryResponseEnvelope returns the MessageEnvelope containing the DeviceCoreCommand payload bytes
func getCommandQueryResponseEnvelope(requestEnvelope types.MessageEnvelope, deviceName string, dic *di.Container) (types.MessageEnvelope, edgexErr.EdgeX) {
	var err error
	var commands any
	var edgexError edgexErr.EdgeX

	switch deviceName {
	case common.All:
		offset, limit := common.DefaultOffset, common.DefaultLimit
		if requestEnvelope.QueryParams != nil {
			if offsetRaw, ok := requestEnvelope.QueryParams[common.Offset]; ok {
				offset, err = strconv.Atoi(offsetRaw)
				if err != nil {
					return types.MessageEnvelope{}, edgexErr.NewCommonEdgeX(edgexErr.KindContractInvalid, fmt.Sprintf("Failed to convert 'offset' query parameter to intger: %s", err.Error()), err)
				}
			}
			if limitRaw, ok := requestEnvelope.QueryParams[common.Limit]; ok {
				limit, err = strconv.Atoi(limitRaw)
				if err != nil {
					return types.MessageEnvelope{}, edgexErr.NewCommonEdgeX(edgexErr.KindContractInvalid, fmt.Sprintf("Failed to convert 'limit' query parameter to integer: %s", err.Error()), err)
				}
			}
		}

		commands, _, edgexError = application.AllCommands(offset, limit, dic)
		if edgexError != nil {
			return types.MessageEnvelope{}, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, fmt.Sprintf("Failed to get all commands: %s", edgexError.Error()), edgexError)
		}
	default:
		commands, edgexError = application.CommandsByDeviceName(deviceName, dic)
		if edgexError != nil {
			return types.MessageEnvelope{}, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, fmt.Sprintf("Failed to get commands by device name '%s': %s", deviceName, edgexError.Error()), edgexError)
		}
	}

	responseBytes, err := json.Marshal(commands)
	if err != nil {
		return types.MessageEnvelope{}, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, fmt.Sprintf("Failed to json encoding device commands payload: %s", err.Error()), err)
	}

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, requestEnvelope.RequestID, requestEnvelope.CorrelationID, common.ContentTypeJSON)
	if err != nil {
		return types.MessageEnvelope{}, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, fmt.Sprintf("Failed to create response MessageEnvelope: %s", err.Error()), err)
	}

	return responseEnvelope, nil
}

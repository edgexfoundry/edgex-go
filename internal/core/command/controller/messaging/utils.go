//
// Copyright (C) 2022-2025 IOTech Ltd
// Copyright (C) 2023 Intel Inc.
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/core/command/application"
)

// retrieveServiceNameByDevice validates the existence of device and device service,
// returns the service name to which the command request will be sent.
func retrieveServiceNameByDevice(deviceName string, dic *di.Container) (string, error) {
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
	return deviceServiceResponse.Service.Name, nil
}

// validateGetCommandQueryParameters validates the value is valid for device service's reserved query parameters
func validateGetCommandQueryParameters(queryParams map[string]string) error {
	if dsReturnEvent, ok := queryParams[common.ReturnEvent]; ok {
		if dsReturnEvent != common.ValueTrue && dsReturnEvent != common.ValueFalse {
			return fmt.Errorf("invalid query parameter, %s has to be '%s' or '%s'", common.ReturnEvent, common.ValueTrue, common.ValueFalse)
		}
	}
	if dsPushEvent, ok := queryParams[common.PushEvent]; ok {
		if dsPushEvent != common.ValueTrue && dsPushEvent != common.ValueFalse {
			return fmt.Errorf("invalid query parameter, %s has to be '%s' or '%s'", common.PushEvent, common.ValueTrue, common.ValueFalse)
		}
	}

	return nil
}

// getCommandQueryResponseEnvelope returns the MessageEnvelope containing the DeviceCoreCommand payload bytes
func getCommandQueryResponseEnvelope(requestEnvelope types.MessageEnvelope, deviceName string, dic *di.Container) (types.MessageEnvelope, error) {
	var commandsResponse any
	var err error

	switch deviceName {
	case common.All:
		offset, limit := common.DefaultOffset, common.DefaultLimit
		if requestEnvelope.QueryParams != nil {
			if offsetRaw, ok := requestEnvelope.QueryParams[common.Offset]; ok {
				offset, err = strconv.Atoi(offsetRaw)
				if err != nil {
					return types.MessageEnvelope{}, fmt.Errorf("failed to convert 'offset' query parameter to intger: %s", err.Error())
				}
			}
			if limitRaw, ok := requestEnvelope.QueryParams[common.Limit]; ok {
				limit, err = strconv.Atoi(limitRaw)
				if err != nil {
					return types.MessageEnvelope{}, fmt.Errorf("failed to convert 'limit' query parameter to integer: %s", err.Error())
				}
			}
		}

		commands, totalCounts, edgexError := application.AllCommands(offset, limit, dic)
		if edgexError != nil {
			return types.MessageEnvelope{}, fmt.Errorf("failed to get all commands: %s", edgexError.Error())
		}

		commandsResponse = responses.NewMultiDeviceCoreCommandsResponse(requestEnvelope.RequestID, "", http.StatusOK, totalCounts, commands)
	default:
		commands, edgexError := application.CommandsByDeviceName(deviceName, dic)
		if err != nil {
			return types.MessageEnvelope{}, fmt.Errorf("failed to get commands by device name '%s': %s", deviceName, edgexError.Error())
		}

		commandsResponse = responses.NewDeviceCoreCommandResponse(requestEnvelope.RequestID, "", http.StatusOK, commands)
	}

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(commandsResponse, requestEnvelope.RequestID, requestEnvelope.CorrelationID, common.ContentTypeJSON)
	if err != nil {
		return types.MessageEnvelope{}, fmt.Errorf("failed to create response MessageEnvelope: %s", err.Error())
	}

	return responseEnvelope, nil
}

//
// Copyright (C) 2022-2025 IOTech Ltd
// Copyright (c) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/edgexfoundry/go-mod-messaging/v4/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

type CommandClient struct {
	messageBus            messaging.MessageClient
	baseTopic             string
	responseTopicPrefix   string
	timeout               time.Duration
	enableNameFieldEscape bool
}

// NewCommandClient returns the command client with the disabled NameFieldEscape
func NewCommandClient(messageBus messaging.MessageClient, baseTopic string, timeout time.Duration) interfaces.CommandClient {
	client := &CommandClient{
		messageBus:          messageBus,
		baseTopic:           baseTopic,
		responseTopicPrefix: common.BuildTopic(baseTopic, common.ResponseTopic, common.CoreCommandServiceKey),
		timeout:             timeout,
	}

	return client
}

// NewCommandClientWithNameFieldEscape returns the command client with the enabled NameFieldEscape
func NewCommandClientWithNameFieldEscape(messageBus messaging.MessageClient, baseTopic string, timeout time.Duration) interfaces.CommandClient {
	client := &CommandClient{
		messageBus:            messageBus,
		baseTopic:             baseTopic,
		responseTopicPrefix:   common.BuildTopic(baseTopic, common.ResponseTopic, common.CoreCommandServiceKey),
		timeout:               timeout,
		enableNameFieldEscape: true,
	}

	return client
}

func (c *CommandClient) AllDeviceCoreCommands(_ context.Context, offset int, limit int) (responses.MultiDeviceCoreCommandsResponse, edgexErr.EdgeX) {
	queryParams := map[string]string{common.Offset: strconv.Itoa(offset), common.Limit: strconv.Itoa(limit)}
	requestEnvelope := types.NewMessageEnvelopeForRequest(nil, queryParams)

	requestTopic := common.BuildTopic(c.baseTopic, common.CoreCommandQueryRequestPublishTopic, common.All)
	responseEnvelope, err := c.messageBus.Request(requestEnvelope, common.CoreCommandServiceKey, requestTopic, c.timeout)
	if err != nil {
		return responses.MultiDeviceCoreCommandsResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	if responseEnvelope.ErrorCode == 1 {
		return responses.MultiDeviceCoreCommandsResponse{}, edgexErr.NewCommonEdgeXWrapper(fmt.Errorf("%v", responseEnvelope.Payload))
	}

	var res responses.MultiDeviceCoreCommandsResponse
	res, err = types.GetMsgPayload[responses.MultiDeviceCoreCommandsResponse](*responseEnvelope)
	if err != nil {
		return responses.MultiDeviceCoreCommandsResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	return res, nil
}

func (c *CommandClient) DeviceCoreCommandsByDeviceName(_ context.Context, deviceName string) (responses.DeviceCoreCommandResponse, edgexErr.EdgeX) {
	requestEnvelope := types.NewMessageEnvelopeForRequest(nil, nil)
	requestTopic := common.NewPathBuilder().EnableNameFieldEscape(c.enableNameFieldEscape).
		SetPath(c.baseTopic).SetPath(common.CoreCommandQueryRequestPublishTopic).SetNameFieldPath(deviceName).BuildPath()
	responseEnvelope, err := c.messageBus.Request(requestEnvelope, requestTopic, c.responseTopicPrefix, c.timeout)
	if err != nil {
		return responses.DeviceCoreCommandResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	if responseEnvelope.ErrorCode == 1 {
		return responses.DeviceCoreCommandResponse{}, edgexErr.NewCommonEdgeXWrapper(fmt.Errorf("%v", responseEnvelope.Payload))
	}

	var res responses.DeviceCoreCommandResponse
	res, err = types.GetMsgPayload[responses.DeviceCoreCommandResponse](*responseEnvelope)
	if err != nil {
		return responses.DeviceCoreCommandResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	return res, nil
}

func (c *CommandClient) IssueGetCommandByName(ctx context.Context, deviceName string, commandName string, dsPushEvent bool, dsReturnEvent bool) (*responses.EventResponse, edgexErr.EdgeX) {
	queryParams := map[string]string{common.PushEvent: strconv.FormatBool(dsPushEvent), common.ReturnEvent: strconv.FormatBool(dsReturnEvent)}
	return c.IssueGetCommandByNameWithQueryParams(ctx, deviceName, commandName, queryParams)
}

func (c *CommandClient) IssueGetCommandByNameWithQueryParams(_ context.Context, deviceName string, commandName string, queryParams map[string]string) (*responses.EventResponse, edgexErr.EdgeX) {
	requestEnvelope := types.NewMessageEnvelopeForRequest(nil, queryParams)
	requestTopic := common.NewPathBuilder().EnableNameFieldEscape(c.enableNameFieldEscape).
		SetPath(c.baseTopic).SetPath(common.CoreCommandRequestPublishTopic).SetNameFieldPath(deviceName).SetNameFieldPath(commandName).SetPath("get").BuildPath()
	responseEnvelope, err := c.messageBus.Request(requestEnvelope, requestTopic, c.responseTopicPrefix, c.timeout)
	if err != nil {
		return nil, edgexErr.NewCommonEdgeXWrapper(err)
	}

	if responseEnvelope.ErrorCode == 1 {
		return nil, edgexErr.NewCommonEdgeXWrapper(fmt.Errorf("%v", responseEnvelope.Payload))
	}

	var res responses.EventResponse
	returnEvent, ok := queryParams[common.ReturnEvent]
	if ok && returnEvent == common.ValueFalse {
		res.ApiVersion = common.ApiVersion
		res.RequestId = responseEnvelope.RequestID
		res.StatusCode = http.StatusOK
	} else {
		res, err = types.GetMsgPayload[responses.EventResponse](*responseEnvelope)
		if err != nil {
			return nil, edgexErr.NewCommonEdgeXWrapper(err)
		}
	}

	return &res, nil
}

func (c *CommandClient) IssueSetCommandByName(_ context.Context, deviceName string, commandName string, settings map[string]any) (commonDTO.BaseResponse, edgexErr.EdgeX) {
	requestEnvelope := types.NewMessageEnvelopeForRequest(settings, nil)
	requestTopic := common.NewPathBuilder().EnableNameFieldEscape(c.enableNameFieldEscape).
		SetPath(c.baseTopic).SetPath(common.CoreCommandRequestPublishTopic).SetNameFieldPath(deviceName).SetNameFieldPath(commandName).SetPath("set").BuildPath()
	responseEnvelope, err := c.messageBus.Request(requestEnvelope, requestTopic, c.responseTopicPrefix, c.timeout)
	if err != nil {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	if responseEnvelope.ErrorCode == 1 {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(fmt.Errorf("%v", responseEnvelope.Payload))
	}

	res := commonDTO.NewBaseResponse(responseEnvelope.RequestID, "", http.StatusOK)
	return res, nil
}

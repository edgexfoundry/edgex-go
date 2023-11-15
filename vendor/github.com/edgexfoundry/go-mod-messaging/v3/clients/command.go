//
// Copyright (C) 2022 IOTech Ltd
// Copyright (c) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/edgexfoundry/go-mod-messaging/v3/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
)

type CommandClient struct {
	messageBus          messaging.MessageClient
	baseTopic           string
	responseTopicPrefix string
	timeout             time.Duration
}

func NewCommandClient(messageBus messaging.MessageClient, baseTopic string, timeout time.Duration) interfaces.CommandClient {
	client := &CommandClient{
		messageBus:          messageBus,
		baseTopic:           baseTopic,
		responseTopicPrefix: common.BuildTopic(baseTopic, common.ResponseTopic, common.CoreCommandServiceKey),
		timeout:             timeout,
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
		return responses.MultiDeviceCoreCommandsResponse{}, edgexErr.NewCommonEdgeXWrapper(errors.New(string(responseEnvelope.Payload)))
	}

	var res responses.MultiDeviceCoreCommandsResponse
	err = json.Unmarshal(responseEnvelope.Payload, &res)
	if err != nil {
		return responses.MultiDeviceCoreCommandsResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	return res, nil
}

func (c *CommandClient) DeviceCoreCommandsByDeviceName(_ context.Context, deviceName string) (responses.DeviceCoreCommandResponse, edgexErr.EdgeX) {
	requestEnvelope := types.NewMessageEnvelopeForRequest(nil, nil)
	requestTopic := common.BuildTopic(c.baseTopic, common.CoreCommandQueryRequestPublishTopic, deviceName)
	responseEnvelope, err := c.messageBus.Request(requestEnvelope, requestTopic, c.responseTopicPrefix, c.timeout)
	if err != nil {
		return responses.DeviceCoreCommandResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	if responseEnvelope.ErrorCode == 1 {
		return responses.DeviceCoreCommandResponse{}, edgexErr.NewCommonEdgeXWrapper(errors.New(string(responseEnvelope.Payload)))
	}

	var res responses.DeviceCoreCommandResponse
	err = json.Unmarshal(responseEnvelope.Payload, &res)
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
	requestTopic := common.BuildTopic(c.baseTopic, common.CoreCommandRequestPublishTopic, deviceName, commandName, "get")
	responseEnvelope, err := c.messageBus.Request(requestEnvelope, requestTopic, c.responseTopicPrefix, c.timeout)
	if err != nil {
		return nil, edgexErr.NewCommonEdgeXWrapper(err)
	}

	if responseEnvelope.ErrorCode == 1 {
		return nil, edgexErr.NewCommonEdgeXWrapper(errors.New(string(responseEnvelope.Payload)))
	}

	var res responses.EventResponse
	returnEvent, ok := queryParams[common.ReturnEvent]
	if ok && returnEvent == common.ValueFalse {
		res.ApiVersion = common.ApiVersion
		res.RequestId = responseEnvelope.RequestID
		res.StatusCode = http.StatusOK
	} else {
		err = json.Unmarshal(responseEnvelope.Payload, &res)
		if err != nil {
			return nil, edgexErr.NewCommonEdgeXWrapper(err)
		}
	}

	return &res, nil
}

func (c *CommandClient) IssueSetCommandByName(_ context.Context, deviceName string, commandName string, settings map[string]string) (commonDTO.BaseResponse, edgexErr.EdgeX) {
	payloadBytes, err := json.Marshal(settings)
	if err != nil {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	requestEnvelope := types.NewMessageEnvelopeForRequest(payloadBytes, nil)
	requestTopic := common.BuildTopic(c.baseTopic, common.CoreCommandRequestPublishTopic, deviceName, commandName, "set")
	responseEnvelope, err := c.messageBus.Request(requestEnvelope, requestTopic, c.responseTopicPrefix, c.timeout)
	if err != nil {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	if responseEnvelope.ErrorCode == 1 {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(errors.New(string(responseEnvelope.Payload)))
	}

	res := commonDTO.NewBaseResponse(responseEnvelope.RequestID, "", http.StatusOK)
	return res, nil
}

func (c *CommandClient) IssueSetCommandByNameWithObject(_ context.Context, deviceName string, commandName string, settings map[string]any) (commonDTO.BaseResponse, edgexErr.EdgeX) {
	payloadBytes, err := json.Marshal(settings)
	if err != nil {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	requestEnvelope := types.NewMessageEnvelopeForRequest(payloadBytes, nil)
	requestTopic := common.BuildTopic(c.baseTopic, common.CoreCommandRequestPublishTopic, deviceName, commandName, "set")
	responseEnvelope, err := c.messageBus.Request(requestEnvelope, requestTopic, c.responseTopicPrefix, c.timeout)
	if err != nil {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	if responseEnvelope.ErrorCode == 1 {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(errors.New(string(responseEnvelope.Payload)))
	}

	res := commonDTO.NewBaseResponse(responseEnvelope.RequestID, "", http.StatusOK)
	return res, nil
}

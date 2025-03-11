//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/url"
	"path"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type ScheduleActionRecordClient struct {
	baseUrlFunc           clients.ClientBaseUrlFunc
	authInjector          interfaces.AuthenticationInjector
	enableNameFieldEscape bool
}

// NewScheduleActionRecordClient creates an instance of ScheduleActionRecordClient
func NewScheduleActionRecordClient(baseUrl string, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.ScheduleActionRecordClient {
	return &ScheduleActionRecordClient{
		baseUrlFunc:           clients.GetDefaultClientBaseUrlFunc(baseUrl),
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

// NewScheduleActionRecordClientWithUrlCallback creates an instance of ScheduleActionRecordClient with ClientBaseUrlFunc.
func NewScheduleActionRecordClientWithUrlCallback(baseUrlFunc clients.ClientBaseUrlFunc, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.ScheduleActionRecordClient {
	return &ScheduleActionRecordClient{
		baseUrlFunc:           baseUrlFunc,
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

// AllScheduleActionRecords query schedule action records with start, end, offset, and limit
func (client *ScheduleActionRecordClient) AllScheduleActionRecords(ctx context.Context, start, end int64, offset, limit int) (res responses.MultiScheduleActionRecordsResponse, err errors.EdgeX) {
	requestParams := url.Values{}
	requestParams.Set(common.Start, strconv.FormatInt(start, 10))
	requestParams.Set(common.End, strconv.FormatInt(end, 10))
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	baseUrl, goErr := clients.GetBaseUrl(client.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, common.ApiAllScheduleActionRecordRoute, requestParams, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// LatestScheduleActionRecordsByJobName query the latest schedule action records by job name
func (client *ScheduleActionRecordClient) LatestScheduleActionRecordsByJobName(ctx context.Context, jobName string) (res responses.MultiScheduleActionRecordsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiScheduleActionRecordRoute, common.Latest, common.Job, common.Name, jobName)
	requestParams := url.Values{}
	requestParams.Set(common.Name, jobName)
	baseUrl, goErr := clients.GetBaseUrl(client.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// ScheduleActionRecordsByStatus queries schedule action records with status, start, end, offset, and limit
func (client *ScheduleActionRecordClient) ScheduleActionRecordsByStatus(ctx context.Context, status string, start, end int64, offset, limit int) (res responses.MultiScheduleActionRecordsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiScheduleActionRecordRoute, common.Status, status)
	requestParams := url.Values{}
	requestParams.Set(common.Start, strconv.FormatInt(start, 10))
	requestParams.Set(common.End, strconv.FormatInt(end, 10))
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	baseUrl, goErr := clients.GetBaseUrl(client.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// ScheduleActionRecordsByJobName queries schedule action records with jobName, start, end, offset, and limit
func (client *ScheduleActionRecordClient) ScheduleActionRecordsByJobName(ctx context.Context, jobName string, start, end int64, offset, limit int) (res responses.MultiScheduleActionRecordsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiScheduleActionRecordRoute, common.Job, common.Name, jobName)
	requestParams := url.Values{}
	requestParams.Set(common.Start, strconv.FormatInt(start, 10))
	requestParams.Set(common.End, strconv.FormatInt(end, 10))
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	baseUrl, goErr := clients.GetBaseUrl(client.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// ScheduleActionRecordsByJobNameAndStatus queries schedule action records with jobName, status, start, end, offset, and limit
func (client *ScheduleActionRecordClient) ScheduleActionRecordsByJobNameAndStatus(ctx context.Context, jobName, status string, start, end int64, offset, limit int) (res responses.MultiScheduleActionRecordsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiScheduleActionRecordRoute, common.Job, common.Name, jobName, common.Status, status)
	requestParams := url.Values{}
	requestParams.Set(common.Start, strconv.FormatInt(start, 10))
	requestParams.Set(common.End, strconv.FormatInt(end, 10))
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	baseUrl, goErr := clients.GetBaseUrl(client.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

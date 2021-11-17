//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/url"
	"path"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

type NotificationClient struct {
	baseUrl string
}

// NewNotificationClient creates an instance of NotificationClient
func NewNotificationClient(baseUrl string) interfaces.NotificationClient {
	return &NotificationClient{
		baseUrl: baseUrl,
	}
}

// SendNotification sends new notifications.
func (client *NotificationClient) SendNotification(ctx context.Context, reqs []requests.AddNotificationRequest) (res []dtoCommon.BaseWithIdResponse, err errors.EdgeX) {
	err = utils.PostRequestWithRawData(ctx, &res, client.baseUrl+common.ApiNotificationRoute, reqs)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// NotificationById query notification by id.
func (client *NotificationClient) NotificationById(ctx context.Context, id string) (res responses.NotificationResponse, err errors.EdgeX) {
	path := path.Join(common.ApiNotificationRoute, common.Id, url.QueryEscape(id))
	err = utils.GetRequest(ctx, &res, client.baseUrl, path, nil)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// DeleteNotificationById deletes a notification by id.
func (client *NotificationClient) DeleteNotificationById(ctx context.Context, id string) (res dtoCommon.BaseResponse, err errors.EdgeX) {
	path := path.Join(common.ApiNotificationRoute, common.Id, url.QueryEscape(id))
	err = utils.DeleteRequest(ctx, &res, client.baseUrl, path)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// NotificationsByCategory queries notifications with category, offset and limit
func (client *NotificationClient) NotificationsByCategory(ctx context.Context, category string, offset int, limit int) (res responses.MultiNotificationsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiNotificationRoute, common.Category, url.QueryEscape(category))
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, requestPath, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// NotificationsByLabel queries notifications with label, offset and limit
func (client *NotificationClient) NotificationsByLabel(ctx context.Context, label string, offset int, limit int) (res responses.MultiNotificationsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiNotificationRoute, common.Label, url.QueryEscape(label))
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, requestPath, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// NotificationsByStatus queries notifications with status, offset and limit
func (client *NotificationClient) NotificationsByStatus(ctx context.Context, status string, offset int, limit int) (res responses.MultiNotificationsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiNotificationRoute, common.Status, url.QueryEscape(status))
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, requestPath, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// NotificationsByTimeRange query notifications with time range, offset and limit
func (client *NotificationClient) NotificationsByTimeRange(ctx context.Context, start int, end int, offset int, limit int) (res responses.MultiNotificationsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiNotificationRoute, common.Start, strconv.Itoa(start), common.End, strconv.Itoa(end))
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, requestPath, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// NotificationsBySubscriptionName query notifications with subscriptionName, offset and limit
func (client *NotificationClient) NotificationsBySubscriptionName(ctx context.Context, subscriptionName string, offset int, limit int) (res responses.MultiNotificationsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiNotificationRoute, common.Subscription, common.Name, url.QueryEscape(subscriptionName))
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, requestPath, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// CleanupNotificationsByAge removes notifications that are older than age. And the corresponding transmissions will also be deleted.
// Age is supposed in milliseconds since modified timestamp
func (client *NotificationClient) CleanupNotificationsByAge(ctx context.Context, age int) (res dtoCommon.BaseResponse, err errors.EdgeX) {
	path := path.Join(common.ApiNotificationCleanupRoute, common.Age, strconv.Itoa(age))
	err = utils.DeleteRequest(ctx, &res, client.baseUrl, path)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// CleanupNotifications removes notifications and the corresponding transmissions.
func (client *NotificationClient) CleanupNotifications(ctx context.Context) (res dtoCommon.BaseResponse, err errors.EdgeX) {
	err = utils.DeleteRequest(ctx, &res, client.baseUrl, common.ApiNotificationCleanupRoute)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// DeleteProcessedNotificationsByAge removes processed notifications that are older than age. And the corresponding transmissions will also be deleted.
// Age is supposed in milliseconds since modified timestamp
// Please notice that this API is only for processed notifications (status = PROCESSED). If the deletion purpose includes each kind of notifications, please refer to cleanup API.
func (client *NotificationClient) DeleteProcessedNotificationsByAge(ctx context.Context, age int) (res dtoCommon.BaseResponse, err errors.EdgeX) {
	path := path.Join(common.ApiNotificationRoute, common.Age, strconv.Itoa(age))
	err = utils.DeleteRequest(ctx, &res, client.baseUrl, path)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

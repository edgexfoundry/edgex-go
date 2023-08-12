//
// Copyright (C) 2021-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"math"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/application"
	notificationContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/labstack/echo/v4"
)

type NotificationController struct {
	reader io.DtoReader
	dic    *di.Container
}

// NewNotificationController creates and initializes an NotificationController
func NewNotificationController(dic *di.Container) *NotificationController {
	return &NotificationController{
		reader: io.NewJsonDtoReader(),
		dic:    dic,
	}
}

func (nc *NotificationController) AddNotification(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(nc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.AddNotificationRequest
	err := nc.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	notifications := requestDTO.AddNotificationReqToNotificationModels(reqDTOs)

	var addResponses []interface{}
	for i, n := range notifications {
		var response interface{}
		reqId := reqDTOs[i].RequestId
		newId, err := application.AddNotification(n, ctx, nc.dic)
		if err != nil {
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(
				reqId,
				err.Message(),
				err.Code())
		} else {
			response = commonDTO.NewBaseWithIdResponse(
				reqId,
				"",
				http.StatusCreated,
				newId)
		}
		addResponses = append(addResponses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	return pkg.EncodeAndWriteResponse(addResponses, w, lc)
}

func (nc *NotificationController) NotificationById(c echo.Context) error {
	lc := container.LoggingClientFrom(nc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	id := c.Param(common.Id)

	notification, err := application.NotificationById(id, nc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewNotificationResponse("", "", http.StatusOK, notification)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (nc *NotificationController) NotificationsByCategory(c echo.Context) error {
	lc := container.LoggingClientFrom(nc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(nc.dic.Get)

	category := c.Param(common.Category)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	notifications, totalCount, err := application.NotificationsByCategory(offset, limit, category, nc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiNotificationsResponse("", "", http.StatusOK, totalCount, notifications)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (nc *NotificationController) NotificationsByLabel(c echo.Context) error {
	lc := container.LoggingClientFrom(nc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(nc.dic.Get)

	label := c.Param(common.Label)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	notifications, totalCount, err := application.NotificationsByLabel(offset, limit, label, nc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiNotificationsResponse("", "", http.StatusOK, totalCount, notifications)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (nc *NotificationController) NotificationsByStatus(c echo.Context) error {
	lc := container.LoggingClientFrom(nc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(nc.dic.Get)

	status := c.Param(common.Status)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	notifications, totalCount, err := application.NotificationsByStatus(offset, limit, status, nc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiNotificationsResponse("", "", http.StatusOK, totalCount, notifications)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (nc *NotificationController) NotificationsByTimeRange(c echo.Context) error {
	lc := container.LoggingClientFrom(nc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(nc.dic.Get)

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	notifications, totalCount, err := application.NotificationsByTimeRange(start, end, offset, limit, nc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiNotificationsResponse("", "", http.StatusOK, totalCount, notifications)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// DeleteNotificationById deletes the notification by id and all of its associated transmissions
func (nc *NotificationController) DeleteNotificationById(c echo.Context) error {
	lc := container.LoggingClientFrom(nc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	id := c.Param(common.Id)

	err := application.DeleteNotificationById(id, nc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// NotificationsBySubscriptionName queries notifications by offset, limit and subscriptionName
func (nc *NotificationController) NotificationsBySubscriptionName(c echo.Context) error {
	lc := container.LoggingClientFrom(nc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(nc.dic.Get)

	subscriptionName := c.Param(common.Name)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	notifications, totalCount, err := application.NotificationsBySubscriptionName(offset, limit, subscriptionName, nc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiNotificationsResponse("", "", http.StatusOK, totalCount, notifications)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// CleanupNotificationsByAge deletes notifications which have age and is less than the specified one, where the age of Notification is calculated by subtracting its last modification timestamp from the current timestamp. Note that the corresponding transmissions will also be deleted.
func (nc *NotificationController) CleanupNotificationsByAge(c echo.Context) error {
	lc := container.LoggingClientFrom(nc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	age, parsingErr := strconv.ParseInt(c.Param(common.Age), 10, 64)
	if parsingErr != nil {
		err := errors.NewCommonEdgeX(errors.KindContractInvalid, "age format parsing failed", parsingErr)
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	err := application.CleanupNotificationsByAge(age, nc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusAccepted)
	utils.WriteHttpHeader(w, ctx, http.StatusAccepted)
	// encode and send out the response
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// CleanupNotifications deletes all notifications and the corresponding transmissions.
func (nc *NotificationController) CleanupNotifications(c echo.Context) error {
	lc := container.LoggingClientFrom(nc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// Use zero as the age to delete all
	err := application.CleanupNotificationsByAge(0, nc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusAccepted)
	utils.WriteHttpHeader(w, ctx, http.StatusAccepted)
	// encode and send out the response
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// DeleteProcessedNotificationsByAge deletes the processed notifications if the current timestamp minus their last modification timestamp is less than the age parameter, and the corresponding transmissions will also be deleted.
// Please notice that this API is only for processed notifications (status = PROCESSED). If the deletion purpose includes each kind of notifications, please refer to /cleanup API.
func (nc *NotificationController) DeleteProcessedNotificationsByAge(c echo.Context) error {
	lc := container.LoggingClientFrom(nc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	age, parsingErr := strconv.ParseInt(c.Param(common.Age), 10, 64)
	if parsingErr != nil {
		err := errors.NewCommonEdgeX(errors.KindContractInvalid, "age format parsing failed", parsingErr)
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	err := application.DeleteProcessedNotificationsByAge(age, nc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusAccepted)
	utils.WriteHttpHeader(w, ctx, http.StatusAccepted)
	// encode and send out the response
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

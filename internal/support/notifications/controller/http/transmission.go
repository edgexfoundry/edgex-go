//
// Copyright (C) 2021-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"math"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/application"
	notificationContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/labstack/echo/v4"
)

type TransmissionController struct {
	dic *di.Container
}

// NewTransmissionController creates and initializes an TransmissionController
func NewTransmissionController(dic *di.Container) *TransmissionController {
	return &TransmissionController{
		dic: dic,
	}
}

// TransmissionById queries transmission by ID
func (tc *TransmissionController) TransmissionById(c echo.Context) error {
	lc := container.LoggingClientFrom(tc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	id := c.Param(common.Id)

	trans, err := application.TransmissionById(id, tc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewTransmissionResponse("", "", http.StatusOK, trans)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// TransmissionsByTimeRange allows querying of transmissions by their creation timestamp within a given time range, sorted in descending order. Results are paginated.
func (tc *TransmissionController) TransmissionsByTimeRange(c echo.Context) error {
	lc := container.LoggingClientFrom(tc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(tc.dic.Get)

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	transmissions, totalCount, err := application.TransmissionsByTimeRange(start, end, offset, limit, tc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiTransmissionsResponse("", "", http.StatusOK, totalCount, transmissions)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (tc *TransmissionController) AllTransmissions(c echo.Context) error {
	lc := container.LoggingClientFrom(tc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(tc.dic.Get)

	// parse URL query string for offset and limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	transmissions, totalCount, err := application.AllTransmissions(offset, limit, tc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiTransmissionsResponse("", "", http.StatusOK, totalCount, transmissions)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// TransmissionsByStatus allows retrieval of the transmissions associated with the specified status. Ordered by create timestamp descending.
func (tc *TransmissionController) TransmissionsByStatus(c echo.Context) error {
	lc := container.LoggingClientFrom(tc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(tc.dic.Get)

	status := c.Param(common.Status)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	transmissions, totalCount, err := application.TransmissionsByStatus(offset, limit, status, tc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiTransmissionsResponse("", "", http.StatusOK, totalCount, transmissions)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// DeleteProcessedTransmissionsByAge deletes the processed transmissions if the current timestamp minus their created timestamp is less than the age parameter.
func (nc *TransmissionController) DeleteProcessedTransmissionsByAge(c echo.Context) error {
	lc := container.LoggingClientFrom(nc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	age, parsingErr := strconv.ParseInt(c.Param(common.Age), 10, 64)
	if parsingErr != nil {
		err := errors.NewCommonEdgeX(errors.KindContractInvalid, "age format parsing failed", parsingErr)
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	err := application.DeleteProcessedTransmissionsByAge(age, nc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusAccepted)
	utils.WriteHttpHeader(w, ctx, http.StatusAccepted)
	// encode and send out the response
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// TransmissionsBySubscriptionName allows retrieval of the transmissions associated with the specified subscription name. Ordered by create timestamp descending.
func (tc *TransmissionController) TransmissionsBySubscriptionName(c echo.Context) error {
	lc := container.LoggingClientFrom(tc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(tc.dic.Get)

	subscriptionName := c.Param(common.Name)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	transmissions, totalCount, err := application.TransmissionsBySubscriptionName(offset, limit, subscriptionName, tc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiTransmissionsResponse("", "", http.StatusOK, totalCount, transmissions)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// TransmissionsByNotificationId queries transmission by Notification ID
func (tc *TransmissionController) TransmissionsByNotificationId(c echo.Context) error {
	lc := container.LoggingClientFrom(tc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(tc.dic.Get)

	// URL parameters
	notificationId := c.Param(common.Id)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	transmissions, totalCount, err := application.TransmissionsByNotificationId(offset, limit, notificationId, tc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiTransmissionsResponse("", "", http.StatusOK, totalCount, transmissions)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

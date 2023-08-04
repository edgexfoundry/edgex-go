//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"math"
	"net/http"

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

	"github.com/labstack/echo/v4"
)

type SubscriptionController struct {
	reader io.DtoReader
	dic    *di.Container
}

// NewSubscriptionController creates and initializes an SubscriptionController
func NewSubscriptionController(dic *di.Container) *SubscriptionController {
	return &SubscriptionController{
		reader: io.NewJsonDtoReader(),
		dic:    dic,
	}
}

func (sc *SubscriptionController) AddSubscription(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(sc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.AddSubscriptionRequest
	err := sc.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	subscriptions := requestDTO.AddSubscriptionReqToSubscriptionModels(reqDTOs)

	var addResponses []interface{}
	for i, s := range subscriptions {
		var response interface{}
		reqId := reqDTOs[i].RequestId
		newId, err := application.AddSubscription(s, ctx, sc.dic)
		if err != nil {
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(reqId, err.Message(), err.Code())
		} else {
			response = commonDTO.NewBaseWithIdResponse(reqId, "", http.StatusCreated, newId)
		}
		addResponses = append(addResponses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	return pkg.EncodeAndWriteResponse(addResponses, w, lc)
}

func (sc *SubscriptionController) AllSubscriptions(c echo.Context) error {
	lc := container.LoggingClientFrom(sc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(sc.dic.Get)

	// parse URL query string for offset and limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	subscriptions, totalCount, err := application.AllSubscriptions(offset, limit, sc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiSubscriptionsResponse("", "", http.StatusOK, totalCount, subscriptions)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (sc *SubscriptionController) SubscriptionByName(c echo.Context) error {
	lc := container.LoggingClientFrom(sc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	name := c.Param(common.Name)

	subscription, err := application.SubscriptionByName(name, sc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewSubscriptionResponse("", "", http.StatusOK, subscription)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (sc *SubscriptionController) SubscriptionsByCategory(c echo.Context) error {
	lc := container.LoggingClientFrom(sc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(sc.dic.Get)

	category := c.Param(common.Category)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	subscriptions, totalCount, err := application.SubscriptionsByCategory(offset, limit, category, sc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiSubscriptionsResponse("", "", http.StatusOK, totalCount, subscriptions)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (sc *SubscriptionController) SubscriptionsByLabel(c echo.Context) error {
	lc := container.LoggingClientFrom(sc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(sc.dic.Get)

	label := c.Param(common.Label)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	subscriptions, totalCount, err := application.SubscriptionsByLabel(offset, limit, label, sc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiSubscriptionsResponse("", "", http.StatusOK, totalCount, subscriptions)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (sc *SubscriptionController) SubscriptionsByReceiver(c echo.Context) error {
	lc := container.LoggingClientFrom(sc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(sc.dic.Get)

	receiver := c.Param(common.Receiver)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	subscriptions, totalCount, err := application.SubscriptionsByReceiver(offset, limit, receiver, sc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiSubscriptionsResponse("", "", http.StatusOK, totalCount, subscriptions)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (sc *SubscriptionController) DeleteSubscriptionByName(c echo.Context) error {
	lc := container.LoggingClientFrom(sc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	name := c.Param(common.Name)

	err := application.DeleteSubscriptionByName(name, ctx, sc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (sc *SubscriptionController) PatchSubscription(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(sc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.UpdateSubscriptionRequest
	err := sc.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	var updateResponses []interface{}
	for _, dto := range reqDTOs {
		var response interface{}
		reqId := dto.RequestId
		err := application.PatchSubscription(ctx, dto.Subscription, sc.dic)
		if err != nil {
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(reqId, err.Message(), err.Code())
		} else {
			response = commonDTO.NewBaseResponse(reqId, "", http.StatusOK)
		}
		updateResponses = append(updateResponses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	return pkg.EncodeAndWriteResponse(updateResponses, w, lc)
}

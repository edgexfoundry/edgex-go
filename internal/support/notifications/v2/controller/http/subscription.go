//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"math"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	notificationContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/io"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
)

type SubscriptionController struct {
	reader io.SubscriptionReader
	dic    *di.Container
}

// NewSubscriptionController creates and initializes an SubscriptionController
func NewSubscriptionController(dic *di.Container) *SubscriptionController {
	return &SubscriptionController{
		reader: io.NewSubscriptionRequestReader(),
		dic:    dic,
	}
}

func (sc *SubscriptionController) AddSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(sc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	addSubscriptionDTOs, err := sc.reader.ReadAddSubscriptionRequest(r.Body)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		errResponses := commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		utils.WriteHttpHeader(w, ctx, err.Code())
		pkg.Encode(errResponses, w, lc)
		return
	}
	subscriptions := requestDTO.AddSubscriptionReqToSubscriptionModels(addSubscriptionDTOs)

	var addResponses []interface{}
	for i, s := range subscriptions {
		var response interface{}
		reqId := addSubscriptionDTOs[i].RequestId
		newId, err := application.AddSubscription(s, ctx, sc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
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
	pkg.Encode(addResponses, w, lc)
}

func (sc *SubscriptionController) AllSubscriptions(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(sc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	config := notificationContainer.ConfigurationFrom(sc.dic.Get)

	var response interface{}
	var statusCode int

	// parse URL query string for offset and limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxUint32, -1, config.Service.MaxResultCount)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		subscriptions, err := application.AllSubscriptions(offset, limit, sc.dic)
		if err != nil {
			if errors.Kind(err) != errors.KindEntityDoesNotExist {
				lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			}
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
			statusCode = err.Code()
		} else {
			response = responseDTO.NewMultiSubscriptionsResponse("", "", http.StatusOK, subscriptions)
			statusCode = http.StatusOK
		}
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

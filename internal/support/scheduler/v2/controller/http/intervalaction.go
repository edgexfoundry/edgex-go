//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"math"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	schedulerContainer "github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/io"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"

	"github.com/gorilla/mux"
)

type IntervalActionController struct {
	reader io.IntervalActionReader
	dic    *di.Container
}

// NewIntervalActionController creates and initializes an IntervalActionController
func NewIntervalActionController(dic *di.Container) *IntervalActionController {
	return &IntervalActionController{
		reader: io.NewIntervalActionRequestReader(),
		dic:    dic,
	}
}

func (ic *IntervalActionController) AddIntervalAction(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(ic.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	actionDTOs, err := ic.reader.ReadAddIntervalActionRequest(r.Body)
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
	actions := requestDTO.AddIntervalActionReqToIntervalActionModels(actionDTOs)

	var addResponses []interface{}
	for i, action := range actions {
		var response interface{}
		reqId := actionDTOs[i].RequestId
		newId, err := application.AddIntervalAction(action, ctx, ic.dic)
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

		// TODO Add the new IntervalAction into scheduler queue
		//err = scClient.AddIntervalActionToQueue(intervalAction)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(addResponses, w, lc)
}

func (ic *IntervalActionController) AllIntervalActions(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(ic.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	config := schedulerContainer.ConfigurationFrom(ic.dic.Get)

	var response interface{}
	var statusCode int

	// parse URL query string for offset, limit, and labels
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		intervalActions, err := application.AllIntervalActions(offset, limit, ic.dic)
		if err != nil {
			if errors.Kind(err) != errors.KindEntityDoesNotExist {
				lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			}
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
			statusCode = err.Code()
		} else {
			response = responseDTO.NewMultiIntervalActionsResponse("", "", http.StatusOK, intervalActions)
			statusCode = http.StatusOK
		}
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (ic *IntervalActionController) IntervalActionByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(ic.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	name := vars[v2.Name]

	var response interface{}
	var statusCode int

	action, err := application.IntervalActionByName(name, ctx, ic.dic)
	if err != nil {
		if errors.Kind(err) != errors.KindEntityDoesNotExist {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		}
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = responseDTO.NewIntervalActionResponse("", "", http.StatusOK, action)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (ic *IntervalActionController) DeleteIntervalActionByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(ic.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	name := vars[v2.Name]

	var response interface{}
	var statusCode int

	err := application.DeleteIntervalActionByName(name, ctx, ic.dic)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = commonDTO.NewBaseResponse("", "", http.StatusOK)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (ic *IntervalActionController) PatchIntervalAction(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(ic.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	reqDTOs, err := ic.reader.ReadUpdateIntervalActionRequest(r.Body)
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

	var responses []interface{}
	for _, dto := range reqDTOs {
		var response interface{}
		reqId := dto.RequestId
		err := application.PatchIntervalAction(dto.Action, ctx, ic.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(
				reqId,
				err.Message(),
				err.Code())
		} else {
			response = commonDTO.NewBaseResponse(
				reqId,
				"",
				http.StatusOK)
		}
		responses = append(responses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(responses, w, lc)
}

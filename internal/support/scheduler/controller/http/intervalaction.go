//
// Copyright (C) 2021 IOTech Ltd
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
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/application"
	schedulerContainer "github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"

	"github.com/gorilla/mux"
)

type IntervalActionController struct {
	reader io.DtoReader
	dic    *di.Container
}

// NewIntervalActionController creates and initializes an IntervalActionController
func NewIntervalActionController(dic *di.Container) *IntervalActionController {
	return &IntervalActionController{
		reader: io.NewJsonDtoReader(),
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

	var reqDTOs []requestDTO.AddIntervalActionRequest
	err := ic.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	actions := requestDTO.AddIntervalActionReqToIntervalActionModels(reqDTOs)

	var addResponses []interface{}
	for i, action := range actions {
		var response interface{}
		reqId := reqDTOs[i].RequestId
		newId, err := application.AddIntervalAction(action, ctx, ic.dic)
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
	pkg.EncodeAndWriteResponse(addResponses, w, lc)
}

func (ic *IntervalActionController) AllIntervalActions(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(ic.dic.Get)
	ctx := r.Context()
	config := schedulerContainer.ConfigurationFrom(ic.dic.Get)

	// parse URL query string for offset, limit, and labels
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	intervalActions, totalCount, err := application.AllIntervalActions(offset, limit, ic.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiIntervalActionsResponse("", "", http.StatusOK, totalCount, intervalActions)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ic *IntervalActionController) IntervalActionByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(ic.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	name := vars[common.Name]

	action, err := application.IntervalActionByName(name, ctx, ic.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewIntervalActionResponse("", "", http.StatusOK, action)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ic *IntervalActionController) DeleteIntervalActionByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(ic.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	name := vars[common.Name]

	err := application.DeleteIntervalActionByName(name, ctx, ic.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ic *IntervalActionController) PatchIntervalAction(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(ic.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.UpdateIntervalActionRequest
	err := ic.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	var responses []interface{}
	for _, dto := range reqDTOs {
		var response interface{}
		reqId := dto.RequestId
		err := application.PatchIntervalAction(dto.Action, ctx, ic.dic)
		if err != nil {
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(reqId, err.Message(), err.Code())
		} else {
			response = commonDTO.NewBaseResponse(reqId, "", http.StatusOK)
		}
		responses = append(responses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.EncodeAndWriteResponse(responses, w, lc)
}

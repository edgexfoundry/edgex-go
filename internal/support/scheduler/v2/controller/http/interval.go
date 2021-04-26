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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"

	"github.com/gorilla/mux"
)

type IntervalController struct {
	reader io.IntervalReader
	dic    *di.Container
}

// NewIntervalController creates and initializes an IntervalController
func NewIntervalController(dic *di.Container) *IntervalController {
	return &IntervalController{
		reader: io.NewIntervalRequestReader(),
		dic:    dic,
	}
}

func (dc *IntervalController) AddInterval(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	addIntervalDTOs, err := dc.reader.ReadAddIntervalRequest(r.Body)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	Intervals := requestDTO.AddIntervalReqToIntervalModels(addIntervalDTOs)

	var addResponses []interface{}
	for i, d := range Intervals {
		var response interface{}
		reqId := addIntervalDTOs[i].RequestId
		newId, err := application.AddInterval(d, ctx, dc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(reqId, err.Message(), err.Code())
		} else {
			response = commonDTO.NewBaseWithIdResponse(reqId, "", http.StatusCreated, newId)
		}
		addResponses = append(addResponses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(addResponses, w, lc)
}

func (dc *IntervalController) IntervalByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	name := vars[v2.Name]

	interval, err := application.IntervalByName(name, ctx, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewIntervalResponse("", "", http.StatusOK, interval)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (dc *IntervalController) AllIntervals(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	config := schedulerContainer.ConfigurationFrom(dc.dic.Get)

	// parse URL query string for offset, limit, and labels
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	intervals, err := application.AllIntervals(offset, limit, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiIntervalsResponse("", "", http.StatusOK, intervals)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (dc *IntervalController) DeleteIntervalByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	name := vars[v2.Name]

	err := application.DeleteIntervalByName(name, ctx, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (dc *IntervalController) PatchInterval(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	reqDTOs, err := dc.reader.ReadUpdateIntervalRequest(r.Body)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	var responses []interface{}
	for _, dto := range reqDTOs {
		var response interface{}
		reqId := dto.RequestId
		err := application.PatchInterval(dto.Interval, ctx, dc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(reqId, err.Message(), err.Code())
		} else {
			response = commonDTO.NewBaseResponse(reqId, "", http.StatusOK)
		}
		responses = append(responses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(responses, w, lc)
}

//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/io"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
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
	Intervals := requestDTO.AddIntervalReqToIntervalModels(addIntervalDTOs)

	var addResponses []interface{}
	for i, d := range Intervals {
		var response interface{}
		reqId := addIntervalDTOs[i].RequestId
		newId, err := application.AddInterval(d, ctx, dc.dic)
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

		// TODO Push the new interval into scheduler queue
		//interval.ID = newId
		//err = scClient.AddIntervalToQueue(interval)
		//if err != nil {
		//	return "", err
		//}
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(addResponses, w, lc)
}

func (dc *IntervalController) IntervalByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	name := vars[contractsV2.Name]

	var response interface{}
	var statusCode int

	interval, err := application.IntervalByName(name, ctx, dc.dic)
	if err != nil {
		if errors.Kind(err) != errors.KindEntityDoesNotExist {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		}
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = responseDTO.NewIntervalResponse("", "", http.StatusOK, interval)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

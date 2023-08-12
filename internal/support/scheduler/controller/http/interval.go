//
// Copyright (C) 2021-2023 IOTech Ltd
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

	"github.com/labstack/echo/v4"
)

type IntervalController struct {
	reader io.DtoReader
	dic    *di.Container
}

// NewIntervalController creates and initializes an IntervalController
func NewIntervalController(dic *di.Container) *IntervalController {
	return &IntervalController{
		reader: io.NewJsonDtoReader(),
		dic:    dic,
	}
}

func (dc *IntervalController) AddInterval(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.AddIntervalRequest
	err := dc.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	Intervals := requestDTO.AddIntervalReqToIntervalModels(reqDTOs)

	var addResponses []interface{}
	for i, d := range Intervals {
		var response interface{}
		reqId := reqDTOs[i].RequestId
		newId, err := application.AddInterval(d, ctx, dc.dic)
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

func (dc *IntervalController) IntervalByName(c echo.Context) error {
	lc := container.LoggingClientFrom(dc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	name := c.Param(common.Name)

	interval, err := application.IntervalByName(name, ctx, dc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewIntervalResponse("", "", http.StatusOK, interval)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (dc *IntervalController) AllIntervals(c echo.Context) error {
	lc := container.LoggingClientFrom(dc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := schedulerContainer.ConfigurationFrom(dc.dic.Get)

	// parse URL query string for offset, limit, and labels
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	intervals, totalCount, err := application.AllIntervals(offset, limit, dc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiIntervalsResponse("", "", http.StatusOK, totalCount, intervals)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (dc *IntervalController) DeleteIntervalByName(c echo.Context) error {
	lc := container.LoggingClientFrom(dc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	name := c.Param(common.Name)

	err := application.DeleteIntervalByName(name, ctx, dc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (dc *IntervalController) PatchInterval(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.UpdateIntervalRequest
	err := dc.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	var responses []interface{}
	for _, dto := range reqDTOs {
		var response interface{}
		reqId := dto.RequestId
		err := application.PatchInterval(dto.Interval, ctx, dc.dic)
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
	return pkg.EncodeAndWriteResponse(responses, w, lc)
}

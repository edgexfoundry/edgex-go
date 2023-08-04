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

func (ic *IntervalActionController) AddIntervalAction(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(ic.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.AddIntervalActionRequest
	err := ic.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
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
	return pkg.EncodeAndWriteResponse(addResponses, w, lc)
}

func (ic *IntervalActionController) AllIntervalActions(c echo.Context) error {
	lc := container.LoggingClientFrom(ic.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := schedulerContainer.ConfigurationFrom(ic.dic.Get)

	// parse URL query string for offset, limit, and labels
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	intervalActions, totalCount, err := application.AllIntervalActions(offset, limit, ic.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiIntervalActionsResponse("", "", http.StatusOK, totalCount, intervalActions)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ic *IntervalActionController) IntervalActionByName(c echo.Context) error {
	lc := container.LoggingClientFrom(ic.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	name := c.Param(common.Name)

	action, err := application.IntervalActionByName(name, ctx, ic.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewIntervalActionResponse("", "", http.StatusOK, action)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ic *IntervalActionController) DeleteIntervalActionByName(c echo.Context) error {
	lc := container.LoggingClientFrom(ic.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	name := c.Param(common.Name)

	err := application.DeleteIntervalActionByName(name, ctx, ic.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ic *IntervalActionController) PatchIntervalAction(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(ic.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.UpdateIntervalActionRequest
	err := ic.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
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
	return pkg.EncodeAndWriteResponse(responses, w, lc)
}

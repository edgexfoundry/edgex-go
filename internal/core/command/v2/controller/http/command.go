//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"fmt"
	"math"
	"net/http"

	commandContainer "github.com/edgexfoundry/edgex-go/internal/core/command/container"
	"github.com/edgexfoundry/edgex-go/internal/core/command/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"

	"github.com/gorilla/mux"
)

type CommandController struct {
	dic *di.Container
}

// NewCommandController creates and initializes an CommandController
func NewCommandController(dic *di.Container) *CommandController {
	return &CommandController{
		dic: dic,
	}
}

func (cc *CommandController) AllCommands(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(cc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	config := commandContainer.ConfigurationFrom(cc.dic.Get)

	var response interface{}
	var statusCode int

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		commands, err := application.AllCommands(offset, limit, cc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
			statusCode = err.Code()
		} else {
			response = responseDTO.NewMultiDeviceCoreCommandsResponse("", "", http.StatusOK, commands)
			statusCode = http.StatusOK
		}
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	// encode and send out the response
	pkg.Encode(response, w, lc)
}

func (cc *CommandController) CommandsByDeviceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(cc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	name := vars[v2.Name]

	var response interface{}
	var statusCode int

	deviceCoreCommand, err := application.CommandsByDeviceName(name, cc.dic)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = responseDTO.NewDeviceCoreCommandResponse("", "", http.StatusOK, deviceCoreCommand)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	// encode and send out the response
	pkg.Encode(response, w, lc)
}

func validateGetCommandParameters(r *http.Request) (err errors.EdgeX) {
	dsReturnEvent := utils.ParseQueryStringToString(r, v2.ReturnEvent, v2.ValueYes)
	dsPushEvent := utils.ParseQueryStringToString(r, v2.PushEvent, v2.ValueNo)
	if dsReturnEvent != v2.ValueYes && dsReturnEvent != v2.ValueNo {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("invalid query parameter, %s has to be %s or %s", dsReturnEvent, v2.ValueYes, v2.ValueNo), nil)
	}
	if dsPushEvent != v2.ValueYes && dsPushEvent != v2.ValueNo {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("invalid query parameter, %s has to be %s or %s", dsPushEvent, v2.ValueYes, v2.ValueNo), nil)
	}
	return nil
}

func (cc *CommandController) IssueGetCommandByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(cc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	deviceName := vars[v2.Name]
	commandName := vars[v2.Command]

	// Query params
	queryParams := r.URL.RawQuery
	err := validateGetCommandParameters(r)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		errResponses := commonDTO.NewBaseResponse("", err.Message(), err.Code())
		utils.WriteHttpHeader(w, ctx, err.Code())
		pkg.Encode(errResponses, w, lc)
		return
	}

	response, err := application.IssueGetCommandByName(deviceName, commandName, queryParams, cc.dic)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		errResponses := commonDTO.NewBaseResponse("", err.Message(), err.Code())
		utils.WriteHttpHeader(w, ctx, err.Code())
		pkg.Encode(errResponses, w, lc)
		return
	}

	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	// encode and send out the response
	// If dsReturnEvent is no, there will be no content returned in the http response
	if response != nil {
		pkg.Encode(response, w, lc)
	}
}

func (cc *CommandController) IssueSetCommandByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(cc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	deviceName := vars[v2.Name]
	commandName := vars[v2.Command]

	// Query params
	queryParams := r.URL.RawQuery

	var response commonDTO.BaseResponse
	var statusCode int

	// Request body
	settings, err := utils.ParseBodyToMap(r)
	if err == nil {
		response, err = application.IssueSetCommandByName(deviceName, commandName, queryParams, settings, cc.dic)
	}

	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		statusCode = response.StatusCode
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	// encode and send out the response
	pkg.Encode(response, w, lc)
}

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
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
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
	config := commandContainer.ConfigurationFrom(cc.dic.Get)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	commands, err := application.AllCommands(offset, limit, cc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiDeviceCoreCommandsResponse("", "", http.StatusOK, commands)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	// encode and send out the response
	pkg.Encode(response, w, lc)
}

func (cc *CommandController) CommandsByDeviceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(cc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	name := vars[v2.Name]

	deviceCoreCommand, err := application.CommandsByDeviceName(name, cc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewDeviceCoreCommandResponse("", "", http.StatusOK, deviceCoreCommand)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
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

	// URL parameters
	vars := mux.Vars(r)
	deviceName := vars[v2.Name]
	commandName := vars[v2.Command]

	// Query params
	queryParams := r.URL.RawQuery
	err := validateGetCommandParameters(r)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response, err := application.IssueGetCommandByName(deviceName, commandName, queryParams, cc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
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

	// URL parameters
	vars := mux.Vars(r)
	deviceName := vars[v2.Name]
	commandName := vars[v2.Command]

	// Query params
	queryParams := r.URL.RawQuery

	// Request body
	settings, err := utils.ParseBodyToMap(r)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	response, err := application.IssueSetCommandByName(deviceName, commandName, queryParams, settings, cc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	utils.WriteHttpHeader(w, ctx, response.StatusCode)
	// encode and send out the response
	pkg.Encode(response, w, lc)
}

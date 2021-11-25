//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"fmt"
	"math"
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/edgexfoundry/edgex-go/internal/core/command/application"
	commandContainer "github.com/edgexfoundry/edgex-go/internal/core/command/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"

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
	commands, totalCount, err := application.AllCommands(offset, limit, cc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiDeviceCoreCommandsResponse("", "", http.StatusOK, totalCount, commands)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	// encode and send out the response
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (cc *CommandController) CommandsByDeviceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(cc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	name := vars[common.Name]

	deviceCoreCommand, err := application.CommandsByDeviceName(name, cc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewDeviceCoreCommandResponse("", "", http.StatusOK, deviceCoreCommand)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	// encode and send out the response
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func validateGetCommandParameters(r *http.Request) (err errors.EdgeX) {
	dsReturnEvent := utils.ParseQueryStringToString(r, common.ReturnEvent, common.ValueYes)
	dsPushEvent := utils.ParseQueryStringToString(r, common.PushEvent, common.ValueNo)
	if dsReturnEvent != common.ValueYes && dsReturnEvent != common.ValueNo {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("invalid query parameter, %s has to be %s or %s", dsReturnEvent, common.ValueYes, common.ValueNo), nil)
	}
	if dsPushEvent != common.ValueYes && dsPushEvent != common.ValueNo {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("invalid query parameter, %s has to be %s or %s", dsPushEvent, common.ValueYes, common.ValueNo), nil)
	}
	return nil
}

func (cc *CommandController) IssueGetCommandByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(cc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	deviceName := vars[common.Name]
	commandName := vars[common.Command]

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
	// encode and send out the response
	if response != nil {
		utils.WriteHttpHeader(w, ctx, response.StatusCode)
		pkg.EncodeAndWriteResponse(response, w, lc)
	} else {
		// If dsReturnEvent is no, there will be no content returned in the http response
		utils.WriteHttpHeader(w, ctx, http.StatusOK)
	}
}

func (cc *CommandController) IssueSetCommandByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(cc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	deviceName := vars[common.Name]
	commandName := vars[common.Command]

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
	pkg.EncodeAndWriteResponse(response, w, lc)
}

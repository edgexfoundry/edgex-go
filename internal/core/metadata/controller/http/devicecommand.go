//
// Copyright (C) 2022-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/application"
	"github.com/edgexfoundry/edgex-go/internal/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"

	"github.com/labstack/echo/v4"
)

type DeviceCommandController struct {
	reader io.DtoReader
	dic    *di.Container
}

// NewDeviceCommandController creates and initializes an DeviceCommandController
func NewDeviceCommandController(dic *di.Container) *DeviceCommandController {
	return &DeviceCommandController{
		reader: io.NewJsonDtoReader(),
		dic:    dic,
	}
}

func (dc *DeviceCommandController) AddDeviceProfileDeviceCommand(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.AddDeviceCommandRequest
	err := dc.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	var addResponses []interface{}
	for _, dto := range reqDTOs {
		var response interface{}
		reqId := dto.RequestId
		profileName := dto.ProfileName
		deviceCommand := dtos.ToDeviceCommandModel(dto.DeviceCommand)
		err = application.AddDeviceProfileDeviceCommand(profileName, deviceCommand, ctx, dc.dic)
		if err != nil {
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(
				reqId,
				err.Message(),
				err.Code())
		} else {
			response = commonDTO.NewBaseResponse(
				reqId,
				"",
				http.StatusCreated)

		}
		addResponses = append(addResponses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	return pkg.EncodeAndWriteResponse(addResponses, w, lc)
}

func (dc *DeviceCommandController) PatchDeviceProfileDeviceCommand(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.UpdateDeviceCommandRequest
	err := dc.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	var updateResponses []interface{}
	for _, dto := range reqDTOs {
		var response interface{}
		reqId := dto.RequestId
		profileName := dto.ProfileName
		err = application.PatchDeviceProfileDeviceCommand(profileName, dto.DeviceCommand, ctx, dc.dic)
		if err != nil {
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
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
		updateResponses = append(updateResponses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	return pkg.EncodeAndWriteResponse(updateResponses, w, lc)
}

func (dc *DeviceCommandController) DeleteDeviceCommandByName(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()

	// URL parameters
	profileName := c.Param(common.Name)
	commandName := c.Param(common.CommandName)

	err := application.DeleteDeviceCommandByName(profileName, commandName, ctx, dc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

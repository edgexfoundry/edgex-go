//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"math"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/application"
	metadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"

	"github.com/gorilla/mux"
)

type DeviceServiceController struct {
	reader io.DtoReader
	dic    *di.Container
}

// NewDeviceServiceController creates and initializes an DeviceServiceController
func NewDeviceServiceController(dic *di.Container) *DeviceServiceController {
	return &DeviceServiceController{
		reader: io.NewJsonDtoReader(),
		dic:    dic,
	}
}

func (dc *DeviceServiceController) AddDeviceService(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.AddDeviceServiceRequest
	err := dc.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	deviceServices := requestDTO.AddDeviceServiceReqToDeviceServiceModels(reqDTOs)

	var addResponses []interface{}
	for i, d := range deviceServices {
		var addDeviceServiceResponse interface{}
		reqId := reqDTOs[i].RequestId
		newId, err := application.AddDeviceService(d, ctx, dc.dic)
		if err == nil {
			addDeviceServiceResponse = commonDTO.NewBaseWithIdResponse(
				reqId,
				"",
				http.StatusCreated,
				newId)
		} else {
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
			addDeviceServiceResponse = commonDTO.NewBaseResponse(
				reqId,
				err.Error(),
				err.Code())
		}
		addResponses = append(addResponses, addDeviceServiceResponse)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	// EncodeAndWriteResponse and send the resp body as JSON format
	pkg.EncodeAndWriteResponse(addResponses, w, lc)
}

func (dc *DeviceServiceController) DeviceServiceByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	name := vars[common.Name]

	deviceService, err := application.DeviceServiceByName(name, ctx, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewDeviceServiceResponse("", "", http.StatusOK, deviceService)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (dc *DeviceServiceController) PatchDeviceService(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.UpdateDeviceServiceRequest
	err := dc.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	var updateResponses []interface{}
	for _, dto := range reqDTOs {
		var response interface{}
		reqId := dto.RequestId
		err := application.PatchDeviceService(dto.Service, ctx, dc.dic)
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
	pkg.EncodeAndWriteResponse(updateResponses, w, lc)
}

func (dc *DeviceServiceController) DeleteDeviceServiceByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	name := vars[common.Name]

	err := application.DeleteDeviceServiceByName(name, ctx, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (dc *DeviceServiceController) AllDeviceServices(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	config := metadataContainer.ConfigurationFrom(dc.dic.Get)

	// parse URL query string for offset, limit, and labels
	offset, limit, labels, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	deviceServices, totalCount, err := application.AllDeviceServices(offset, limit, labels, ctx, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiDeviceServicesResponse("", "", http.StatusOK, totalCount, deviceServices)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	// encode and send out the response
	pkg.EncodeAndWriteResponse(response, w, lc)
}

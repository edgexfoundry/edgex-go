//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"math"
	"net/http"

	metadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"

	"github.com/gorilla/mux"
)

type DeviceProfileController struct {
	reader io.DeviceProfileReader
	dic    *di.Container
}

// NewDeviceProfileController creates and initializes an DeviceProfileController
func NewDeviceProfileController(dic *di.Container) *DeviceProfileController {
	return &DeviceProfileController{
		reader: io.NewDeviceProfileRequestReader(),
		dic:    dic,
	}
}

func (dc *DeviceProfileController) AddDeviceProfile(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	addDeviceProfileDTOs, err := dc.reader.ReadDeviceProfileRequest(r.Body)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	deviceProfiles := requestDTO.DeviceProfileReqToDeviceProfileModels(addDeviceProfileDTOs)

	var addResponses []interface{}
	for i, d := range deviceProfiles {
		var addDeviceProfileResponse interface{}
		// get the requestID from AddDeviceProfileDTO
		reqId := addDeviceProfileDTOs[i].RequestId
		newId, err := application.AddDeviceProfile(d, ctx, dc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			addDeviceProfileResponse = commonDTO.NewBaseResponse(
				reqId,
				err.Message(),
				err.Code())
		} else {
			addDeviceProfileResponse = commonDTO.NewBaseWithIdResponse(
				reqId,
				"",
				http.StatusCreated,
				newId)
		}
		addResponses = append(addResponses, addDeviceProfileResponse)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(addResponses, w, lc)
}

func (dc *DeviceProfileController) UpdateDeviceProfile(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	updateDeviceProfileReq, err := dc.reader.ReadDeviceProfileRequest(r.Body)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	deviceProfiles := requestDTO.DeviceProfileReqToDeviceProfileModels(updateDeviceProfileReq)

	var responses []interface{}
	for i, d := range deviceProfiles {
		var response interface{}
		reqId := updateDeviceProfileReq[i].RequestId
		err := application.UpdateDeviceProfile(d, ctx, dc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
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
		responses = append(responses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(responses, w, lc)
}

func (dc *DeviceProfileController) AddDeviceProfileByYaml(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()

	deviceProfileDTO, err := dc.reader.ReadDeviceProfileYaml(r)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	deviceProfile := dtos.ToDeviceProfileModel(deviceProfileDTO)

	newId, err := application.AddDeviceProfile(deviceProfile, ctx, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewBaseWithIdResponse("", "", http.StatusCreated, newId)
	utils.WriteHttpHeader(w, ctx, http.StatusCreated)
	// Encode and send the resp body as JSON format
	pkg.Encode(response, w, lc)
}

func (dc *DeviceProfileController) UpdateDeviceProfileByYaml(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()

	deviceProfileDTO, err := dc.reader.ReadDeviceProfileYaml(r)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	deviceProfile := dtos.ToDeviceProfileModel(deviceProfileDTO)
	err = application.UpdateDeviceProfile(deviceProfile, ctx, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (dc *DeviceProfileController) DeviceProfileByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	name := vars[v2.Name]

	deviceProfile, err := application.DeviceProfileByName(name, ctx, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewDeviceProfileResponse("", "", http.StatusOK, deviceProfile)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc) // encode and send out the response
}

func (dc *DeviceProfileController) DeleteDeviceProfileByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	name := vars[v2.Name]

	err := application.DeleteDeviceProfileByName(name, ctx, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (dc *DeviceProfileController) AllDeviceProfiles(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	config := metadataContainer.ConfigurationFrom(dc.dic.Get)

	// parse URL query string for offset, limit, and labels
	offset, limit, labels, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	deviceProfiles, err := application.AllDeviceProfiles(offset, limit, labels, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiDeviceProfilesResponse("", "", http.StatusOK, deviceProfiles)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (dc *DeviceProfileController) DeviceProfilesByModel(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	config := metadataContainer.ConfigurationFrom(dc.dic.Get)

	vars := mux.Vars(r)
	model := vars[v2.Model]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	deviceProfiles, err := application.DeviceProfilesByModel(offset, limit, model, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiDeviceProfilesResponse("", "", http.StatusOK, deviceProfiles)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (dc *DeviceProfileController) DeviceProfilesByManufacturer(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	config := metadataContainer.ConfigurationFrom(dc.dic.Get)

	vars := mux.Vars(r)
	manufacturer := vars[v2.Manufacturer]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	deviceProfiles, err := application.DeviceProfilesByManufacturer(offset, limit, manufacturer, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiDeviceProfilesResponse("", "", http.StatusOK, deviceProfiles)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (dc *DeviceProfileController) DeviceProfilesByManufacturerAndModel(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	config := metadataContainer.ConfigurationFrom(dc.dic.Get)

	vars := mux.Vars(r)
	manufacturer := vars[v2.Manufacturer]
	model := vars[v2.Model]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	deviceProfiles, err := application.DeviceProfilesByManufacturerAndModel(offset, limit, manufacturer, model, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiDeviceProfilesResponse("", "", http.StatusOK, deviceProfiles)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

//
// Copyright (C) 2021-2023 IOTech Ltd
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

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/labstack/echo/v4"
)

const yamlFileName = "file"

type DeviceProfileController struct {
	jsonDtoReader io.DtoReader
	yamlDtoReader io.DtoReader
	dic           *di.Container
}

// NewDeviceProfileController creates and initializes an DeviceProfileController
func NewDeviceProfileController(dic *di.Container) *DeviceProfileController {
	return &DeviceProfileController{
		jsonDtoReader: io.NewJsonDtoReader(),
		yamlDtoReader: io.NewYamlDtoReader(),
		dic:           dic,
	}
}

func (dc *DeviceProfileController) AddDeviceProfile(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.DeviceProfileRequest
	err := dc.jsonDtoReader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	deviceProfiles := requestDTO.DeviceProfileReqToDeviceProfileModels(reqDTOs)

	var addResponses []interface{}
	for i, d := range deviceProfiles {
		var addDeviceProfileResponse interface{}
		reqId := reqDTOs[i].RequestId
		newId, err := application.AddDeviceProfile(d, ctx, dc.dic)
		if err != nil {
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
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
	return pkg.EncodeAndWriteResponse(addResponses, w, lc)
}

func (dc *DeviceProfileController) UpdateDeviceProfile(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	strictProfileChanges := metadataContainer.ConfigurationFrom(dc.dic.Get).Writable.ProfileChange.StrictDeviceProfileChanges
	if strictProfileChanges {
		return utils.WriteErrorResponse(w, ctx, lc, errors.NewCommonEdgeX(errors.KindServiceLocked, "profile change is not allowed when StrictDeviceProfileChanges config is enabled", nil), "")
	}

	var reqDTOs []requestDTO.DeviceProfileRequest
	err := dc.jsonDtoReader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	deviceProfiles := requestDTO.DeviceProfileReqToDeviceProfileModels(reqDTOs)

	var responses []interface{}
	for i, d := range deviceProfiles {
		var response interface{}
		reqId := reqDTOs[i].RequestId
		err := application.UpdateDeviceProfile(d, ctx, dc.dic)
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
		responses = append(responses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	return pkg.EncodeAndWriteResponse(responses, w, lc)
}

func (dc *DeviceProfileController) AddDeviceProfileByYaml(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()

	file, _, fileErr := r.FormFile(yamlFileName)
	if fileErr == http.ErrMissingFile {
		return utils.WriteErrorResponse(w, ctx, lc, errors.NewCommonEdgeX(errors.KindContractInvalid, "missing yaml file", nil), "")
	} else if fileErr != nil {
		return utils.WriteErrorResponse(w, ctx, lc, errors.NewCommonEdgeX(errors.KindServerError, fileErr.Error(), nil), "")
	}

	var deviceProfileDTO dtos.DeviceProfile
	err := dc.yamlDtoReader.Read(file, &deviceProfileDTO)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	deviceProfile := dtos.ToDeviceProfileModel(deviceProfileDTO)

	newId, err := application.AddDeviceProfile(deviceProfile, ctx, dc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseWithIdResponse("", "", http.StatusCreated, newId)
	utils.WriteHttpHeader(w, ctx, http.StatusCreated)
	// EncodeAndWriteResponse and send the resp body as JSON format
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (dc *DeviceProfileController) UpdateDeviceProfileByYaml(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()

	file, _, fileErr := r.FormFile(yamlFileName)
	if fileErr == http.ErrMissingFile {
		return utils.WriteErrorResponse(w, ctx, lc, errors.NewCommonEdgeX(errors.KindContractInvalid, "missing yaml file", nil), "")
	} else if fileErr != nil {
		return utils.WriteErrorResponse(w, ctx, lc, errors.NewCommonEdgeX(errors.KindServerError, fileErr.Error(), nil), "")
	}

	strictProfileChanges := metadataContainer.ConfigurationFrom(dc.dic.Get).Writable.ProfileChange.StrictDeviceProfileChanges
	if strictProfileChanges {
		return utils.WriteErrorResponse(w, ctx, lc, errors.NewCommonEdgeX(errors.KindServiceLocked, "profile change is not allowed when StrictDeviceProfileChanges config is enabled", nil), "")
	}

	var deviceProfileDTO dtos.DeviceProfile
	err := dc.yamlDtoReader.Read(file, &deviceProfileDTO)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	deviceProfile := dtos.ToDeviceProfileModel(deviceProfileDTO)
	err = application.UpdateDeviceProfile(deviceProfile, ctx, dc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (dc *DeviceProfileController) DeviceProfileByName(c echo.Context) error {
	lc := container.LoggingClientFrom(dc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	name := c.Param(common.Name)

	deviceProfile, err := application.DeviceProfileByName(name, ctx, dc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewDeviceProfileResponse("", "", http.StatusOK, deviceProfile)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc) // encode and send out the response
}

func (dc *DeviceProfileController) DeleteDeviceProfileByName(c echo.Context) error {
	lc := container.LoggingClientFrom(dc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	name := c.Param(common.Name)

	err := application.DeleteDeviceProfileByName(name, ctx, dc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (dc *DeviceProfileController) AllDeviceProfiles(c echo.Context) error {
	lc := container.LoggingClientFrom(dc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := metadataContainer.ConfigurationFrom(dc.dic.Get)

	// parse URL query string for offset, limit, and labels
	offset, limit, labels, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	deviceProfiles, totalCount, err := application.AllDeviceProfiles(offset, limit, labels, dc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiDeviceProfilesResponse("", "", http.StatusOK, totalCount, deviceProfiles)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (dc *DeviceProfileController) DeviceProfilesByModel(c echo.Context) error {
	lc := container.LoggingClientFrom(dc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := metadataContainer.ConfigurationFrom(dc.dic.Get)

	model := c.Param(common.Model)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	deviceProfiles, totalCount, err := application.DeviceProfilesByModel(offset, limit, model, dc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiDeviceProfilesResponse("", "", http.StatusOK, totalCount, deviceProfiles)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (dc *DeviceProfileController) DeviceProfilesByManufacturer(c echo.Context) error {
	lc := container.LoggingClientFrom(dc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := metadataContainer.ConfigurationFrom(dc.dic.Get)

	manufacturer := c.Param(common.Manufacturer)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	deviceProfiles, totalCount, err := application.DeviceProfilesByManufacturer(offset, limit, manufacturer, dc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiDeviceProfilesResponse("", "", http.StatusOK, totalCount, deviceProfiles)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (dc *DeviceProfileController) DeviceProfilesByManufacturerAndModel(c echo.Context) error {
	lc := container.LoggingClientFrom(dc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := metadataContainer.ConfigurationFrom(dc.dic.Get)

	manufacturer := c.Param(common.Manufacturer)
	model := c.Param(common.Model)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	deviceProfiles, totalCount, err := application.DeviceProfilesByManufacturerAndModel(offset, limit, manufacturer, model, dc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiDeviceProfilesResponse("", "", http.StatusOK, totalCount, deviceProfiles)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (dc *DeviceProfileController) PatchDeviceProfileBasicInfo(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	lc := container.LoggingClientFrom(dc.dic.Get)

	var reqDTOs []requestDTO.DeviceProfileBasicInfoRequest
	err := dc.jsonDtoReader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	var updateResponses []interface{}
	for _, dto := range reqDTOs {
		var response interface{}
		reqId := dto.RequestId
		err := application.PatchDeviceProfileBasicInfo(ctx, dto.BasicInfo, dc.dic)
		if err != nil {
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(reqId, err.Message(), err.Code())
		} else {
			response = commonDTO.NewBaseResponse(reqId, "", http.StatusOK)
		}
		updateResponses = append(updateResponses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	return pkg.EncodeAndWriteResponse(updateResponses, w, lc)
}

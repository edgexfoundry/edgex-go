//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"fmt"
	"math"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/data/application"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/gorilla/mux"
)

type ReadingController struct {
	reader io.DtoReader
	dic    *di.Container
}

// NewReadingController creates and initializes a ReadingController
func NewReadingController(dic *di.Container) *ReadingController {
	return &ReadingController{
		reader: io.NewJsonDtoReader(),
		dic:    dic,
	}
}

func (rc *ReadingController) ReadingTotalCount(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(rc.dic.Get)

	ctx := r.Context()

	// Count readings
	count, err := application.ReadingTotalCount(rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewCountResponse("", "", http.StatusOK, count)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc) // encode and send out the countResponse
}

func (rc *ReadingController) AllReadings(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	// parse URL query string for offset, and limit, and labels
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	readings, totalCount, err := application.AllReadings(offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingsByTimeRange(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	readings, totalCount, err := application.ReadingsByTimeRange(start, end, offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingsByResourceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	vars := mux.Vars(r)
	resourceName := vars[common.ResourceName]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	readings, totalCount, err := application.ReadingsByResourceName(offset, limit, resourceName, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingsByDeviceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	vars := mux.Vars(r)
	name := vars[common.Name]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	readings, totalCount, err := application.ReadingsByDeviceName(offset, limit, name, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingCountByDeviceName(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(rc.dic.Get)

	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	deviceName := vars[common.Name]

	// Count the event by device
	count, err := application.ReadingCountByDeviceName(deviceName, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewCountResponse("", "", http.StatusOK, count)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc) // encode and send out the response
}

// ReadingsByResourceNameAndTimeRange returns readings by resource name and specified time range. Readings are sorted in descending order of origin time.
func (rc *ReadingController) ReadingsByResourceNameAndTimeRange(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	vars := mux.Vars(r)
	resourceName := vars[common.ResourceName]

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	readings, totalCount, err := application.ReadingsByResourceNameAndTimeRange(resourceName, start, end, offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingsByDeviceNameAndResourceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	vars := mux.Vars(r)
	deviceName := vars[common.Name]
	resourceName := vars[common.ResourceName]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	readings, totalCount, err := application.ReadingsByDeviceNameAndResourceName(deviceName, resourceName, offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingsByDeviceNameAndResourceNameAndTimeRange(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	vars := mux.Vars(r)
	deviceName := vars[common.Name]
	resourceName := vars[common.ResourceName]

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	readings, totalCount, err := application.ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName, resourceName, start, end, offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingsByDeviceNameAndResourceNamesAndTimeRange(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	vars := mux.Vars(r)
	deviceName := vars[common.Name]

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	var queryPayload map[string]interface{}
	if r.Body != http.NoBody { //only parse request body when there are contents provided
		err = rc.reader.Read(r.Body, &queryPayload)
		if err != nil {
			utils.WriteErrorResponse(w, ctx, lc, err, "")
			return
		}
	}

	var resourceNames []string
	if val, ok := queryPayload[common.ResourceNames]; ok { //look for
		switch t := val.(type) {
		case []interface{}:
			for _, v := range t {
				if strVal, ok := v.(string); ok {
					resourceNames = append(resourceNames, strVal)
				}
			}
		default:
			err = errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("query criteria [%v] not in expected format", common.ResourceNames), nil)
			utils.WriteErrorResponse(w, ctx, lc, err, "")
			return
		}
	}

	readings, totalCount, err := application.ReadingsByDeviceNameAndResourceNamesAndTimeRange(deviceName, resourceNames, start, end, offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

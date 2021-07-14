//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"math"
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"

	"github.com/edgexfoundry/edgex-go/internal/core/data/application"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
)

type ReadingController struct {
	dic *di.Container
}

// NewReadingController creates and initializes a ReadingController
func NewReadingController(dic *di.Container) *ReadingController {
	return &ReadingController{
		dic: dic,
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
	pkg.Encode(response, w, lc) // encode and send out the countResponse
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
	readings, err := application.AllReadings(offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
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
	readings, err := application.ReadingsByTimeRange(start, end, offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
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
	readings, err := application.ReadingsByResourceName(offset, limit, resourceName, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
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
	readings, err := application.ReadingsByDeviceName(offset, limit, name, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
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
	pkg.Encode(response, w, lc) // encode and send out the response
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
	readings, err := application.ReadingsByResourceNameAndTimeRange(resourceName, start, end, offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
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
	readings, err := application.ReadingsByDeviceNameAndResourceName(deviceName, resourceName, offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
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

	readings, err := application.ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName, resourceName, start, end, offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

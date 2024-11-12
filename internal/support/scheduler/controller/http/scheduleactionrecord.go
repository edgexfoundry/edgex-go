//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"math"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"

	"github.com/edgexfoundry/edgex-go/internal/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/application"
	schedulerContainer "github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
)

type ScheduleActionRecordController struct {
	reader io.DtoReader
	dic    *di.Container
}

// NewScheduleActionRecordController creates and initializes an ScheduleActionRecordController
func NewScheduleActionRecordController(dic *di.Container) *ScheduleActionRecordController {
	return &ScheduleActionRecordController{
		reader: io.NewJsonDtoReader(),
		dic:    dic,
	}
}

// AllScheduleActionRecords handles the GET request of querying all ScheduleActionRecords
func (rc *ScheduleActionRecordController) AllScheduleActionRecords(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	lc := container.LoggingClientFrom(rc.dic.Get)
	config := schedulerContainer.ConfigurationFrom(rc.dic.Get)

	// Parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseQueryStringTimeRangeOffsetLimit(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	records, totalCount, err := application.AllScheduleActionRecords(ctx, start, end, offset, limit, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiScheduleActionRecordsResponse("", "", http.StatusOK, totalCount, records)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// ScheduleActionRecordsByStatus handles the GET request of querying ScheduleActionRecords by status
func (rc *ScheduleActionRecordController) ScheduleActionRecordsByStatus(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	lc := container.LoggingClientFrom(rc.dic.Get)
	config := schedulerContainer.ConfigurationFrom(rc.dic.Get)

	// URL parameters
	status := c.Param(common.Status)

	// Parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseQueryStringTimeRangeOffsetLimit(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	records, totalCount, err := application.ScheduleActionRecordsByStatus(ctx, status, start, end, offset, limit, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiScheduleActionRecordsResponse("", "", http.StatusOK, totalCount, records)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// ScheduleActionRecordsByJobName handles the GET request of querying ScheduleActionRecords by job name
func (rc *ScheduleActionRecordController) ScheduleActionRecordsByJobName(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	lc := container.LoggingClientFrom(rc.dic.Get)
	config := schedulerContainer.ConfigurationFrom(rc.dic.Get)

	// URL parameters
	name := c.Param(common.Name)

	// Parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseQueryStringTimeRangeOffsetLimit(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	records, totalCount, err := application.ScheduleActionRecordsByJobName(ctx, name, start, end, offset, limit, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiScheduleActionRecordsResponse("", "", http.StatusOK, totalCount, records)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// ScheduleActionRecordsByJobNameAndStatus handles the GET request of querying ScheduleActionRecords by job name and status
func (rc *ScheduleActionRecordController) ScheduleActionRecordsByJobNameAndStatus(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	lc := container.LoggingClientFrom(rc.dic.Get)
	config := schedulerContainer.ConfigurationFrom(rc.dic.Get)

	// URL parameters
	name := c.Param(common.Name)
	status := c.Param(common.Status)

	// Parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseQueryStringTimeRangeOffsetLimit(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	records, totalCount, err := application.ScheduleActionRecordsByJobNameAndStatus(ctx, name, status, start, end, offset, limit, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiScheduleActionRecordsResponse("", "", http.StatusOK, totalCount, records)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// LatestScheduleActionRecordsByJobName handles the GET request of querying the latest ScheduleActionRecords of a job by name
func (rc *ScheduleActionRecordController) LatestScheduleActionRecordsByJobName(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	lc := container.LoggingClientFrom(rc.dic.Get)

	// URL parameters
	name := c.Param(common.Name)

	records, totalCount, err := application.LatestScheduleActionRecordsByJobName(ctx, name, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiScheduleActionRecordsResponse("", "", http.StatusOK, totalCount, records)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

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
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/application"
	schedulerContainer "github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
)

type ScheduleJobController struct {
	reader io.DtoReader
	dic    *di.Container
}

// NewScheduleJobController creates and initializes an ScheduleJobController
func NewScheduleJobController(dic *di.Container) *ScheduleJobController {
	return &ScheduleJobController{
		reader: io.NewJsonDtoReader(),
		dic:    dic,
	}
}

// AddScheduleJob handles the POST request of adding new ScheduleJob
func (jc *ScheduleJobController) AddScheduleJob(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(jc.dic.Get)
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.AddScheduleJobRequest
	err := jc.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	var jobs []models.ScheduleJob
	for _, req := range reqDTOs {
		job := dtos.ToScheduleJobModel(req.ScheduleJob)
		jobs = append(jobs, job)
	}

	var addResponses []any
	for i, d := range jobs {
		var response any
		reqId := reqDTOs[i].RequestId
		newId, err := application.AddScheduleJob(ctx, d, jc.dic)
		if err != nil {
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(reqId, err.Message(), err.Code())
		} else {
			response = commonDTO.NewBaseWithIdResponse(reqId, "", http.StatusCreated, newId)
		}
		addResponses = append(addResponses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	return pkg.EncodeAndWriteResponse(addResponses, w, lc)
}

// TriggerScheduleJobByName handles the POST request of triggering ScheduleJob by name
func (jc *ScheduleJobController) TriggerScheduleJobByName(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	lc := container.LoggingClientFrom(jc.dic.Get)

	// URL parameters
	name := c.Param(common.Name)

	err := application.TriggerScheduleJobByName(ctx, name, jc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusAccepted)
	utils.WriteHttpHeader(w, ctx, http.StatusAccepted)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// ScheduleJobByName handles the GET request of querying ScheduleJob by name
func (jc *ScheduleJobController) ScheduleJobByName(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	lc := container.LoggingClientFrom(jc.dic.Get)

	// URL parameters
	name := c.Param(common.Name)

	job, err := application.ScheduleJobByName(ctx, name, jc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewScheduleJobResponse("", "", http.StatusOK, job)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// AllScheduleJobs handles the GET request of querying all ScheduleJobs
func (jc *ScheduleJobController) AllScheduleJobs(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	lc := container.LoggingClientFrom(jc.dic.Get)
	config := schedulerContainer.ConfigurationFrom(jc.dic.Get)

	// parse URL query string for offset and limit
	offset, limit, labels, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	jobs, totalCount, err := application.AllScheduleJobs(ctx, labels, offset, limit, jc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiScheduleJobsResponse("", "", http.StatusOK, totalCount, jobs)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// PatchScheduleJob handles the PATCH request of updating ScheduleJob
func (jc *ScheduleJobController) PatchScheduleJob(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(jc.dic.Get)
	correlationId := correlation.FromContext(ctx)

	var reqDTOs []requestDTO.UpdateScheduleJobRequest
	err := jc.reader.Read(r.Body, &reqDTOs)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	var responses []any
	for _, dto := range reqDTOs {
		var response any
		reqId := dto.RequestId
		err := application.PatchScheduleJob(ctx, dto.ScheduleJob, jc.dic)
		if err != nil {
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(reqId, err.Message(), err.Code())
		} else {
			response = commonDTO.NewBaseResponse(reqId, "", http.StatusOK)
		}
		responses = append(responses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	return pkg.EncodeAndWriteResponse(responses, w, lc)
}

// DeleteScheduleJobByName handles the DELETE request of deleting ScheduleJob by name
func (jc *ScheduleJobController) DeleteScheduleJobByName(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	lc := container.LoggingClientFrom(jc.dic.Get)

	// URL parameters
	name := c.Param(common.Name)

	err := application.DeleteScheduleJobByName(ctx, name, jc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

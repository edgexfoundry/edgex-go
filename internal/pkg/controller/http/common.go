//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// CommonController controller for V2 REST APIs
type CommonController struct {
	dic         *di.Container
	serviceName string
}

// NewCommonController creates and initializes an CommonController
func NewCommonController(dic *di.Container, serviceName string) *CommonController {
	return &CommonController{
		dic:         dic,
		serviceName: serviceName,
	}
}

// Ping handles the request to /ping endpoint. Is used to test if the service is working
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *CommonController) Ping(writer http.ResponseWriter, request *http.Request) {
	response := commonDTO.NewPingResponse(c.serviceName)
	c.sendResponse(writer, request, common.ApiPingRoute, response, http.StatusOK)
}

// Version handles the request to /version endpoint. Is used to request the service's versions
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *CommonController) Version(writer http.ResponseWriter, request *http.Request) {
	response := commonDTO.NewVersionResponse(edgex.Version, c.serviceName)
	c.sendResponse(writer, request, common.ApiVersionRoute, response, http.StatusOK)
}

// Config handles the request to /config endpoint. Is used to request the service's configuration
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *CommonController) Config(writer http.ResponseWriter, request *http.Request) {
	response := commonDTO.NewConfigResponse(container.ConfigurationFrom(c.dic.Get), c.serviceName)
	c.sendResponse(writer, request, common.ApiVersionRoute, response, http.StatusOK)
}

// Metrics handles the request to the /metrics endpoint, memory and cpu utilization stats
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *CommonController) Metrics(writer http.ResponseWriter, request *http.Request) {
	telem := telemetry.NewSystemUsage()
	metrics := commonDTO.Metrics{
		MemAlloc:       telem.Memory.Alloc,
		MemFrees:       telem.Memory.Frees,
		MemLiveObjects: telem.Memory.LiveObjects,
		MemMallocs:     telem.Memory.Mallocs,
		MemSys:         telem.Memory.Sys,
		MemTotalAlloc:  telem.Memory.TotalAlloc,
		CpuBusyAvg:     uint8(telem.CpuBusyAvg),
	}

	response := commonDTO.NewMetricsResponse(metrics, c.serviceName)
	c.sendResponse(writer, request, common.ApiMetricsRoute, response, http.StatusOK)
}

// sendResponse puts together the response packet for the V2 API
func (c *CommonController) sendResponse(
	writer http.ResponseWriter,
	request *http.Request,
	api string,
	response interface{},
	statusCode int) {
	lc := container.LoggingClientFrom(c.dic.Get)

	correlationID := request.Header.Get(common.CorrelationHeader)

	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, common.ContentTypeJSON)
	writer.WriteHeader(statusCode)

	data, err := json.Marshal(response)
	if err != nil {
		lc.Error(fmt.Sprintf("Unable to marshal %s response", api), "error", err.Error(), common.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = writer.Write(data)
	if err != nil {
		lc.Error(fmt.Sprintf("Unable to write %s response", api), "error", err.Error(), common.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *CommonController) sendError(
	writer http.ResponseWriter,
	request *http.Request,
	errKind errors.ErrKind,
	message string,
	err error,
	api string,
	requestID string) {
	lc := container.LoggingClientFrom(c.dic.Get)
	edgeXerr := errors.NewCommonEdgeX(errKind, message, err)
	lc.Error(edgeXerr.Error())
	lc.Debug(edgeXerr.DebugMessages())
	response := commonDTO.NewBaseResponse(requestID, edgeXerr.Message(), edgeXerr.Code())
	c.sendResponse(writer, request, api, response, edgeXerr.Code())
}

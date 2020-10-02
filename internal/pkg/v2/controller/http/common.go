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

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
)

// V2CommonController controller for V2 REST APIs
type V2CommonController struct {
	dic *di.Container
}

// NewV2CommonController creates and initializes an V2CommonController
func NewV2CommonController(dic *di.Container) *V2CommonController {
	return &V2CommonController{
		dic: dic,
	}
}

// Ping handles the request to /ping endpoint. Is used to test if the service is working
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *V2CommonController) Ping(writer http.ResponseWriter, request *http.Request) {
	response := common.NewPingResponse()
	c.sendResponse(writer, request, contractsV2.ApiPingRoute, response, http.StatusOK)
}

// Version handles the request to /version endpoint. Is used to request the service's versions
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *V2CommonController) Version(writer http.ResponseWriter, request *http.Request) {
	response := common.NewVersionResponse(edgex.Version)
	c.sendResponse(writer, request, contractsV2.ApiVersionRoute, response, http.StatusOK)
}

// Config handles the request to /config endpoint. Is used to request the service's configuration
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *V2CommonController) Config(writer http.ResponseWriter, request *http.Request) {
	response := common.NewConfigResponse(container.ConfigurationFrom(c.dic.Get))
	c.sendResponse(writer, request, contractsV2.ApiVersionRoute, response, http.StatusOK)
}

// Metrics handles the request to the /metrics endpoint, memory and cpu utilization stats
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *V2CommonController) Metrics(writer http.ResponseWriter, request *http.Request) {
	telem := telemetry.NewSystemUsage()
	metrics := common.Metrics{
		MemAlloc:       telem.Memory.Alloc,
		MemFrees:       telem.Memory.Frees,
		MemLiveObjects: telem.Memory.LiveObjects,
		MemMallocs:     telem.Memory.Mallocs,
		MemSys:         telem.Memory.Sys,
		MemTotalAlloc:  telem.Memory.TotalAlloc,
		CpuBusyAvg:     uint8(telem.CpuBusyAvg),
	}

	response := common.NewMetricsResponse(metrics)
	c.sendResponse(writer, request, contractsV2.ApiMetricsRoute, response, http.StatusOK)
}

// sendResponse puts together the response packet for the V2 API
func (c *V2CommonController) sendResponse(
	writer http.ResponseWriter,
	request *http.Request,
	api string,
	response interface{},
	statusCode int) {
	lc := container.LoggingClientFrom(c.dic.Get)

	correlationID := request.Header.Get(clients.CorrelationHeader)

	writer.Header().Set(clients.CorrelationHeader, correlationID)
	writer.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	writer.WriteHeader(statusCode)

	data, err := json.Marshal(response)
	if err != nil {
		lc.Error(fmt.Sprintf("Unable to marshal %s response", api), "error", err.Error(), clients.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = writer.Write(data)
	if err != nil {
		lc.Error(fmt.Sprintf("Unable to write %s response", api), "error", err.Error(), clients.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *V2CommonController) sendError(
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
	response := common.NewBaseResponse(requestID, edgeXerr.Message(), edgeXerr.Code())
	c.sendResponse(writer, request, api, response, edgeXerr.Code())
}

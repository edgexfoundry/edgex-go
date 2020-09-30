//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"fmt"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
)

type DeviceServiceController struct {
	reader io.DeviceServiceReader
	dic    *di.Container
}

// NewDeviceServiceController creates and initializes an DeviceServiceController
func NewDeviceServiceController(dic *di.Container) *DeviceServiceController {
	return &DeviceServiceController{
		reader: io.NewDeviceServiceRequestReader(),
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

	addDeviceServiceDTOs, err := dc.reader.ReadAddDeviceServiceRequest(r.Body, &ctx)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		http.Error(w, err.Message(), err.Code())
		return
	}
	deviceServices := requestDTO.AddDeviceServiceReqToDeviceServiceModels(addDeviceServiceDTOs)

	var addResponses []interface{}
	for i, d := range deviceServices {
		newId, err := application.AddDeviceService(d, ctx, dc.dic)
		var addDeviceServiceResponse interface{}
		// get the requestID from addDeviceServiceDTOs
		reqId := addDeviceServiceDTOs[i].RequestID

		if err == nil {
			addDeviceServiceResponse = commonDTO.NewBaseWithIdResponse(
				reqId,
				fmt.Sprintf("Add device service %s successfully", d.Name),
				http.StatusCreated,
				newId)
		} else {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			addDeviceServiceResponse = commonDTO.NewBaseResponse(
				reqId,
				err.Error(),
				err.Code())
		}
		addResponses = append(addResponses, addDeviceServiceResponse)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	// Encode and send the resp body as JSON format
	pkg.Encode(addResponses, w, lc)
}

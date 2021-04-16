//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"

	"github.com/gorilla/mux"
)

type DeviceResourceController struct {
	reader io.DeviceProfileReader
	dic    *di.Container
}

// NewDeviceResourceController creates and initializes an DeviceResourceController
func NewDeviceResourceController(dic *di.Container) *DeviceResourceController {
	return &DeviceResourceController{
		dic: dic,
	}
}

// DeviceResourceByProfileNameAndResourceName query the device resource by profileName and resourceName
func (dc *DeviceResourceController) DeviceResourceByProfileNameAndResourceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	profileName := vars[v2.ProfileName]
	resourceName := vars[v2.ResourceName]

	var response interface{}
	var statusCode int

	resource, err := application.DeviceResourceByProfileNameAndResourceName(profileName, resourceName, dc.dic)
	if err != nil {
		if errors.Kind(err) != errors.KindEntityDoesNotExist {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		}
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = responseDTO.NewDeviceResourceResponse("", "", http.StatusOK, resource)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

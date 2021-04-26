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
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
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

	// URL parameters
	vars := mux.Vars(r)
	profileName := vars[v2.ProfileName]
	resourceName := vars[v2.ResourceName]

	resource, err := application.DeviceResourceByProfileNameAndResourceName(profileName, resourceName, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewDeviceResourceResponse("", "", http.StatusOK, resource)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

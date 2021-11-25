//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"

	"github.com/gorilla/mux"
)

type DeviceResourceController struct {
	dic *di.Container
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
	profileName := vars[common.ProfileName]
	resourceName := vars[common.ResourceName]

	resource, err := application.DeviceResourceByProfileNameAndResourceName(profileName, resourceName, dc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewDeviceResourceResponse("", "", http.StatusOK, resource)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.EncodeAndWriteResponse(response, w, lc)
}

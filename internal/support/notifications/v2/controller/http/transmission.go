//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/application"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"

	"github.com/gorilla/mux"
)

type TransmissionController struct {
	dic *di.Container
}

// NewTransmissionController creates and initializes an TransmissionController
func NewTransmissionController(dic *di.Container) *TransmissionController {
	return &TransmissionController{
		dic: dic,
	}
}

// TransmissionById queries transmission by ID
func (nc *TransmissionController) TransmissionById(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(nc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	id := vars[v2.Id]

	trans, err := application.TransmissionById(id, nc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewTransmissionResponse("", "", http.StatusOK, trans)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

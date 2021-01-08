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
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
)

type ProvisionWatcherController struct {
	reader io.ProvisionWatcherReader
	dic    *di.Container
}

// NewProvisionWatcherController creates and initializes an ProvisionWatcherController
func NewProvisionWatcherController(dic *di.Container) *ProvisionWatcherController {
	return &ProvisionWatcherController{
		reader: io.NewProvisionWatcherRequestReader(),
		dic:    dic,
	}
}

func (pwc *ProvisionWatcherController) AddProvisionWatcher(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(pwc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	addProvisionWatcherDTOs, err := pwc.reader.ReadAddProvisionWatcherRequest(r.Body)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		errResponses := commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		utils.WriteHttpHeader(w, ctx, err.Code())
		// Encode and send the resp body as JSON format
		pkg.Encode(errResponses, w, lc)
		return
	}
	provisionWatchers := requestDTO.AddProvisionWatcherReqToProvisionWatcherModels(addProvisionWatcherDTOs)

	var addResponses []interface{}
	for i, pw := range provisionWatchers {
		newId, err := application.AddProvisionWatcher(pw, ctx, pwc.dic)
		var addProvisionWatcherResponse interface{}
		// get the requestID from addProvisionWatcherDTOs
		reqId := addProvisionWatcherDTOs[i].RequestId

		if err == nil {
			addProvisionWatcherResponse = commonDTO.NewBaseWithIdResponse(
				reqId,
				"",
				http.StatusCreated,
				newId)
		} else {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			addProvisionWatcherResponse = commonDTO.NewBaseResponse(
				reqId,
				err.Error(),
				err.Code())
		}
		addResponses = append(addResponses, addProvisionWatcherResponse)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	// Encode and send the resp body as JSON format
	pkg.Encode(addResponses, w, lc)
}

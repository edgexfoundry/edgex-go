//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"math"
	"net/http"

	metadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/gorilla/mux"
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

func (pwc *ProvisionWatcherController) ProvisionWatcherByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(pwc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	name := vars[contractsV2.Name]

	var response interface{}
	var statusCode int

	provisionWatcher, err := application.ProvisionWatcherByName(name, pwc.dic)
	if err != nil {
		if errors.Kind(err) != errors.KindEntityDoesNotExist {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		}
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = responseDTO.NewProvisionWatcherResponse("", "", http.StatusOK, provisionWatcher)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (pwc *ProvisionWatcherController) ProvisionWatchersByServiceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(pwc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	config := metadataContainer.ConfigurationFrom(pwc.dic.Get)

	vars := mux.Vars(r)
	name := vars[contractsV2.Name]

	var response interface{}
	var statusCode int

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		provisionWatchers, err := application.ProvisionWatchersByServiceName(offset, limit, name, pwc.dic)
		if err != nil {
			if errors.Kind(err) != errors.KindEntityDoesNotExist {
				lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			}
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
			statusCode = err.Code()
		} else {
			response = responseDTO.NewMultiProvisionWatchersResponse("", "", http.StatusOK, provisionWatchers)
			statusCode = http.StatusOK
		}
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (pwc *ProvisionWatcherController) ProvisionWatchersByProfileName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(pwc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	config := metadataContainer.ConfigurationFrom(pwc.dic.Get)

	vars := mux.Vars(r)
	name := vars[contractsV2.Name]

	var response interface{}
	var statusCode int

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		provisionWatchers, err := application.ProvisionWatchersByProfileName(offset, limit, name, pwc.dic)
		if err != nil {
			if errors.Kind(err) != errors.KindEntityDoesNotExist {
				lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			}
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
			statusCode = err.Code()
		} else {
			response = responseDTO.NewMultiProvisionWatchersResponse("", "", http.StatusOK, provisionWatchers)
			statusCode = http.StatusOK
		}
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (pwc *ProvisionWatcherController) AllProvisionWatchers(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(pwc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	config := metadataContainer.ConfigurationFrom(pwc.dic.Get)

	var response interface{}
	var statusCode int

	// parse URL query string for offset, limit
	offset, limit, labels, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		provisionWatchers, err := application.AllProvisionWatchers(offset, limit, labels, pwc.dic)
		if err != nil {
			if errors.Kind(err) != errors.KindEntityDoesNotExist {
				lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			}
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
			statusCode = err.Code()
		} else {
			response = responseDTO.NewMultiProvisionWatchersResponse("", "", http.StatusOK, provisionWatchers)
			statusCode = http.StatusOK
		}
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (pwc *ProvisionWatcherController) DeleteProvisionWatcherByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(pwc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	name := vars[contractsV2.Name]

	var response interface{}
	var statusCode int

	err := application.DeleteProvisionWatcherByName(name, pwc.dic)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = commonDTO.NewBaseResponse(
			"",
			"",
			http.StatusOK)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (pwc *ProvisionWatcherController) PatchProvisionWatcher(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(pwc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	updateProvisionWatcherDTOs, err := pwc.reader.ReadUpdateProvisionWatcherRequest(r.Body)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		errResponses := commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		utils.WriteHttpHeader(w, ctx, err.Code())
		pkg.Encode(errResponses, w, lc)
		return
	}

	var updateResponses []interface{}
	for _, dto := range updateProvisionWatcherDTOs {
		var response interface{}
		reqId := dto.RequestId
		err := application.PatchProvisionWatcher(ctx, dto.ProvisionWatcher, pwc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(
				reqId,
				err.Message(),
				err.Code())
		} else {
			response = commonDTO.NewBaseResponse(
				reqId,
				"",
				http.StatusOK)
		}
		updateResponses = append(updateResponses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(updateResponses, w, lc)
}

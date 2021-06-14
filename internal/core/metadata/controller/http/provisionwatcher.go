//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"math"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/application"
	metadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
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
		utils.WriteErrorResponse(w, ctx, lc, err, "")
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
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
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

	// URL parameters
	vars := mux.Vars(r)
	name := vars[common.Name]

	provisionWatcher, err := application.ProvisionWatcherByName(name, pwc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewProvisionWatcherResponse("", "", http.StatusOK, provisionWatcher)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (pwc *ProvisionWatcherController) ProvisionWatchersByServiceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(pwc.dic.Get)
	ctx := r.Context()
	config := metadataContainer.ConfigurationFrom(pwc.dic.Get)

	vars := mux.Vars(r)
	name := vars[common.Name]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	provisionWatchers, err := application.ProvisionWatchersByServiceName(offset, limit, name, pwc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiProvisionWatchersResponse("", "", http.StatusOK, provisionWatchers)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (pwc *ProvisionWatcherController) ProvisionWatchersByProfileName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(pwc.dic.Get)
	ctx := r.Context()
	config := metadataContainer.ConfigurationFrom(pwc.dic.Get)

	vars := mux.Vars(r)
	name := vars[common.Name]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	provisionWatchers, err := application.ProvisionWatchersByProfileName(offset, limit, name, pwc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiProvisionWatchersResponse("", "", http.StatusOK, provisionWatchers)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (pwc *ProvisionWatcherController) AllProvisionWatchers(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(pwc.dic.Get)
	ctx := r.Context()
	config := metadataContainer.ConfigurationFrom(pwc.dic.Get)

	// parse URL query string for offset, limit
	offset, limit, labels, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	provisionWatchers, err := application.AllProvisionWatchers(offset, limit, labels, pwc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiProvisionWatchersResponse("", "", http.StatusOK, provisionWatchers)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (pwc *ProvisionWatcherController) DeleteProvisionWatcherByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(pwc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	name := vars[common.Name]

	err := application.DeleteProvisionWatcherByName(ctx, name, pwc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
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
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	var updateResponses []interface{}
	for _, dto := range updateProvisionWatcherDTOs {
		var response interface{}
		reqId := dto.RequestId
		err := application.PatchProvisionWatcher(ctx, dto.ProvisionWatcher, pwc.dic)
		if err != nil {
			lc.Error(err.Error(), common.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
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

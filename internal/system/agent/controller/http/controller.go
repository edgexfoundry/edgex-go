//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/application/direct"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/application/direct/config"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/application/executor"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/container"
)

type AgentController struct {
	dic *di.Container
}

func NewAgentController(dic *di.Container) *AgentController {
	return &AgentController{dic: dic}
}

func (c *AgentController) GetHealth(w http.ResponseWriter, r *http.Request) {
	rc := bootstrapContainer.RegistryFrom(c.dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	ctx := r.Context()

	services, err := parseServicesFromQuery(r)
	if err != nil {
		utils.WriteErrorResponse(w, r.Context(), lc, err, "")
		return
	}

	res, err := direct.GetHealth(services, rc)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	utils.WriteHttpHeader(w, r.Context(), http.StatusMultiStatus)
	pkg.EncodeAndWriteResponse(res, w, lc)
}

func (c *AgentController) GetMetrics(w http.ResponseWriter, r *http.Request) {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	ctx := r.Context()

	services, err := parseServicesFromQuery(r)
	if err != nil {
		utils.WriteErrorResponse(w, r.Context(), lc, err, "")
		return
	}

	metricsImpl := container.V2MetricsFrom(c.dic.Get)
	res, err := metricsImpl.Get(ctx, services)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	utils.WriteHttpHeader(w, r.Context(), http.StatusMultiStatus)
	pkg.EncodeAndWriteResponse(res, w, lc)
}

func (c *AgentController) GetConfigs(w http.ResponseWriter, r *http.Request) {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	ctx := r.Context()

	services, err := parseServicesFromQuery(r)
	if err != nil {
		utils.WriteErrorResponse(w, r.Context(), lc, err, "")
		return
	}

	res := config.GetConfigs(ctx, services, c.dic)
	utils.WriteHttpHeader(w, r.Context(), http.StatusMultiStatus)
	pkg.EncodeAndWriteResponse(res, w, lc)
}

func (c *AgentController) PostOperations(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	ctx := r.Context()

	var operations []requests.OperationRequest
	err := json.NewDecoder(r.Body).Decode(&operations)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindContractInvalid, "OperationRequest json decoding failed", err)
		utils.WriteErrorResponse(w, ctx, lc, edgexErr, "")
		return
	}

	configuration := container.ConfigurationFrom(c.dic.Get)
	operator := executor.NewOperation(executor.CommandExecutor, configuration.ExecutorPath, lc)
	res, edgexErr := operator.Do(ctx, operations)
	if edgexErr != nil {
		utils.WriteErrorResponse(w, ctx, lc, edgexErr, "")
		return
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.EncodeAndWriteResponse(res, w, lc)
}

func parseServicesFromQuery(r *http.Request) ([]string, errors.EdgeX) {
	queryString := r.URL.Query()[common.Services]
	if queryString == nil || (len(queryString) == 1 && queryString[0] == "") {
		err := errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to parse query", nil)
		return nil, err
	}

	services := strings.Split(queryString[0], common.CommaSeparator)
	return services, nil
}

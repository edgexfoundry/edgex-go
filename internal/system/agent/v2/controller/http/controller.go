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
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/container"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/v2/application/direct"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/v2/application/direct/config"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/v2/application/executor"
	v2Container "github.com/edgexfoundry/edgex-go/internal/system/agent/v2/container"
)

type AgentController struct {
	dic *di.Container
}

func NewAgentController(dic *di.Container) *AgentController {
	return &AgentController{dic: dic}
}

func (c *AgentController) GetHealth(w http.ResponseWriter, r *http.Request) {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)

	vars := mux.Vars(r)
	services := strings.Split(vars[v2.Services], v2.CommaSeparator)

	health := direct.GetHealth(services, bootstrapContainer.RegistryFrom(c.dic.Get))
	res := responses.NewHealthResponse("", "", http.StatusOK, health)
	pkg.Encode(res, w, lc)
}

func (c *AgentController) GetMetrics(w http.ResponseWriter, r *http.Request) {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	ctx := r.Context()
	vars := mux.Vars(r)
	services := strings.Split(vars[v2.Services], v2.CommaSeparator)

	metricsImpl := v2Container.V2MetricsFrom(c.dic.Get)
	res, err := metricsImpl.Get(ctx, services)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(res, w, lc)
}

func (c *AgentController) GetConfigs(w http.ResponseWriter, r *http.Request) {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	ctx := r.Context()

	vars := mux.Vars(r)
	services := strings.Split(vars[v2.Services], v2.CommaSeparator)

	configs, err := config.GetConfigs(ctx, services, c.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(configs, w, lc)
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
	pkg.Encode(res, w, lc)
}

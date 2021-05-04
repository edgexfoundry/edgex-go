//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"
	"strings"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/v2/application"
)

type AgentController struct {
	dic *di.Container
}

func NewAgentController(dic *di.Container) *AgentController {
	return &AgentController{dic: dic}
}

func (c *AgentController) GetHealth(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(c.dic.Get)

	vars := mux.Vars(r)
	services := strings.Split(vars["services"], ",")

	health := application.GetHealth(services, container.RegistryFrom(c.dic.Get))
	res := responses.NewHealthResponse("", "", http.StatusOK, health)
	pkg.Encode(res, w, lc)
}

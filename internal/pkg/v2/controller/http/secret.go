//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
)

// AddSecret handles the request to the /secret endpoint. Is used to add EdgeX Service exclusive secret to the Secret Store
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *V2CommonController) AddSecret(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		_ = request.Body.Close()
	}()

	secretRequest := common.SecretRequest{}
	err := json.NewDecoder(request.Body).Decode(&secretRequest)
	if err != nil {
		c.sendError(writer, request, errors.KindContractInvalid, "JSON decode failed", err, v2.ApiSecretRoute, "")
		return
	}

	err = application.AddSecret(c.dic, secretRequest)
	if err != nil {
		c.sendError(writer, request, errors.Kind(err), err.Error(), err, v2.ApiSecretRoute, secretRequest.RequestId)
		return
	}

	response := common.NewBaseResponse(secretRequest.RequestId, "", http.StatusCreated)
	c.sendResponse(writer, request, v2.ApiSecretRoute, response, http.StatusCreated)
}

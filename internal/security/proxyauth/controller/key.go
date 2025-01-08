//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/edgex-go/internal/security/proxyauth/application"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"

	"github.com/labstack/echo/v4"
)

func (a *AuthController) AddKey(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := bootstrapContainer.LoggingClientFrom(a.dic.Get)
	ctx := r.Context()

	var req requests.AddKeyDataRequest
	err := a.reader.Read(r.Body, &req)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	var response any
	reqId := req.RequestId

	err = application.AddKey(a.dic, dtos.ToKeyDataModel(req.KeyData))
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response = commonDTO.NewBaseResponse(
		reqId,
		"",
		http.StatusCreated)
	utils.WriteHttpHeader(w, ctx, http.StatusCreated)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (a *AuthController) VerificationKeyByIssuer(c echo.Context) error {
	lc := bootstrapContainer.LoggingClientFrom(a.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	issuer := c.Param(common.Issuer)

	keyData, err := application.VerificationKeyByIssuer(a.dic, issuer)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responses.NewKeyDataResponse("", "", http.StatusOK, keyData)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

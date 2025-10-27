//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	goio "io"
	"net/http"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/edgexfoundry/edgex-go/internal/core/keeper/application"
	"github.com/edgexfoundry/edgex-go/internal/core/keeper/constants"
	httpUtils "github.com/edgexfoundry/edgex-go/internal/core/keeper/utils"
	"github.com/edgexfoundry/edgex-go/internal/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"

	"github.com/labstack/echo/v4"
)

type RegistryController struct {
	reader io.DtoReader
	dic    *di.Container
}

func NewRegistryController(dic *di.Container) *RegistryController {
	return &RegistryController{
		reader: io.NewJsonDtoReader(),
		dic:    dic,
	}
}

func (rc *RegistryController) Register(c echo.Context) error {
	r := c.Request()
	w := c.Response()

	if r.Body != nil {
		defer func(Body goio.ReadCloser) {
			err := Body.Close()
			if err != nil {
				bootstrapContainer.LoggingClientFrom(rc.dic.Get).Warnf("error occured while closing the request body: %s", err.Error())
			}
		}(r.Body)
	}

	lc := bootstrapContainer.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()

	var reqDTO requests.AddRegistrationRequest
	edgexErr := rc.reader.Read(r.Body, &reqDTO)
	if edgexErr != nil {
		return utils.WriteErrorResponse(w, ctx, lc, edgexErr, "")
	}

	err := reqDTO.Validate()
	if err != nil {
		edgexErr = errors.NewCommonEdgeX(errors.KindContractInvalid, "bad AddRegistrationRequest type", err)
		return utils.WriteErrorResponse(w, ctx, lc, edgexErr, "")
	}

	registry := dtos.ToRegistrationModel(reqDTO.Registration)
	edgexErr = application.AddRegistration(registry, rc.dic)
	if edgexErr != nil {
		return utils.WriteErrorResponse(w, ctx, lc, edgexErr, "")
	}

	utils.WriteHttpHeader(w, ctx, http.StatusCreated)
	response := commonDTO.BaseWithServiceNameResponse{
		BaseResponse: commonDTO.NewBaseResponse("", "", http.StatusCreated),
		ServiceName:  registry.ServiceId,
	}
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *RegistryController) UpdateRegister(c echo.Context) error {
	r := c.Request()
	w := c.Response()

	if r.Body != nil {
		defer func(Body goio.ReadCloser) {
			err := Body.Close()
			if err != nil {
				bootstrapContainer.LoggingClientFrom(rc.dic.Get).Warnf("error occured while closing the request body: %s", err.Error())
			}
		}(r.Body)
	}

	lc := bootstrapContainer.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()

	var reqDTO requests.AddRegistrationRequest
	edgexErr := rc.reader.Read(r.Body, &reqDTO)
	if edgexErr != nil {
		return utils.WriteErrorResponse(w, ctx, lc, edgexErr, "")
	}

	err := reqDTO.Validate()
	if err != nil {
		edgexErr = errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid registration request", err)
		return utils.WriteErrorResponse(w, ctx, lc, edgexErr, "")
	}

	registry := dtos.ToRegistrationModel(reqDTO.Registration)
	edgexErr = application.UpdateRegistration(registry, rc.dic)
	if edgexErr != nil {
		return utils.WriteErrorResponse(w, ctx, lc, edgexErr, "")
	}

	utils.WriteHttpHeader(w, ctx, http.StatusNoContent)
	return nil
}

func (rc *RegistryController) Deregister(c echo.Context) error {
	r := c.Request()
	w := c.Response()

	lc := bootstrapContainer.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()

	// URL parameters
	id := c.Param(constants.ServiceId)

	err := application.DeleteRegistration(id, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	utils.WriteHttpHeader(w, ctx, http.StatusNoContent)
	return nil
}

func (rc *RegistryController) Registrations(c echo.Context) error {
	r := c.Request()
	w := c.Response()

	lc := bootstrapContainer.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()

	deregistered, err := httpUtils.ParseQueryStringToBool(r, constants.Deregistered)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	dtos, err := application.Registrations(rc.dic, deregistered)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responses.NewMultiRegistrationsResponse("", "", http.StatusOK, uint32(len(dtos)), dtos)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *RegistryController) RegistrationByServiceId(c echo.Context) error {
	r := c.Request()
	w := c.Response()

	lc := bootstrapContainer.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()

	// URL parameters
	id := c.Param(constants.ServiceId)

	dto, err := application.RegistrationByServiceId(id, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responses.NewRegistrationResponse("", "", http.StatusOK, dto)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

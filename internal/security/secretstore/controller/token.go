//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/container"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/tokenprovider"
	secretUtils "github.com/edgexfoundry/edgex-go/internal/security/secretstore/utils"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/fileioperformer"

	"github.com/labstack/echo/v4"
)

type TokenController struct {
	dic *di.Container
}

func NewTokenController(dic *di.Container) *TokenController {
	return &TokenController{
		dic: dic,
	}
}

func (a *TokenController) RegenToken(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	// URL parameters
	entityId := c.Param("entityId")

	lc := bootstrapContainer.LoggingClientFrom(a.dic.Get)
	tokenLoader := bootstrapContainer.AuthTokenLoaderFrom(a.dic.Get)
	if tokenLoader == nil {
		tokenLoader = authtokenloader.NewAuthTokenLoader(fileioperformer.NewDefaultFileIoPerformer())
	}
	configuration := container.ConfigurationFrom(a.dic.Get)
	secretStoreConfig := configuration.SecretStore
	ctx := r.Context()

	tokenProvider := tokenprovider.NewTokenProvider(ctx, lc, secretUtils.NewDefaultExecRunner())
	if secretStoreConfig.TokenProvider != "" {
		if err := tokenProvider.SetConfiguration(secretStoreConfig); err != nil {
			lc.Errorf("failed to configure token provider: %v", err)
			return err
		}

		err := tokenProvider.LaunchRegenToken(entityId)
		if err != nil {
			lc.Errorf("failed to call LaunchRefreshToken from token provider: %v", err)
			return err
		}
	} else {
		lc.Info("no token provider configured")
	}

	response := commonDTO.NewBaseResponse(
		"",
		"",
		http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

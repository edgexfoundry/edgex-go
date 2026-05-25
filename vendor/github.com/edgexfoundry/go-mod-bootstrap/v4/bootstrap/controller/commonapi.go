//
// Copyright (C) 2023-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
)

// CommonController controller for common REST APIs
type CommonController struct {
	dic         *di.Container
	serviceName string
	router      *echo.Echo
	version     version
	config      config
	lc          logger.LoggingClient
}

type version struct {
	serviceVersion string
	sdkVersion     string
}

type config struct {
	configuration interfaces.Configuration
	customConfig  interfaces.UpdatableConfig
}

func NewCommonController(dic *di.Container, r *echo.Echo, serviceName string, serviceVersion string) *CommonController {
	lc := container.LoggingClientFrom(dic.Get)
	authenticationHook := handlers.AutoConfigAuthenticationFunc(dic)
	configuration := container.ConfigurationFrom(dic.Get)
	c := CommonController{
		dic:         dic,
		serviceName: serviceName,
		router:      r,
		lc:          lc,
		version: version{
			serviceVersion: serviceVersion,
		},
		config: config{
			configuration: configuration,
		},
	}
	r.GET(common.ApiPingRoute, c.Ping) // Health check is always unauthenticated
	r.GET(common.ApiVersionRoute, c.Version, authenticationHook)
	r.GET(common.ApiConfigRoute, c.Config, authenticationHook)
	r.POST(common.ApiSecretRoute, c.AddSecret, authenticationHook)

	return &c
}

// SetSDKVersion sets the service's skd version
func (c *CommonController) SetSDKVersion(sdkVersion string) {
	c.version.sdkVersion = sdkVersion
}

// SetCustomConfigInfo sets the custom configuration, which is used to include the service's custom config in the /config endpoint response.
func (c *CommonController) SetCustomConfigInfo(customConfig interfaces.UpdatableConfig) {
	c.config.customConfig = customConfig
}

// Ping handles the request to /ping endpoint. Is used to test if the service is working
// It returns a response as specified by the API swagger in the openapi directory
func (c *CommonController) Ping(e echo.Context) error {
	request := e.Request()
	writer := e.Response()
	response := commonDTO.NewPingResponse(c.serviceName)

	return utils.SendJsonResp(c.lc, writer, request, response, http.StatusOK)
}

// Version handles the request to /version endpoint. Is used to request the service's versions
// It returns a response as specified by the API swagger in the openapi directory
func (c *CommonController) Version(e echo.Context) error {
	request := e.Request()
	writer := e.Response()
	var response interface{}
	if c.version.sdkVersion != "" {
		response = commonDTO.NewVersionSdkResponse(c.version.serviceVersion, c.version.sdkVersion, c.serviceName)
	} else {
		response = commonDTO.NewVersionResponse(c.version.serviceVersion, c.serviceName)
	}
	return utils.SendJsonResp(c.lc, writer, request, response, http.StatusOK)
}

// Config handles the request to /config endpoint. Is used to request the service's configuration
// It returns a response as specified by the swagger in openapi/common
func (c *CommonController) Config(e echo.Context) error {
	request := e.Request()
	writer := e.Response()
	var fullConfig interface{}
	m := make(map[string]any)
	err := mapstructure.Decode(c.config.configuration, &m)
	if err != nil {
		c.lc.Errorf("%v", err.Error())
		return utils.SendJsonErrResp(c.lc, writer, request, errors.KindServerError, "config can not convert to map", err, "")
	}
	if c.config.customConfig != nil {
		m["CustomConfiguration"] = c.config.customConfig
	}
	fullConfig = m

	response := commonDTO.NewConfigResponse(fullConfig, c.serviceName)
	return utils.SendJsonResp(c.lc, writer, request, response, http.StatusOK)
}

// AddSecret handles the request to the /secret endpoint. Is used to add EdgeX Service exclusive secret to the Secret Store
// It returns a response as specified by the API swagger in the openapi directory
func (c *CommonController) AddSecret(e echo.Context) error {
	request := e.Request()
	writer := e.Response()

	defer func() {
		_ = request.Body.Close()
	}()

	secretRequest := commonDTO.SecretRequest{}
	err := json.NewDecoder(request.Body).Decode(&secretRequest)
	if err != nil {
		c.lc.Errorf("%v", err.Error())
		return utils.SendJsonErrResp(c.lc, writer, request, errors.KindContractInvalid, "JSON decode failed", err, "")
	}

	err = addSecret(c.dic, secretRequest)
	if err != nil {
		return utils.SendJsonErrResp(c.lc, writer, request, errors.Kind(err), err.Error(), err, secretRequest.RequestId)
	}

	response := commonDTO.NewBaseResponse(secretRequest.RequestId, "", http.StatusCreated)

	return utils.SendJsonResp(c.lc, writer, request, response, http.StatusCreated)
}

// addSecret adds EdgeX Service exclusive secret to the Secret Store
func addSecret(dic *di.Container, request commonDTO.SecretRequest) errors.EdgeX {
	secretName, secret := prepareSecret(request)

	secretProvider := container.SecretProviderFrom(dic.Get)
	if secretProvider == nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "secret provider is missing. Make sure it is specified to be used in bootstrap.Run()", nil)
	}

	if err := secretProvider.StoreSecret(secretName, secret); err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "adding secret failed", err)
	}
	return nil
}

func prepareSecret(request commonDTO.SecretRequest) (string, map[string]string) {
	var secretsKV = make(map[string]string)
	for _, secret := range request.SecretData {
		secretsKV[secret.Key] = secret.Value
	}

	secretName := strings.TrimSpace(request.SecretName)

	return secretName, secretsKV
}

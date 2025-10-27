//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"io"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/keeper/application"
	"github.com/edgexfoundry/edgex-go/internal/core/keeper/constants"
	kpContrUtils "github.com/edgexfoundry/edgex-go/internal/core/keeper/utils"
	edgexIO "github.com/edgexfoundry/edgex-go/internal/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/labstack/echo/v4"
)

type KVController struct {
	reader edgexIO.DtoReader
	dic    *di.Container
}

// NewKVController creates and initializes a KVController
func NewKVController(dic *di.Container) *KVController {
	return &KVController{
		reader: edgexIO.NewJsonDtoReader(),
		dic:    dic,
	}
}

func (rc *KVController) Keys(c echo.Context) error {
	r := c.Request()
	w := c.Response()

	lc := bootstrapContainer.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()

	// URL parameters
	key := c.Param(constants.Key)

	// parse URL query string for keyOnly and plaintext
	keysOnly, isRaw, err := kpContrUtils.ParseGetKeyRequestQueryString(r)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	resp, err := application.Keys(key, keysOnly, isRaw, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responses.NewMultiKVResponse("", "", http.StatusOK, resp)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *KVController) AddKeys(c echo.Context) error {
	r := c.Request()
	w := c.Response()

	if r.Body != nil {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				bootstrapContainer.LoggingClientFrom(rc.dic.Get).Warnf("error occured while closing the request body: %s", err.Error())
			}
		}(r.Body)
	}

	lc := bootstrapContainer.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()

	// URL parameters
	key := c.Param(constants.Key)

	// parse URL query string for flatten
	isFlatten, err := kpContrUtils.ParseAddKeyRequestQueryString(r)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	var reqDTO requests.UpdateKeysRequest
	err = rc.reader.Read(r.Body, &reqDTO)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	err = reqDTO.Validate()
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	kvModel := requests.UpdateKeysReqToKVModels(reqDTO, key)
	keys, err := application.AddKeys(kvModel, isFlatten, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	// publish the key change event
	go application.PublishKeyChange(kvModel, key, ctx, rc.dic)

	response := responses.NewKeysResponse("", "", http.StatusOK, keys)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *KVController) DeleteKeys(c echo.Context) error {
	r := c.Request()
	w := c.Response()

	lc := bootstrapContainer.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()

	// URL parameters
	key := c.Param(constants.Key)

	// parse URL query string for prefixMatch
	prefixMatch, err := kpContrUtils.ParseDeleteKeyRequestQueryString(r)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	resp, err := application.DeleteKeys(key, prefixMatch, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	// publish the key change event
	go application.PublishKeyChange(models.KVS{Key: key}, key, ctx, rc.dic)

	response := responses.NewKeysResponse("", "", http.StatusOK, resp)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
)

type UnitOfMeasureController struct {
	dic *di.Container
}

func NewUnitOfMeasureController(dic *di.Container) *UnitOfMeasureController {
	return &UnitOfMeasureController{
		dic: dic,
	}
}

func (uc *UnitOfMeasureController) UnitsOfMeasure(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := container.UnitsOfMeasureFrom(uc.dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(uc.dic.Get)

	response := responses.NewUnitsOfMeasureResponse("", "", http.StatusOK, u)

	utils.WriteHttpHeader(w, ctx, http.StatusOK)

	switch r.Header.Get(common.Accept) {
	case common.ContentTypeTOML:
		pkg.EncodeAndWriteTomlResponse(u, w, lc)
	default:
		pkg.EncodeAndWriteResponse(response, w, lc)
	}
}

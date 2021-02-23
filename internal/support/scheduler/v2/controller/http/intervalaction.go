//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/io"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
)

type IntervalActionController struct {
	reader io.IntervalActionReader
	dic    *di.Container
}

// NewIntervalActionController creates and initializes an IntervalActionController
func NewIntervalActionController(dic *di.Container) *IntervalActionController {
	return &IntervalActionController{
		reader: io.NewIntervalActionRequestReader(),
		dic:    dic,
	}
}

func (dc *IntervalActionController) AddIntervalAction(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	actionDTOs, err := dc.reader.ReadAddIntervalActionRequest(r.Body)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		errResponses := commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		utils.WriteHttpHeader(w, ctx, err.Code())
		pkg.Encode(errResponses, w, lc)
		return
	}
	actions := requestDTO.AddIntervalActionReqToIntervalActionModels(actionDTOs)

	var addResponses []interface{}
	for i, action := range actions {
		var response interface{}
		reqId := actionDTOs[i].RequestId
		newId, err := application.AddIntervalAction(action, ctx, dc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(
				reqId,
				err.Message(),
				err.Code())
		} else {
			response = commonDTO.NewBaseWithIdResponse(
				reqId,
				"",
				http.StatusCreated,
				newId)
		}
		addResponses = append(addResponses, response)

		// TODO Add the new IntervalAction into scheduler queue
		//err = scClient.AddIntervalActionToQueue(intervalAction)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(addResponses, w, lc)
}

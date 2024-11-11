//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
)

const (
	exampleUUID         = "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"
	testCorrelationID   = ""
	testScheduleJobName = "jobName"
	testStatus          = "SUCCEEDED"
	testTimestamp       = 1723642440000
)

var (
	testScheduleJobLabels = []string{"label"}
	testScheduleDef       = dtos.ScheduleDef{
		Type: common.DefInterval,
		IntervalScheduleDef: dtos.IntervalScheduleDef{
			Interval: "10m",
		},
	}
	testScheduleAction = dtos.ScheduleAction{
		Type:        common.ActionEdgeXMessageBus,
		ContentType: common.ContentTypeJSON,
		Payload:     nil,
		EdgeXMessageBusAction: dtos.EdgeXMessageBusAction{
			Topic: "testTopic",
		},
	}
	testScheduleActions = []dtos.ScheduleAction{
		testScheduleAction,
		{
			Type:        common.ActionREST,
			ContentType: common.ContentTypeJSON,
			Payload:     nil,
			RESTAction: dtos.RESTAction{
				Address: "testAddress",
				Method:  http.MethodGet,
			},
		},
	}
)

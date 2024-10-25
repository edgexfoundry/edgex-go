//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cronscheduler

import (
	"github.com/labstack/echo/v4"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/edgexfoundry/edgex-go"
	schedulerController "github.com/edgexfoundry/edgex-go/internal/support/cronscheduler/controller/http"
)

func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	lc := container.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderExtFrom(dic.Get)
	authenticationHook := handlers.AutoConfigAuthenticationFunc(secretProvider, lc)

	// Common
	_ = controller.NewCommonController(dic, r, serviceName, edgex.Version)

	// ScheduleJob
	jc := schedulerController.NewScheduleJobController(dic)
	r.POST(common.ApiScheduleJobRoute, jc.AddScheduleJob, authenticationHook)
	r.POST(common.ApiTriggerScheduleJobByNameEchoRoute, jc.TriggerScheduleJobByName, authenticationHook)
	r.PATCH(common.ApiScheduleJobRoute, jc.PatchScheduleJob, authenticationHook)
	r.GET(common.ApiAllScheduleJobRoute, jc.AllScheduleJobs, authenticationHook)
	r.GET(common.ApiScheduleJobByNameEchoRoute, jc.ScheduleJobByName, authenticationHook)
	r.DELETE(common.ApiScheduleJobByNameEchoRoute, jc.DeleteScheduleJobByName, authenticationHook)

	// ScheduleActionRecord
	rc := schedulerController.NewScheduleActionRecordController(dic)
	r.GET(common.ApiAllScheduleActionRecordRoute, rc.AllScheduleActionRecords, authenticationHook)
	r.GET(common.ApiScheduleActionRecordRouteByStatusEchoRoute, rc.ScheduleActionRecordsByStatus, authenticationHook)
	r.GET(common.ApiScheduleActionRecordRouteByJobNameEchoRoute, rc.ScheduleActionRecordsByJobName, authenticationHook)
	r.GET(common.ApiScheduleActionRecordRouteByJobNameAndStatusEchoRoute, rc.ScheduleActionRecordsByJobNameAndStatus, authenticationHook)
	r.GET(common.ApiLatestScheduleActionRecordByJobNameEchoRoute, rc.LatestScheduleActionRecordsByJobName, authenticationHook)
}

//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"github.com/labstack/echo/v4"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/edgexfoundry/edgex-go"
	schedulerController "github.com/edgexfoundry/edgex-go/internal/support/scheduler/controller/http"
)

func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	authenticationHook := handlers.AutoConfigAuthenticationFunc(dic)

	// Common
	_ = controller.NewCommonController(dic, r, serviceName, edgex.Version)

	// ScheduleJob
	jc := schedulerController.NewScheduleJobController(dic)
	r.POST(common.ApiScheduleJobRoute, jc.AddScheduleJob, authenticationHook)
	r.POST(common.ApiTriggerScheduleJobByNameRoute, jc.TriggerScheduleJobByName, authenticationHook)
	r.PATCH(common.ApiScheduleJobRoute, jc.PatchScheduleJob, authenticationHook)
	r.GET(common.ApiAllScheduleJobRoute, jc.AllScheduleJobs, authenticationHook)
	r.GET(common.ApiScheduleJobByNameRoute, jc.ScheduleJobByName, authenticationHook)
	r.DELETE(common.ApiScheduleJobByNameRoute, jc.DeleteScheduleJobByName, authenticationHook)

	// ScheduleActionRecord
	rc := schedulerController.NewScheduleActionRecordController(dic)
	r.GET(common.ApiAllScheduleActionRecordRoute, rc.AllScheduleActionRecords, authenticationHook)
	r.GET(common.ApiScheduleActionRecordRouteByStatusRoute, rc.ScheduleActionRecordsByStatus, authenticationHook)
	r.GET(common.ApiScheduleActionRecordRouteByJobNameRoute, rc.ScheduleActionRecordsByJobName, authenticationHook)
	r.GET(common.ApiScheduleActionRecordRouteByJobNameAndStatusRoute, rc.ScheduleActionRecordsByJobNameAndStatus, authenticationHook)
	r.GET(common.ApiLatestScheduleActionRecordByJobNameRoute, rc.LatestScheduleActionRecordsByJobName, authenticationHook)
}

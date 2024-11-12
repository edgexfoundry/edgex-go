//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"

	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

func ToGocronJobDef(def models.ScheduleDef) (gocron.JobDefinition, errors.EdgeX) {
	var definition gocron.JobDefinition
	switch def.GetBaseScheduleDef().Type {
	case common.DefCron:
		cronDef, ok := def.(models.CronScheduleDef)
		if !ok {
			return definition, errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast ScheduleDefinition to CronScheduleDef", nil)
		}

		// An optional 6th field can be used at the beginning if withSeconds is set to true: `* * * * * *`
		definition = gocron.CronJob(cronDef.Crontab, true)
	case common.DefInterval:
		intervalDef, ok := def.(models.IntervalScheduleDef)
		if !ok {
			return definition, errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast ScheduleDefinition to IntervalScheduleDef", nil)
		}

		duration, err := time.ParseDuration(intervalDef.Interval)
		if err != nil {
			return definition, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("fail to parse interval string %s to a duration time value", intervalDef.Interval), err)
		}

		definition = gocron.DurationJob(duration)
	default:
		return definition, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unsupported schedule definition type: %s", def.GetBaseScheduleDef().Type), nil)
	}

	return definition, nil
}

func ToGocronTask(lc logger.LoggingClient, dic *di.Container, secretProvider bootstrapInterfaces.SecretProviderExt, action models.ScheduleAction) (gocron.Task, errors.EdgeX) {
	var task gocron.Task
	switch action.GetBaseScheduleAction().Type {
	case common.ActionEdgeXMessageBus:
		edgeXMessageBusAction, ok := action.(models.EdgeXMessageBusAction)
		if !ok {
			return task, errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast ScheduleAction to EdgeXMessageBusAction", nil)
		}
		task = edgeXMessageBusActionTask(lc, dic, edgeXMessageBusAction)
	case common.ActionREST:
		restAction, ok := action.(models.RESTAction)
		if !ok {
			return task, errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast ScheduleAction to RESTAction", nil)
		}
		task = restActionTask(lc, secretProvider, restAction)
	case common.ActionDeviceControl:
		deviceControlAction, ok := action.(models.DeviceControlAction)
		if !ok {
			return task, errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast ScheduleAction to DeviceControlAction", nil)
		}
		task = deviceControlActionTask(lc, dic, deviceControlAction)
	default:
		return task, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unsupported schedule action type: %s", action.GetBaseScheduleAction().Type), nil)
	}

	return task, nil
}

func edgeXMessageBusActionTask(lc logger.LoggingClient, dic *di.Container, action models.EdgeXMessageBusAction) gocron.Task {
	return gocron.NewTask(func() errors.EdgeX {
		if err := publishEdgeXMessageBus(dic, action); err != nil {
			lc.Debugf("Failed to execute the EdgeX message bus action: %v", err)
			return err
		}
		lc.Debugf("EdgeX message bus action was executed successfully")
		return nil
	})
}

func restActionTask(lc logger.LoggingClient, secretProvider bootstrapInterfaces.SecretProviderExt, action models.RESTAction) gocron.Task {
	var injector interfaces.AuthenticationInjector
	if action.InjectEdgeXAuth {
		injector = secret.NewJWTSecretProvider(secretProvider)
	}

	return gocron.NewTask(func() errors.EdgeX {
		resp, err := sendRESTRequest(lc, action, injector)
		if err != nil {
			lc.Debugf("Failed to execute the rest action: %v", err)
			return err
		}
		lc.Debugf("REST action was executed successfully, response: %s", resp)
		return nil
	})
}

func deviceControlActionTask(lc logger.LoggingClient, dic *di.Container, action models.DeviceControlAction) gocron.Task {
	return gocron.NewTask(func() errors.EdgeX {
		resp, err := issueSetCommand(dic, action)
		if err != nil {
			lc.Debugf("Failed to execute the device control action: %v", err)
			return err
		}
		lc.Debugf("DeviceControl action was executed successfully, response: %s", resp)
		return nil
	})
}

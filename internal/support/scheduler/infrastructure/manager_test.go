//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package infrastructure

import (
	"github.com/stretchr/testify/assert"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
)

const (
	testUUID          = "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"
	testName          = "testName"
	testLabel         = "testLabel"
	testCorrelationID = ""
)

var (
	testIntervalScheduleDef = models.IntervalScheduleDef{
		BaseScheduleDef: models.BaseScheduleDef{
			Type: common.DefInterval,
		},
		Interval: "10m",
	}
	testCronScheduleDef = models.CronScheduleDef{
		BaseScheduleDef: models.BaseScheduleDef{
			Type: common.DefInterval,
		},
		Crontab: "* * * * *",
	}
	testRestScheduleAction = models.RESTAction{
		BaseScheduleAction: models.BaseScheduleAction{
			Id:          testUUID,
			Type:        common.ActionREST,
			ContentType: common.ContentTypeJSON,
			Payload:     []byte{},
		},
		Address: "testAddress",
	}
	testEdgeXMessageBusScheduleAction = models.EdgeXMessageBusAction{
		BaseScheduleAction: models.BaseScheduleAction{
			Id:          testUUID,
			Type:        common.ActionEdgeXMessageBus,
			ContentType: common.ContentTypeJSON,
			Payload:     []byte{},
		},
		Topic: "testTopic",
	}
	testDeviceControlScheduleAction = models.DeviceControlAction{
		BaseScheduleAction: models.BaseScheduleAction{
			Id:          testUUID,
			Type:        common.ActionDeviceControl,
			ContentType: common.ContentTypeJSON,
			Payload:     []byte{},
		},
		DeviceName: "testDevice",
		SourceName: "testSource",
	}
)

func mockDic() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) any {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					LogLevel: "DEBUG",
				},
				Service: bootstrapConfig.ServiceInfo{
					MaxResultCount: 20,
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) any {
			return logger.NewMockClient()
		},
	})
}

func validScheduleJob() models.ScheduleJob {
	return models.ScheduleJob{
		Id:         testUUID,
		Name:       testName,
		Definition: testIntervalScheduleDef,
		Actions:    []models.ScheduleAction{testEdgeXMessageBusScheduleAction},
		AdminState: models.Unlocked,
		Labels:     []string{testLabel},
	}
}

func TestValidateUpdatingScheduleJob(t *testing.T) {
	dic := mockDic()
	mockManager := NewManager(dic)

	// Add a valid schedule job first
	validJob := validScheduleJob()
	err := mockManager.AddScheduleJob(validJob, testCorrelationID)
	assert.NoError(t, err)

	emptyInterval := testIntervalScheduleDef
	emptyInterval.Interval = ""
	invalidInterval := testIntervalScheduleDef
	invalidInterval.Interval = "abc"

	emptyCrontab := testCronScheduleDef
	emptyCrontab.Crontab = ""
	invalidCrontab := testCronScheduleDef
	invalidCrontab.Crontab = "abc"
	wrongFormatCrontab := testCronScheduleDef
	wrongFormatCrontab.Crontab = "1 2 3 4 5 6 7"

	invalidRestAction := testRestScheduleAction
	invalidRestAction.Type = "invalid"

	invalidEdgeXMessageBusAction := testEdgeXMessageBusScheduleAction
	invalidEdgeXMessageBusAction.Type = "invalid"

	invalidDeviceControlAction := testDeviceControlScheduleAction
	invalidDeviceControlAction.Type = "invalid"

	tests := []struct {
		name          string
		job           models.ScheduleJob
		expectedError bool
	}{
		{"valid schedule job", validJob, false},
		{"empty name", models.ScheduleJob{}, true},
		{"empty id", models.ScheduleJob{}, true},
		{"empty interval", models.ScheduleJob{Name: testName, Definition: emptyInterval, Actions: []models.ScheduleAction{testRestScheduleAction}}, true},
		{"invalid interval", models.ScheduleJob{Name: testName, Definition: invalidInterval, Actions: []models.ScheduleAction{testRestScheduleAction}}, true},
		{"empty crontab", models.ScheduleJob{Name: testName, Definition: emptyCrontab, Actions: []models.ScheduleAction{testRestScheduleAction}}, true},
		{"invalid crontab", models.ScheduleJob{Name: testName, Definition: invalidCrontab, Actions: []models.ScheduleAction{testRestScheduleAction}}, true},
		{"wrong format crontab", models.ScheduleJob{Name: testName, Definition: wrongFormatCrontab, Actions: []models.ScheduleAction{testRestScheduleAction}}, true},
		{"invalid REST actions", models.ScheduleJob{Name: testName, Definition: testIntervalScheduleDef, Actions: []models.ScheduleAction{invalidRestAction}}, true},
		{"invalid EDGEXMESSAGEBUS actions", models.ScheduleJob{Name: testName, Definition: testIntervalScheduleDef, Actions: []models.ScheduleAction{invalidEdgeXMessageBusAction}}, true},
		{"invalid DEVICECONTROL actions", models.ScheduleJob{Name: testName, Definition: testIntervalScheduleDef, Actions: []models.ScheduleAction{invalidDeviceControlAction}}, true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := mockManager.ValidateUpdatingScheduleJob(testCase.job)
			if testCase.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

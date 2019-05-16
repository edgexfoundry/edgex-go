/*******************************************************************************
 * Copyright 2019 Dell Technologies Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package agent

import (
	"encoding/json"
	"fmt"
	"testing"

	servicesMock "github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/stretchr/testify/mock"
)

/*
TODO: Note that two "services-related" functions are tested elsewhere, in:
TODO: 	/github.com/edgexfoundry/go-mod-core-contracts/clients/general/client_test.go
TODO: Those (two "services-related" functions ) are the following
TODO:	getConfig(...)
TODO:	getMetrics(...)
TODO: Therefore, coverage for those two is not provided here.
*/

func TestProcessResponse(t *testing.T) {

	LoggingClient = logger.NewMockClient()

	var responseJSON = "{\"ApplicationName\":\"support-notifications\",\"RegistryProfilesActive\":\"go\",\"HeartBeatTime\":300000,\"HeartBeatMsg\":\"Support Notifications heart beat\",\"AppOpenMsg\":\"This is the Support Notifications Microservice\", \"FormatSpecifier\":\"%(\\\\d=\\\\$)?([-#= 0(\\\\\u003c]*)?(\\\\d=)?(\\\\.\\\\d=)?([tT])?([a-zA-Z%])\", \"ServicePort\":48060, \"ServiceTimeout\":5000, \"ServiceAddress\":\"localhost\", \"ServiceName\":\"support-notifications\", \"RegistryHost\":\"localhost\", \"RegistryCheckAddress\":\"http://localhost:48060/api/v1/ping\", \"RegistryPort\":8500, \"CheckInterval\":\"10s\", \"EnableRemoteLogging\":false, \"LoggingFile\":\"./logs/edgex-support-notifications.log\", \"LoggingRemoteURL\":\"http://localhost:48061/api/v1/logs\", \"MongoDBUserName\":\"\", \"MongoDBPassword\":\"\", \"MongoDatabaseName\":\"notifications\", \"MongoDBHost\":\"localhost\", \"MongoDBPort\":27017, \"MongoDBConnectTimeout\":60000, \"MongoDBMaxWaitTime\":120000, \"MongoDBKeepAlive\":true, \"MaxResultCount\":50000, \"ResendLimit\":2, \"CleanupDefaultAge\":86400001, \"SchedulerNormalDuration\":\"59 * * * * *\", \"SchedulerNormalResendDuration\":\"59 * * * * *\", \"SchedulerCriticalResendDelay\":10, \"SMTPPort\":\"587\", \"SMTPHost\":\"smtp.gmail.com\", \"SMTPSender\":\"jdoe@gmail.com\", \"SMTPPassword\":\"mypassword\", \"SMTPSubject\":\"EdgeX Notification\", \"DBType\":\"mongodb\"}"

	expResponseJSON := map[string]interface{}{
		"ApplicationName":               "support-notifications",
		"RegistryProfilesActive":        "go",
		"HeartBeatTime":                 "300000",
		"HeartBeatMsg":                  "Support Notifications heart beat",
		"AppOpenMsg":                    "This is the Support Notifications Microservice",
		"FormatSpecifier":               "%(\\\\d=\\\\$)?([-#= 0(\\\\\u003c]*)?(\\\\d=)?(\\\\.\\\\d=)?([tT])?([a-zA-Z%])",
		"ServicePort":                   "48060",
		"ServiceTimeout":                "5000",
		"ServiceAddress":                "localhost",
		"ServiceName":                   "support-notifications",
		"RegistryHost":                  "localhost",
		"RegistryCheckAddress":          "http://localhost:48060/api/v1/ping",
		"RegistryPort":                  "8500",
		"CheckInterval":                 "10s",
		"EnableRemoteLogging":           "false",
		"LoggingFile":                   "./logs/edgex-support-notifications.log",
		"LoggingRemoteURL":              "http://localhost:48061/api/v1/logs",
		"MongoDBUserName":               "",
		"MongoDBPassword":               "",
		"MongoDatabaseName":             "notifications",
		"MongoDBHost":                   "localhost",
		"MongoDBPort":                   "27017",
		"MongoDBConnectTimeout":         "60000",
		"MongoDBMaxWaitTime":            "120000",
		"MongoDBKeepAlive":              "true",
		"MaxResultCount":                "1000",
		"ResendLimit":                   "2",
		"CleanupDefaultAge":             "86400001",
		"SchedulerNormalDuration":       "59 * * * * *",
		"SchedulerNormalResendDuration": "59 * * * * *",
		"SchedulerCriticalResendDelay":  "10",
		"SMTPPort":                      "587",
		"SMTPHost":                      "smtp.gmail.com",
		"SMTPSender":                    "jdoe@gmail.com",
		"SMTPPassword":                  "mypassword",
		"SMTPSubject":                   "EdgeX Notification",
		"DBType":                        "mongodb",
	}

	send := ProcessResponse(responseJSON)
	LoggingClient.Info(fmt.Sprintf("Actual Response: %v", send))

	expected, err := json.Marshal(expResponseJSON)
	if err != nil {
		fmt.Println("Error encoding JSON")
		return
	}

	var exp = make(map[string]interface{})
	err = json.Unmarshal([]byte(expected), &exp)
	if err != nil {
		LoggingClient.Error(err.Error())
	}

	// TODO: Ran into an issue here with the call to reflect.DeepEqual()...
	/*	if !reflect.DeepEqual(send, exp) {
			t.Fatalf("Objects should be equals: %v %v", send.Config, exp)
		}
	*/
}

func reset() {
	Configuration = &ConfigurationStruct{}
}

func TestStartOperation(t *testing.T) {
	reset()
	LoggingClient = logger.NewMockClient()
	serviceList := []string{clients.CoreDataServiceKey, clients.CoreCommandServiceKey}
	mockStarter := &servicesMock.ServiceStarter{}
	mockStarter.On("Start", mock.AnythingOfType("string")).Return(nil)

	tests := []struct {
		name        string
		starter     interface{}
		services    []string
		expectError bool
	}{
		{"start services", mockStarter, serviceList, false},
		{"type check failure", "abc", serviceList, true},
	}
	for _, tt := range tests {
		executorClient = tt.starter
		t.Run(tt.name, func(t *testing.T) {
			err := InvokeOperation("start", tt.services)
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err == nil {
				t.Errorf("did not receive expected error: %s", tt.name)
			}
		})
	}
}

func TestStopOperation(t *testing.T) {
	reset()
	LoggingClient = logger.NewMockClient()
	serviceList := []string{clients.CoreDataServiceKey, clients.CoreCommandServiceKey}
	mockStopper := &servicesMock.ServiceStopper{}
	mockStopper.On("Stop", mock.AnythingOfType("string")).Return(nil)

	tests := []struct {
		name        string
		stopper     interface{}
		services    []string
		expectError bool
	}{
		{"stop services", mockStopper, serviceList, false},
		{"type check failure", "xyz", serviceList, true},
	}
	for _, tt := range tests {
		executorClient = tt.stopper
		t.Run(tt.name, func(t *testing.T) {
			err := InvokeOperation("stop", tt.services)
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err == nil {
				t.Errorf("did not receive expected error: %s", tt.name)
			}
		})
	}
}

func TestRestartOperation(t *testing.T) {
	reset()
	LoggingClient = logger.NewMockClient()
	serviceList := []string{clients.CoreDataServiceKey, clients.CoreCommandServiceKey}
	mockRestarter := &servicesMock.ServiceRestarter{}
	mockRestarter.On("Restart", mock.AnythingOfType("string")).Return(nil)

	tests := []struct {
		name        string
		restarter   interface{}
		services    []string
		expectError bool
	}{
		{"restart services", mockRestarter, serviceList, false},
		{"type check failure", "qrs", serviceList, true},
	}
	for _, tt := range tests {
		executorClient = tt.restarter
		t.Run(tt.name, func(t *testing.T) {
			err := InvokeOperation("restart", tt.services)
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err == nil {
				t.Errorf("did not receive expected error: %s", tt.name)
			}
		})
	}
}

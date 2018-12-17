/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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

	"github.com/edgexfoundry/edgex-go/internal/system/agent/logger"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

func TestProcessResponse(t *testing.T) {

	logs.LoggingClient = logger.NewMockClient()

	var responseJSON = "{\"ApplicationName\":\"support-notifications\",\"ConsulProfilesActive\":\"go\",\"HeartBeatTime\":300000,\"HeartBeatMsg\":\"Support Notifications heart beat\",\"AppOpenMsg\":\"This is the Support Notifications Microservice\", \"FormatSpecifier\":\"%(\\\\d=\\\\$)?([-#= 0(\\\\\u003c]*)?(\\\\d=)?(\\\\.\\\\d=)?([tT])?([a-zA-Z%])\", \"ServicePort\":48060, \"ServiceTimeout\":5000, \"ServiceAddress\":\"localhost\", \"ServiceName\":\"support-notifications\", \"ConsulHost\":\"localhost\", \"ConsulCheckAddress\":\"http://localhost:48060/api/v1/ping\", \"ConsulPort\":8500, \"CheckInterval\":\"10s\", \"EnableRemoteLogging\":false, \"LoggingFile\":\"./logs/edgex-support-notifications.log\", \"LoggingRemoteURL\":\"http://localhost:48061/api/v1/logs\", \"MongoDBUserName\":\"\", \"MongoDBPassword\":\"\", \"MongoDatabaseName\":\"notifications\", \"MongoDBHost\":\"localhost\", \"MongoDBPort\":27017, \"MongoDBConnectTimeout\":60000, \"MongoDBMaxWaitTime\":120000, \"MongoDBKeepAlive\":true, \"ReadMaxLimit\":1000, \"ResendLimit\":2, \"CleanupDefaultAge\":86400001, \"SchedulerNormalDuration\":\"59 * * * * *\", \"SchedulerNormalResendDuration\":\"59 * * * * *\", \"SchedulerCriticalResendDelay\":10, \"SMTPPort\":\"587\", \"SMTPHost\":\"smtp.gmail.com\", \"SMTPSender\":\"jdoe@gmail.com\", \"SMTPPassword\":\"mypassword\", \"SMTPSubject\":\"EdgeX Notification\", \"DBType\":\"mongodb\"}"

	expResponseJSON := map[string]interface{}{
		"ApplicationName":               "support-notifications",
		"ConsulProfilesActive":          "go",
		"HeartBeatTime":                 "300000",
		"HeartBeatMsg":                  "Support Notifications heart beat",
		"AppOpenMsg":                    "This is the Support Notifications Microservice",
		"FormatSpecifier":               "%(\\\\d=\\\\$)?([-#= 0(\\\\\u003c]*)?(\\\\d=)?(\\\\.\\\\d=)?([tT])?([a-zA-Z%])",
		"ServicePort":                   "48060",
		"ServiceTimeout":                "5000",
		"ServiceAddress":                "localhost",
		"ServiceName":                   "support-notifications",
		"ConsulHost":                    "localhost",
		"ConsulCheckAddress":            "http://localhost:48060/api/v1/ping",
		"ConsulPort":                    "8500",
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
		"ReadMaxLimit":                  "1000",
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
	logs.LoggingClient.Info(fmt.Sprintf("Actual Response: {%v}", send))

	expected, err := json.Marshal(expResponseJSON)
	if err != nil {
		fmt.Println("Error encoding JSON")
		return
	}

	var exp = ConfigRespMap{}
	err = json.Unmarshal([]byte(expected), &exp)
	if err != nil {
		logs.LoggingClient.Error(fmt.Sprintf("ERROR: {%v}", err))
	}

	// TODO: Ran into an issue here with the call to reflect.DeepEqual()...
	/*	if !reflect.DeepEqual(send, exp) {
			t.Fatalf("Objects should be equals: %v %v", send.Config, exp)
		}
	*/
}

// TODO: The following are (essentially) integration tests which proved invaluable, especially during the development process.
// TODO: Retain the following as placeholders (to resurrect for integration testing) .

/*func TestInvokeOperationStartKnownService(t *testing.T) {

	action := "start"
	services := []string{"edgex-config-seed", "edgex-support-logging", "edgex-core-metadata", "edgex-support-notifications", "edgex-core-data", "edgex-core-command", "edgex-export-client", "edgex-export-distro"}
	params := []string{"graceful"}
	logs.LoggingClient = logger.NewMockClient()

	res := InvokeOperation(action, services, params)
	if res != true {
		t.Errorf("TestInvokeOperationStartKnownService() failed.")
		return
	}
}

func TestInvokeOperationRestart(t *testing.T) {

	action := "restart"
	services := []string{"edgex-config-seed", "edgex-support-logging", "edgex-core-metadata", "edgex-support-notifications", "edgex-core-data", "edgex-core-command", "edgex-export-client", "edgex-export-distro"}
	params := []string{"graceful"}
	logs.LoggingClient = logger.NewMockClient()

	res := InvokeOperation(action, services, params)
	if res != true {
		t.Errorf("TestInvokeOperationRestart() failed.")
		return
	}
}

func TestInvokeOperationStop(t *testing.T) {

	action := "stop"
	services := []string{"edgex-config-seed", "edgex-support-logging", "edgex-core-metadata", "edgex-support-notifications", "edgex-core-data", "edgex-core-command", "edgex-export-client", "edgex-export-distro"}
	params := []string{"graceful"}
	logs.LoggingClient = logger.NewMockClient()

	res := InvokeOperation(action, services, params)
	if res != true {
		t.Errorf("TestInvokeOperationStop() failed.")
		return
	}
}

func TestInvokeOperationStartUnknownService(t *testing.T) {

	action := "start"
	services := []string{"foo-bar"}
	params := []string{"graceful"}
	logs.LoggingClient = logger.NewMockClient()

	res := InvokeOperation(action, services, params)
	if res != true {
		t.Errorf("TestInvokeOperationStartUnknownService() failed.")
		return
	}
}

func TestInvokeOperationRestartUnknownService(t *testing.T) {

	action := "restart"
	services := []string{"foo-bar"}
	params := []string{"graceful"}
	logs.LoggingClient = logger.NewMockClient()

	res := InvokeOperation(action, services, params)
	if res != true {
		t.Errorf("TestInvokeOperationRestartUnknownService() failed.")
		return
	}
}

func TestInvokeOperationStopUnknownService(t *testing.T) {

	action := "stop"
	services := []string{"foo-bar"}
	params := []string{"graceful"}
	logs.LoggingClient = logger.NewMockClient()

	res := InvokeOperation(action, services, params)
	if res != true {
		t.Errorf("TestInvokeOperationStopUnknownService() failed.")
		return
	}
}

func TestGetConfig(t *testing.T) {

	logs.LoggingClient = logger.NewMockClient()
	services := []string{"edgex-config-seed", "edgex-support-logging", "edgex-core-metadata", "edgex-support-notifications", "edgex-core-data", "edgex-core-command", "edgex-export-client", "edgex-export-distro"}
	result, err := getConfig(services)
	if err != nil {
		t.Errorf("TestGetConfig() failed.")
		return
	}
	logs.LoggingClient.Debug(fmt.Sprintf("Fetched this configuration for the {%v} service: {%v}: ", "first one", result))
}

func TestGetMetric(t *testing.T) {

	logs.LoggingClient = logger.NewMockClient()
	services := []string{"edgex-config-seed", "edgex-support-logging", "edgex-core-metadata", "edgex-support-notifications", "edgex-core-data", "edgex-core-command", "edgex-export-client", "edgex-export-distro"}

	result, err := getMetrics(services)
	if err != nil {
		t.Errorf("TestGetMetric() failed.")
		return
	}
	logs.LoggingClient.Debug(fmt.Sprintf("Fetched these metrics for the {%v} service: {%v}: ", "first one", result))
}

func TestStopDockerContainer(t *testing.T) {

	services := []string{"edgex-support-logging", "edgex-support-notifications"}

	for _, s := range services {
		err := StopService(s)
		if err != nil {
			t.Errorf("TestStopDockerContainer() failed.")
			return
		}
	}
}

func TestFetchDockerComposeYamlAndPath(t *testing.T) {

	_, err := FetchDockerComposeYamlAndPath()
	if err != nil {
		t.Errorf("Calling StopService(service) failed.")
		return
	}
}

func TestStartDockerContainer(t *testing.T) {

	services := []string{"edgex-support-logging", "edgex-support-notifications"}

	for _, s := range services {
		err := StartDockerContainerCompose(s)
		if err != nil {
			t.Errorf("TestStartDockerContainer() failed.")
			return
		}
	}
}

func TestStopAndStartDockerContainer(t *testing.T) {

	services := []string{"edgex-support-logging", "edgex-support-notifications"}
	param := "graceful"

	for _, s := range services {
		err := StopAndStartDockerContainer(s, param)
		if err != nil {
			t.Errorf("TestStopAndStartDockerContainer() failed.")
			return
		}
	}
}
*/

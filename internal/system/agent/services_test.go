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
	"testing"
	"github.com/gorilla/mux"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

var testRoutes *mux.Router

// TODO: Look into mocking relevant infrastructure components...

func TestInvokeActionStart(t *testing.T) {

	action := "START"
	services := []string{"edgex-support-notifications", "edgex-core-data", "edgex-core-metadata"}
	params := map[string]string{"force": "true"}
	LoggingClient = logger.NewMockClient()

	res := invokeAction(action, services, params)
	if res != true {
		t.Errorf("Calling invokeAction() with %v failed.", action)
		return
	}
}

func TestInvokeActionStop(t *testing.T) {

	action := "STOP"
	services := []string{"edgex-support-notifications", "edgex-core-data", "edgex-core-metadata"}
	params := map[string]string{"force": "true"}
	LoggingClient = logger.NewMockClient()

	res := invokeAction(action, services, params)
	if res != true {
		t.Errorf("Calling invokeAction() failed.")
		return
	}
}

func TestInvokeActionRestart(t *testing.T) {

	action := "RESTART"
	services := []string{"edgex-support-notifications", "edgex-core-data", "edgex-core-metadata"}
	params := map[string]string{"force": "true"}
	LoggingClient = logger.NewMockClient()

	res := invokeAction(action, services, params)
	if res != true {
		t.Errorf("Calling invokeAction() failed.")
		return
	}
}

func TestGetConfig(t *testing.T) {

	services := []string{"edgex-support-notifications", "edgex-core-data", "edgex-core-metadata"}

	err := getConfig(services)
	if err != true {
		t.Errorf("Calling invokeAction() failed.")
		return
	}
}

func TestGetMetric(t *testing.T) {

	services := []string{"edgex-support-notifications", "edgex-core-data", "edgex-core-metadata"}
	metrics := []string{"memory", "CPU"}

	err := getMetric(services, metrics)
	if err != true {
		t.Errorf("Calling invokeAction() failed.")
		return
	}
}

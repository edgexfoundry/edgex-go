/*******************************************************************************
 * Copyright 2018 IOTech Ltd
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
 *******************************************************************************/

package metadata

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Test updating a device profile using the device profile client
func TestUpdateDeviceProfile(t *testing.T) {
	p := models.DeviceProfile{
		Id:   "1234",
		Name: "Test name for device profile",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodPut {
			t.Errorf("expected http method is %s, active http method is : %s", http.MethodPut, r.Method)
		}

		if r.URL.EscapedPath() != clients.ApiDeviceProfileRoute {
			t.Errorf("expected uri path is %s, actual uri path is %s", clients.ApiDeviceProfileRoute, r.URL.EscapedPath())
		}

		w.Write([]byte("true"))

	}))

	defer ts.Close()

	url := ts.URL + clients.ApiDeviceProfileRoute

	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        clients.ApiDeviceProfileRoute,
		UseRegistry: false,
		Url:         url,
		Interval:    clients.ClientMonitorDefault}
	dpc := NewDeviceProfileClient(params, MockEndpoint{})

	err := dpc.Update(p)
	if err != nil {
		t.Error(err.Error())
	}
}

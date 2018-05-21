/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/core/clients/types"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/edgexfoundry/edgex-go/internal"
)

const (
	deviceUriPath = "/api/v1/device"
)

// Test adding a device using the device client

// Test adding a device using the device client
func TestAddDevice(t *testing.T) {
	d := models.Device{
		Id:             "1234",
		Addressable:    models.Addressable{},
		AdminState:     "UNLOCKED",
		Name:           "Test name for device",
		OperatingState: "ENABLED",
		Profile:        models.DeviceProfile{},
		Service:        models.DeviceService{},
	}

	addingDeviceId := d.Id.Hex()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodPost {
			t.Errorf("expected http method is %s, active http method is : %s", http.MethodPost, r.Method)
		}

		if r.URL.EscapedPath() != deviceUriPath {
			t.Errorf("expected uri path is %s, actual uri path is %s", deviceUriPath, r.URL.EscapedPath())
		}

		w.Write([]byte(addingDeviceId))

	}))

	defer ts.Close()

	url := ts.URL + deviceUriPath

	params := types.EndpointParams{
		ServiceKey:internal.MetaDataServiceKey,
		Path:deviceUriPath,
		UseRegistry:false,
		Url:url}
	dc, err := NewDeviceClient(params, MockEndpoint{})

	receivedDeviceId, err := dc.Add(&d)
	if err != nil {
		t.Error(err.Error())
	}

	if receivedDeviceId != addingDeviceId {
		t.Errorf("expected device id : %s, actual device id : %s", receivedDeviceId, addingDeviceId)
	}
}


func TestNewDeviceClientWithConsul(t *testing.T) {
	deviceUrl := "http://localhost:48081" + deviceUriPath
	params := types.EndpointParams{
		ServiceKey:internal.MetaDataServiceKey,
		Path:deviceUriPath,
		UseRegistry:true,
		Url:deviceUrl}

	dc, err := NewDeviceClient(params, MockEndpoint{})
	if err != nil {
		t.Error(err)
	}
	r, ok := dc.(*DeviceRestClient)
	if !ok {
		t.Error("dc is not of expected type")
	}

	time.Sleep(25 * time.Millisecond)
	if len(r.url) == 0 {
		t.Error("url was not initialized")
	} else if r.url != deviceUrl {
		t.Errorf("unexpected url value %s", r.url)
	}
}

type MockEndpoint struct {

}

func(e MockEndpoint) Monitor(params types.EndpointParams, ch chan string) {
	switch (params.ServiceKey) {
	case internal.MetaDataServiceKey:
		url := fmt.Sprintf("http://%s:%v%s", "localhost", 48081, params.Path)
		ch <- url
		break
	default:
		ch <- ""
	}
}
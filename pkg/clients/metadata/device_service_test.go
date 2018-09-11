/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

func TestNewDeviceServiceClientWithConsul(t *testing.T) {
	deviceServiceUrl := "http://localhost:48081" + clients.ApiDeviceServiceRoute
	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        clients.ApiDeviceServiceRoute,
		UseRegistry: true,
		Url:         deviceServiceUrl,
		Interval:    clients.ClientMonitorDefault}

	dsc := NewDeviceServiceClient(params, MockEndpoint{})
	r, ok := dsc.(*DeviceServiceRestClient)
	if !ok {
		t.Error("dsc is not of expected type")
	}

	time.Sleep(25 * time.Millisecond)
	if len(r.url) == 0 {
		t.Error("url was not initialized")
	} else if r.url != deviceServiceUrl {
		t.Errorf("unexpected url value %s", r.url)
	}
}

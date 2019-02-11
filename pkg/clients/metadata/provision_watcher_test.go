/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright (C) 2018 Canonical Ltd
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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Test adding a provision watcher using the client
func TestAddProvisionWatcher(t *testing.T) {
	se := models.ProvisionWatcher{
		Id:             "1234",
		Name:           "Test name for provision watcher",
		Profile:        models.DeviceProfile{},
		Service:        models.DeviceService{},
		OperatingState: models.Enabled,
	}

	addingProvisionWatcherID := se.Id

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodPost {
			t.Errorf("expected http method is %s, active http method is : %s", http.MethodPost, r.Method)
		}

		if r.URL.EscapedPath() != clients.ApiProvisionWatcherRoute {
			t.Errorf("expected uri path is %s, actual uri path is %s", clients.ApiProvisionWatcherRoute, r.URL.EscapedPath())
		}

		w.Write([]byte(addingProvisionWatcherID))

	}))

	defer ts.Close()

	url := ts.URL + clients.ApiProvisionWatcherRoute

	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        clients.ApiProvisionWatcherRoute,
		UseRegistry: false,
		Url:         url,
		Interval:    clients.ClientMonitorDefault}
	sc := NewProvisionWatcherClient(params, MockEndpoint{})

	receivedProvisionWatcherID, err := sc.Add(&se, context.Background())
	if err != nil {
		t.Error(err.Error())
	}

	if receivedProvisionWatcherID != addingProvisionWatcherID {
		t.Errorf("expected provision watcher id : %s, actual provision watcher id : %s", receivedProvisionWatcherID, addingProvisionWatcherID)
	}
}

func TestNewProvisionWatcherClientWithConsul(t *testing.T) {
	provisionWatcherURL := "http://localhost:48081" + clients.ApiProvisionWatcherRoute
	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        clients.ApiProvisionWatcherRoute,
		UseRegistry: true,
		Url:         provisionWatcherURL,
		Interval:    clients.ClientMonitorDefault}

	sc := NewProvisionWatcherClient(params, MockEndpoint{})

	r, ok := sc.(*ProvisionWatcherRestClient)
	if !ok {
		t.Error("sc is not of expected type")
	}

	time.Sleep(25 * time.Millisecond)
	if len(r.url) == 0 {
		t.Error("url was not initialized")
	} else if r.url != provisionWatcherURL {
		t.Errorf("unexpected url value %s", r.url)
	}
}

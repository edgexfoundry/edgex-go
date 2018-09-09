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
package coredata

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

const (
	ValueDescriptorUriPath         = "/api/v1/valuedescriptor"
	TestValueDesciptorDescription1 = "value descriptor1"
	TestValueDesciptorDescription2 = "value descriptor2"
)

func TestGetvaluedescriptors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodGet {
			t.Errorf("expected http method is GET, active http method is : %s", r.Method)
		}

		if r.URL.EscapedPath() != ValueDescriptorUriPath {
			t.Errorf("expected uri path is %s, actual uri path is %s", ValueDescriptorUriPath, r.URL.EscapedPath())
		}

		w.Write([]byte("[" +
			"{" +
			"\"Description\" : \"" + TestValueDesciptorDescription1 + "\"" +
			"}," +
			"{" +
			"\"Description\" : \"" + TestValueDesciptorDescription2 + "\"" +
			"}" +
			"]"))

	}))

	defer ts.Close()

	url := ts.URL + ValueDescriptorUriPath

	params := types.EndpointParams{
		ServiceKey:  internal.CoreDataServiceKey,
		Path:        ValueDescriptorUriPath,
		UseRegistry: false,
		Url:         url,
		Interval:    clients.ClientMonitorDefault}

	vdc := NewValueDescriptorClient(params, mockEndpoint{})

	vdArr, err := vdc.ValueDescriptors()
	if err != nil {
		t.FailNow()
	}

	if len(vdArr) != 2 {
		t.Errorf("expected value descriptor array's length is 2, actual array's length is : %d", len(vdArr))
	}

	vd1 := vdArr[0]
	if vd1.Description != TestValueDesciptorDescription1 {
		t.Errorf("expected first value descriptor's description is : %s, actual description is : %s", TestValueDesciptorDescription1, vd1.Description)
	}

	vd2 := vdArr[1]
	if vd2.Description != TestValueDesciptorDescription2 {
		t.Errorf("expected second value descriptor's description is : %s, actual description is : %s ", TestValueDesciptorDescription2, vd2.Description)
	}
}

func TestNewValueDescriptorClientWithConsul(t *testing.T) {
	deviceUrl := "http://localhost:48080" + ValueDescriptorUriPath
	params := types.EndpointParams{
		ServiceKey:  internal.CoreDataServiceKey,
		Path:        ValueDescriptorUriPath,
		UseRegistry: true,
		Url:         deviceUrl,
		Interval:    clients.ClientMonitorDefault}

	vdc := NewValueDescriptorClient(params, mockEndpoint{})

	r, ok := vdc.(*ValueDescriptorRestClient)
	if !ok {
		t.Error("vdc is not of expected type")
	}

	time.Sleep(25 * time.Millisecond)
	if len(r.url) == 0 {
		t.Error("url was not initialized")
	} else if r.url != deviceUrl {
		t.Errorf("unexpected url value %s", r.url)
	}
}

type mockEndpoint struct {
}

func (e mockEndpoint) Monitor(params types.EndpointParams, ch chan string) {
	switch params.ServiceKey {
	case internal.CoreDataServiceKey:
		url := fmt.Sprintf("http://%s:%v%s", "localhost", 48080, params.Path)
		ch <- url
		break
	default:
		ch <- ""
	}
}

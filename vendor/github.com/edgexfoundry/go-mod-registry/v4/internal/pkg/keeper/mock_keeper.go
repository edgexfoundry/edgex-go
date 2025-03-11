//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
)

type MockKeeper struct {
	serviceStore map[string]dtos.Registration
	serviceLock  sync.Mutex
}

func NewMockKeeper() *MockKeeper {
	mock := MockKeeper{
		serviceStore: make(map[string]dtos.Registration),
	}

	return &mock
}

func (mock *MockKeeper) Start() *httptest.Server {
	testMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasSuffix(request.URL.Path, common.ApiRegisterRoute) {
			switch request.Method {
			case http.MethodPost:
				mock.serviceLock.Lock()
				defer mock.serviceLock.Unlock()

				bodyBytes, err := io.ReadAll(request.Body)
				if err != nil {
					log.Printf("error reading request body: %s", err.Error())
				}

				var req requests.AddRegistrationRequest
				err = json.Unmarshal(bodyBytes, &req)
				if err != nil {
					log.Printf("error decoding request body: %s", err.Error())
				}

				resp, err := http.Get(req.Registration.HealthCheck.Type + "://" + req.Registration.Host + ":" + strconv.Itoa(req.Registration.Port) + req.Registration.HealthCheck.Path)
				if err != nil {
					log.Printf("error health checking: %s", err.Error())
				} else {
					if resp.StatusCode == http.StatusOK {
						req.Registration.Status = "UP"
					} else {
						req.Registration.Status = "DOWN"
					}
				}
				mock.serviceStore[req.Registration.ServiceId] = req.Registration

				writer.Header().Set(common.ContentTypeJSON, common.ContentTypeJSON)
				writer.WriteHeader(http.StatusCreated)
			case http.MethodPut:
				mock.serviceLock.Lock()
				defer mock.serviceLock.Unlock()

				bodyBytes, err := io.ReadAll(request.Body)
				if err != nil {
					log.Printf("error reading request body: %s", err.Error())
				}

				var req requests.AddRegistrationRequest
				err = json.Unmarshal(bodyBytes, &req)
				if err != nil {
					log.Printf("error decoding request body: %s", err.Error())
				}
				mock.serviceStore[req.Registration.ServiceId] = req.Registration

				writer.WriteHeader(http.StatusNoContent)
			}
		} else if strings.HasSuffix(request.URL.Path, common.ApiAllRegistrationsRoute) {
			switch request.Method {
			case http.MethodGet:
				mock.serviceLock.Lock()
				defer mock.serviceLock.Unlock()

				var registrations []dtos.Registration
				for _, r := range mock.serviceStore {
					registrations = append(registrations, r)
				}
				resp := responses.MultiRegistrationsResponse{
					BaseWithTotalCountResponse: dtoCommon.BaseWithTotalCountResponse{
						BaseResponse: dtoCommon.BaseResponse{
							Versionable: dtoCommon.Versionable{ApiVersion: common.ApiVersion},
							RequestId:   "",
							Message:     "",
							StatusCode:  200,
						},
						TotalCount: uint32(len(mock.serviceStore)),
					},
					Registrations: registrations,
				}
				jsonData, _ := json.Marshal(resp)
				writer.Header().Set(common.ContentType, common.ContentTypeJSON)
				writer.WriteHeader(http.StatusOK)
				_, err := writer.Write(jsonData)
				if err != nil {
					log.Printf("error writing data response: %s", err.Error())
				}
			}
		} else if strings.Contains(request.URL.Path, ApiRegistrationByServiceIdRoute) {
			key := strings.Replace(request.URL.Path, ApiRegistrationByServiceIdRoute, "", 1)
			switch request.Method {
			case http.MethodGet:
				var resp interface{}
				r, ok := mock.serviceStore[key]
				if !ok {
					resp = dtoCommon.BaseResponse{
						Versionable: dtoCommon.Versionable{ApiVersion: common.ApiVersion},
						RequestId:   "",
						Message:     "not found",
						StatusCode:  http.StatusNotFound,
					}
				} else {
					resp = responses.RegistrationResponse{
						BaseResponse: dtoCommon.BaseResponse{
							Versionable: dtoCommon.Versionable{ApiVersion: common.ApiVersion},
							RequestId:   "",
							Message:     "",
							StatusCode:  http.StatusOK,
						},
						Registration: r,
					}
				}

				jsonData, _ := json.Marshal(resp)
				writer.Header().Set(common.ContentType, common.ContentTypeJSON)
				_, err := writer.Write(jsonData)
				if err != nil {
					log.Printf("error writing data response: %s", err.Error())
				}
			case http.MethodDelete:
				mock.serviceLock.Lock()
				defer mock.serviceLock.Unlock()

				_, ok := mock.serviceStore[key]
				if ok {
					delete(mock.serviceStore, key)
				}

				writer.WriteHeader(http.StatusNoContent)
			}
		} else if strings.Contains(request.URL.Path, common.ApiPingRoute) {
			switch request.Method {
			case http.MethodGet:
				resp := dtoCommon.PingResponse{
					Versionable: dtoCommon.Versionable{ApiVersion: common.ApiVersion},
					Timestamp:   "",
					ServiceName: "",
				}
				jsonData, _ := json.Marshal(resp)
				writer.Header().Set(common.ContentType, common.ContentTypeJSON)
				_, _ = writer.Write(jsonData)
			}
		}
	}))

	return testMockServer
}

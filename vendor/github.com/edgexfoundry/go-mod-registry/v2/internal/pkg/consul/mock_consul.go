//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package consul

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	consulapi "github.com/hashicorp/consul/api"
)

const (
	verbose  = false
	TokenKey = "X-Consul-Token"
)

type MockConsul struct {
	keyValueStore       map[string]*consulapi.KVPair
	serviceStore        map[string]consulapi.AgentService
	serviceCheckStore   map[string]consulapi.AgentCheck
	serviceLock         sync.Mutex
	expectedAccessToken string
}

func NewMockConsul() *MockConsul {
	mock := MockConsul{
		keyValueStore:     make(map[string]*consulapi.KVPair),
		serviceStore:      make(map[string]consulapi.AgentService),
		serviceCheckStore: make(map[string]consulapi.AgentCheck),
	}

	return &mock
}

func (mock *MockConsul) Start() *httptest.Server {
	testMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if len(mock.expectedAccessToken) > 0 {
			token := request.Header.Get(TokenKey)
			if len(mock.expectedAccessToken) > 0 && token != mock.expectedAccessToken {
				writer.WriteHeader(http.StatusForbidden)
				return
			}
		}

		if strings.HasSuffix(request.URL.Path, "/v1/agent/service/register") {
			switch request.Method {
			case "PUT":
				mock.serviceLock.Lock()
				defer mock.serviceLock.Unlock()

				body := make([]byte, request.ContentLength)
				if _, err := io.ReadFull(request.Body, body); err != nil {
					log.Printf("error reading request body: %s", err.Error())
				}
				// AgentServiceRegistration struct represents how service registration information is received
				var mockServiceRegister consulapi.AgentServiceRegistration

				// AgentService struct represent how service information is store internally
				var mockService consulapi.AgentService
				// unmarshal request body
				if err := json.Unmarshal(body, &mockServiceRegister); err != nil {
					log.Printf("error reading request body: %s", err.Error())
				}

				// Copying over basic fields required for current test cases.
				mockService.ID = mockServiceRegister.Name
				mockService.Service = mockServiceRegister.Name
				mockService.Address = mockServiceRegister.Address
				mockService.Port = mockServiceRegister.Port

				mock.serviceStore[mockService.ID] = mockService
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)

			}
		} else if strings.HasSuffix(request.URL.Path, "/v1/agent/services") {
			switch request.Method {
			case "GET":
				mock.serviceLock.Lock()
				defer mock.serviceLock.Unlock()

				jsonData, _ := json.MarshalIndent(&mock.serviceStore, "", "  ")

				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)
				if _, err := writer.Write(jsonData); err != nil {
					log.Printf("error writing data response: %s", err.Error())
				}

			}
		} else if strings.Contains(request.URL.Path, "/v1/agent/service/deregister/") {
			key := strings.Replace(request.URL.Path, "/v1/agent/service/deregister/", "", 1)
			switch request.Method {
			case "PUT":
				mock.serviceLock.Lock()
				defer mock.serviceLock.Unlock()

				_, ok := mock.serviceStore[key]
				if ok {
					delete(mock.serviceStore, key)
				}

				_, ok = mock.serviceCheckStore[key]
				if ok {
					delete(mock.serviceCheckStore, key)
				}
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)

			}
		} else if strings.Contains(request.URL.Path, "/v1/status/leader") {
			switch request.Method {
			case "GET":
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)

			}

		} else if strings.Contains(request.URL.Path, "/agent/check/register") {
			switch request.Method {
			case "PUT":
				body := make([]byte, request.ContentLength)
				if _, err := io.ReadFull(request.Body, body); err != nil {
					log.Printf("error reading request body: %s", err.Error())
				}

				var healthCheck consulapi.AgentCheckRegistration
				if err := json.Unmarshal(body, &healthCheck); err != nil {
					log.Printf("error reading request body: %s", err.Error())
				}
				fmt.Println(fmt.Sprintf("%+v", healthCheck))
				// if endpoint for health check is set, then try call the endpoint once after interval.
				if healthCheck.AgentServiceCheck.HTTP != "" && healthCheck.AgentServiceCheck.Interval != "" {
					go func() {
						mock.serviceLock.Lock()
						defer mock.serviceLock.Unlock()

						check := consulapi.AgentCheck{
							Node:        "Mock Consul server",
							CheckID:     healthCheck.ID,
							Name:        "Health Check: " + healthCheck.ServiceID,
							Status:      "TBD",
							Output:      "TBD",
							ServiceID:   healthCheck.ServiceID,
							ServiceName: healthCheck.ServiceID,
						}

						response, err := http.Get(healthCheck.AgentServiceCheck.HTTP)

						if err != nil || response.StatusCode != http.StatusOK {
							check.Status = "critical"
							check.Output = "HTTP GET " + healthCheck.AgentServiceCheck.HTTP + ": health check endpoint unreachable"

							if verbose {
								log.Print("Not able to reach health check endpoint")
							}
						} else {

							check.Status = "passing"
							check.Output = "HTTP GET " + healthCheck.AgentServiceCheck.HTTP + ": 200 OK Output: pong"

							if verbose {
								log.Print("Health check endpoint is reachable!")
							}
						}

						mock.serviceCheckStore[healthCheck.ID] = check

					}()

				}

				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)
			}
		} else if strings.Contains(request.URL.Path, "/agent/check/deregister/") {
			key := strings.Replace(request.URL.Path, "/v1/agent/check/deregister/", "", 1)
			switch request.Method {
			case "PUT":
				mock.serviceLock.Lock()
				defer mock.serviceLock.Unlock()
				fmt.Println(key)
				_, ok := mock.serviceCheckStore[key]
				if ok {
					delete(mock.serviceCheckStore, key)
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
				} else {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusBadRequest)
				}
			}
		} else if strings.Contains(request.URL.Path, "/v1/health/checks") {
			switch request.Method {
			case "GET":
				mock.serviceLock.Lock()
				defer mock.serviceLock.Unlock()

				agentChecks := make([]consulapi.AgentCheck, 0)
				key := strings.Replace(request.URL.Path, "/v1/health/checks/", "", 1)
				check, ok := mock.serviceCheckStore[key]
				if ok {
					agentChecks = append(agentChecks, check)
				}

				jsonData, _ := json.MarshalIndent(&agentChecks, "", "  ")

				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)
				if _, err := writer.Write(jsonData); err != nil {
					log.Printf("error writing data response: %s", err.Error())
				}
			}
		}
	}))

	return testMockServer
}

func (mock *MockConsul) SetExpectedAccessToken(token string) {
	mock.expectedAccessToken = token
}

func (mock *MockConsul) ClearExpectedAccessToken() {
	mock.expectedAccessToken = ""
}

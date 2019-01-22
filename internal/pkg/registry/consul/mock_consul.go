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
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

type MockConsul struct {
	keyValueStore map[string]*consulapi.KVPair
	serviceStore  map[string]consulapi.AgentService
}

func NewMockConsul() *MockConsul {
	mock := MockConsul{}
	mock.keyValueStore = make(map[string]*consulapi.KVPair)
	mock.serviceStore = make(map[string]consulapi.AgentService)
	return &mock
}

var keyChannels map[string]chan bool
var PrefixChannels map[string]chan bool

func (mock *MockConsul) Start() *httptest.Server {
	keyChannels = make(map[string]chan bool)
	var consulIndex = 1
	PrefixChannels = make(map[string]chan bool)

	testMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.URL.Path, "/v1/kv/") {
			key := strings.Replace(request.URL.Path, "/v1/kv/", "", 1)

			switch request.Method {
			case "PUT":
				body := make([]byte, request.ContentLength)
				if _, err := io.ReadFull(request.Body, body); err != nil {
					log.Printf("error reading request body: %s", err.Error())
				}

				keyValuePair, found := mock.keyValueStore[key]
				if found {
					keyValuePair.ModifyIndex++
					keyValuePair.Value = body
				} else {
					keyValuePair = &consulapi.KVPair{
						Key:         key,
						Value:       body,
						ModifyIndex: 1,
						CreateIndex: 1,
						Flags:       0,
						LockIndex:   0,
					}
				}

				mock.keyValueStore[key] = keyValuePair

				log.Printf("PUTing new value for %s", key)
				channel, found := keyChannels[key]
				if found {
					channel <- true
				}
				for prefix, channel := range PrefixChannels {
					if strings.HasPrefix(key, prefix) {
						consulIndex++
						if channel != nil {
							channel <- true
						}
					}
				}
			case "GET":
				// this is what the wait query parameters will look like "index=1&wait=600000ms"
				var pairs consulapi.KVPairs
				var prefixFound bool
				query := request.URL.Query()
				waitTime := query.Get("wait")
				// Recurse parameters are usually set when prefix is monitored,
				// if found we need to find all keys with prefix set in URL.
				_, recursefound := query["recurse"]
				if recursefound {
					pairs, prefixFound = mock.checkForPrefix(key)
					if !prefixFound {
						http.NotFound(writer, request)
						return
					}
					//Default waitime is 30 minutes, overiding it for unit test purpose
					waitTime = "2s"
					if waitTime != "" {
						waitForNextPutPrefix(key, waitTime)
					}
					writer.Header().Set("X-Consul-Index", strconv.Itoa(consulIndex))
				} else {
					keyValuePair, found := mock.keyValueStore[key]
					pairs = consulapi.KVPairs{keyValuePair}
					if !found {
						http.NotFound(writer, request)
						return
					}
					if waitTime != "" {
						waitForNextPut(key, waitTime)
					}

				}

				jsonData, _ := json.MarshalIndent(&pairs, "", "  ")

				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)
				if _, err := writer.Write(jsonData); err != nil {
					log.Printf("error writing data response: %s", err.Error())
				}
			}
		} else if strings.HasSuffix(request.URL.Path, "/v1/agent/service/register") {
			switch request.Method {
			case "PUT":
				body := make([]byte, request.ContentLength)
				if _, err := io.ReadFull(request.Body, body); err != nil {
					log.Printf("error reading request body: %s", err.Error())
				}
				//AgentServiceRegistration struct represents how service registration information is recieved
				var mockServiceRegister consulapi.AgentServiceRegistration

				//AgentService struct represent how service information is store internally
				var mockService consulapi.AgentService
				// unmarshal request body
				if err := json.Unmarshal(body, &mockServiceRegister); err != nil {
					log.Printf("error reading request body: %s", err.Error())
				}

				//Copying over basic fields required for current test cases.
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
				_, ok := mock.serviceStore[key]
				if ok {
					delete(mock.serviceStore, key)
				}
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)

			}
		} else if strings.Contains(request.URL.Path, "/v1/agent/self") {
			switch request.Method {
			case "GET":
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)

			}

		} else if strings.Contains(request.URL.Path, "/agent/check/register") {
			switch request.Method {
			case "PUT":
				//TODO: Need to put in the logic how this data needs to be handled
				body := make([]byte, request.ContentLength)
				if _, err := io.ReadFull(request.Body, body); err != nil {
					log.Printf("error reading request body: %s", err.Error())
				}

				var healthCheck consulapi.AgentCheckRegistration
				if err := json.Unmarshal(body, &healthCheck); err != nil {
					log.Printf("error reading request body: %s", err.Error())
				}

				//if endpoint for healthcheck is set, then try call the endpoint once after interval.
				if healthCheck.AgentServiceCheck.HTTP != "" && healthCheck.AgentServiceCheck.Interval != "" {
					timeout, err := time.ParseDuration(healthCheck.AgentServiceCheck.Interval)
					if err != nil {
						log.Printf("Error parsing waitTime %s into a duration: %s", timeout, err.Error())
					}
					go func() {
						time.Sleep(timeout)
						_, err := http.Get(healthCheck.AgentServiceCheck.HTTP + consulStatusPath)
						if err != nil {
							log.Print("Not able to reach health check endpoint")
						}
						log.Print("Health check endpoint is reachable!")
					}()

				}

				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)

			}

		}

	}))

	return testMockServer
}

func waitForNextPut(key string, waitTime string) {
	timeout, err := time.ParseDuration(waitTime)
	if err != nil {
		log.Printf("Error parsing waitTime %s into a duration: %s", waitTime, err.Error())
	}
	channel := make(chan bool)
	keyChannels[key] = channel
	timedOut := false
	go func() {
		time.Sleep(timeout)
		timedOut = true
		if keyChannels[key] != nil {
			keyChannels[key] <- true
			log.Printf("Timed out watching for change on %s", key)
		}
	}()

	log.Printf("Watching for change on %s", key)
	<-channel
	close(channel)
	keyChannels[key] = nil
	if !timedOut {
		log.Printf("%s changed", key)
	}
}
func waitForNextPutPrefix(key string, waitTime string) {
	timeout, err := time.ParseDuration(waitTime)
	if err != nil {
		log.Printf("Error parsing waitTime %s into a duration: %s", waitTime, err.Error())
	}
	channel := make(chan bool)
	PrefixChannels[key] = channel
	timedOut := false
	go func() {
		time.Sleep(timeout)
		timedOut = true
		if PrefixChannels[key] != nil {
			PrefixChannels[key] <- true
			log.Printf("Timed out watching for change on %s", key)
		}
	}()

	log.Printf("Watching for change on %s", key)
	<-channel
	close(channel)
	PrefixChannels[key] = nil
	if !timedOut {
		log.Printf("%s changed", key)
	}
}

func (mock *MockConsul) checkForPrefix(prefix string) (consulapi.KVPairs, bool) {
	var pairs consulapi.KVPairs
	for k, v := range mock.keyValueStore {
		if strings.HasPrefix(k, prefix) {
			pairs = append(pairs, v)
		}
	}
	if len(pairs) == 0 {
		return nil, false
	}
	return pairs, true

}

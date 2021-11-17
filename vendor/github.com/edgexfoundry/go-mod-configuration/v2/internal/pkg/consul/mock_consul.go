//
// Copyright (c) 2021 Intel Corporation
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

const (
	verbose  = false
	TokenKey = "X-Consul-Token"
)

type MockConsul struct {
	keyValueStore       map[string]*consulapi.KVPair
	serviceStore        map[string]consulapi.AgentService
	serviceCheckStore   map[string]consulapi.AgentCheck
	expectedAccessToken string
}

func NewMockConsul() *MockConsul {
	return &MockConsul{
		keyValueStore:     make(map[string]*consulapi.KVPair),
		serviceStore:      make(map[string]consulapi.AgentService),
		serviceCheckStore: make(map[string]consulapi.AgentCheck),
	}
}

var keyChannels map[string]chan bool
var PrefixChannels map[string]chan bool

func (mock *MockConsul) Reset() {
	mock.keyValueStore = make(map[string]*consulapi.KVPair)
	mock.serviceStore = make(map[string]consulapi.AgentService)
	mock.serviceCheckStore = make(map[string]consulapi.AgentCheck)
}

func (mock *MockConsul) Start() *httptest.Server {
	keyChannels = make(map[string]chan bool)
	var consulIndex = 1
	PrefixChannels = make(map[string]chan bool)

	testMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if len(mock.expectedAccessToken) > 0 {
			token := request.Header.Get(TokenKey)
			if len(mock.expectedAccessToken) > 0 && token != mock.expectedAccessToken {
				writer.WriteHeader(http.StatusForbidden)
				return
			}
		}

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

				if verbose {
					log.Printf("PUTing new value for %s", key)
				}

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
				_, recurseFound := query["recurse"]
				_, allKeysRequested := query["keys"]
				if recurseFound {
					pairs, prefixFound = mock.checkForPrefix(key)
					if !prefixFound {
						http.NotFound(writer, request)
						return
					}
					//Default wait time is 30 minutes, over riding it for unit test purpose
					waitTime = "1s"
					if waitTime != "" {
						waitForNextPutPrefix(key, waitTime)
					}
					writer.Header().Set("X-Consul-Index", strconv.Itoa(consulIndex))
				} else if allKeysRequested {
					// Just returning array of key names
					var keys []string

					pairs, prefixFound = mock.checkForPrefix(key)
					if !prefixFound {
						http.NotFound(writer, request)
						return
					}

					for _, key := range pairs {
						keys = append(keys, key.Key)
					}

					jsonData, _ := json.MarshalIndent(&keys, "", "  ")

					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					if _, err := writer.Write(jsonData); err != nil {
						log.Printf("error writing data response: %s", err.Error())
					}

				} else {
					keyValuePair, found := mock.keyValueStore[key]
					pairs = consulapi.KVPairs{keyValuePair}
					if !found {
						http.NotFound(writer, request)
						return
					}
				}

				jsonData, _ := json.MarshalIndent(&pairs, "", "  ")

				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)
				if _, err := writer.Write(jsonData); err != nil {
					log.Printf("error writing data response: %s", err.Error())
				}
			}
		} else if strings.Contains(request.URL.Path, "/v1/status/leader") {
			switch request.Method {
			case "GET":
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)

			}
		}
	}))

	return testMockServer
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
			if verbose {
				log.Printf("Timed out watching for change on %s", key)
			}
		}
	}()

	if verbose {
		log.Printf("Watching for change on %s", key)
	}

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

func (mock *MockConsul) SetExpectedAccessToken(token string) {
	mock.expectedAccessToken = token
}

func (mock *MockConsul) ClearExpectedAccessToken() {
	mock.expectedAccessToken = ""
}

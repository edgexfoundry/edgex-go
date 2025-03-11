//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

const apiKVRoute = common.ApiKVSRoute + "/" + common.Key

type MockCoreKeeper struct {
	keyValueStore map[string]models.KVS
}

func NewMockCoreKeeper() *MockCoreKeeper {
	return &MockCoreKeeper{
		keyValueStore: make(map[string]models.KVS),
	}
}

func (mock *MockCoreKeeper) Reset() {
	mock.keyValueStore = make(map[string]models.KVS)
}

func (mock *MockCoreKeeper) Start() *httptest.Server {
	testMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.URL.Path, apiKVRoute) {
			key := strings.Replace(request.URL.Path, apiKVRoute+"/", "", 1)

			switch request.Method {
			case http.MethodPut:
				body, err := io.ReadAll(request.Body)
				if err != nil {
					log.Printf("error reading request body: %s", err.Error())
				}
				var updateKeysRequest requests.UpdateKeysRequest
				err = json.Unmarshal(body, &updateKeysRequest)
				if err != nil {
					log.Printf("error decode the request body: %s", err.Error())
				}
				query := request.URL.Query()
				_, isFlatten := query[common.Flatten]
				if isFlatten {
					kvPairs := convertInterfaceToPairs(key, updateKeysRequest.Value)
					for _, kvPair := range kvPairs {
						mock.updateKVStore(kvPair.Key, kvPair.Value)
					}
				} else {
					mock.updateKVStore(key, updateKeysRequest.Value)
				}
			case http.MethodGet:
				query := request.URL.Query()
				_, allKeysRequested := query[common.KeyOnly]

				var resp interface{}
				pairs, prefixFound := mock.checkForPrefix(key)
				if !prefixFound {
					resp = dtoCommon.BaseResponse{
						Message:    fmt.Sprintf("query key %s not found", key),
						StatusCode: http.StatusNotFound,
					}
					writer.WriteHeader(http.StatusNotFound)
				} else {
					if allKeysRequested {
						var keys []models.KeyOnly
						// Just returning array of key paths
						for _, kvPair := range pairs {
							keys = append(keys, models.KeyOnly(kvPair.Key))
						}
						resp = responses.KeysResponse{Response: keys}
					} else {
						var kvs []models.KVS

						// Just returning array of key-value pairs
						for _, kvPair := range pairs {
							kv := models.KVS{
								Key: kvPair.Key,
								StoredData: models.StoredData{
									Value: kvPair.Value,
								},
							}
							kvs = append(kvs, kv)
						}
						resp = responses.MultiKeyValueResponse{Response: kvs}
					}
					writer.WriteHeader(http.StatusOK)
				}
				writer.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(writer).Encode(resp); err != nil {
					log.Printf("error writing data response: %s", err.Error())
				}
			}
		} else if strings.Contains(request.URL.Path, common.ApiPingRoute) {
			switch request.Method {
			case http.MethodGet:
				writer.WriteHeader(http.StatusOK)

			}
		}
	}))

	return testMockServer
}

func (mock *MockCoreKeeper) checkForPrefix(prefix string) ([]models.KVS, bool) {
	var pairs []models.KVS
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

// updateKVStore updates the value of the specified key from the mock key-value store map
func (mock *MockCoreKeeper) updateKVStore(key string, value interface{}) {
	keyValuePair, found := mock.keyValueStore[key]
	if found {
		keyValuePair.Value = value
	} else {
		keyValuePair = models.KVS{
			Key: key,
			StoredData: models.StoredData{
				Value: value,
			},
		}
	}
	mock.keyValueStore[key] = keyValuePair
}

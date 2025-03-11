/*
	Copyright 2019 NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package posture

import (
	"github.com/go-openapi/runtime"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/openziti/foundation/v2/stringz"
	cmap "github.com/orcaman/concurrent-map/v2"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type CacheData struct {
	Processes    cmap.ConcurrentMap[string, ProcessInfo] // map[processPath]ProcessInfo
	MacAddresses []string
	Os           OsInfo
	Domain       string
	Evaluated    atomic.Bool //marks whether posture responses for this data have been sent out
}

func NewCacheData() *CacheData {
	return &CacheData{
		Processes:    cmap.New[ProcessInfo](),
		MacAddresses: []string{},
		Os: OsInfo{
			Type:    "",
			Version: "",
		},
		Domain: "",
	}
}

type Cache struct {
	currentData  *CacheData
	previousData *CacheData

	watchedProcesses cmap.ConcurrentMap[string, struct{}] //map[processPath]struct{}{}

	serviceQueryMap concurrenz.AtomicValue[map[string]map[string]rest_model.PostureQuery] //map[serviceId]map[queryId]query
	activeServices  cmap.ConcurrentMap[string, struct{}]                                  // map[serviceId]

	lastSent   cmap.ConcurrentMap[string, time.Time] //map[type|processQueryId]time.Time
	ctrlClient Submitter

	startOnce           sync.Once
	doSingleSubmissions bool
	closeNotify         <-chan struct{}

	DomainFunc func() string
	lock       sync.Mutex
}

func NewCache(submitter Submitter, closeNotify <-chan struct{}) *Cache {
	cache := &Cache{
		currentData:      NewCacheData(),
		previousData:     NewCacheData(),
		watchedProcesses: cmap.New[struct{}](),
		activeServices:   cmap.New[struct{}](),
		lastSent:         cmap.New[time.Time](),
		ctrlClient:       submitter,
		startOnce:        sync.Once{},
		closeNotify:      closeNotify,
		DomainFunc:       Domain,
	}
	cache.serviceQueryMap.Store(map[string]map[string]rest_model.PostureQuery{})
	cache.start()

	return cache
}

// Set the current list of processes paths that are being observed
func (cache *Cache) setWatchedProcesses(processPaths []string) {
	processMap := map[string]struct{}{}

	for _, processPath := range processPaths {
		processMap[processPath] = struct{}{}
		cache.watchedProcesses.Set(processPath, struct{}{})
	}

	var processesToRemove []string
	cache.watchedProcesses.IterCb(func(processPath string, _ struct{}) {
		if _, ok := processMap[processPath]; !ok {
			processesToRemove = append(processesToRemove, processPath)
		}
	})

	for _, processPath := range processesToRemove {
		cache.watchedProcesses.Remove(processPath)
	}
}

// Evaluate refreshes all posture data and determines if new posture responses should be sent out
func (cache *Cache) Evaluate() {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	cache.Refresh()
	if responses := cache.GetChangedResponses(); len(responses) > 0 {
		if err := cache.SendResponses(responses); err != nil {
			pfxlog.Logger().Error(err)
		}
	}
}

// GetChangedResponses determines if posture responses should be sent out.
func (cache *Cache) GetChangedResponses() []rest_model.PostureResponseCreate {
	if !cache.currentData.Evaluated.CompareAndSwap(false, true) {
		return nil
	}

	activeQueryTypes := map[string]string{} // map[queryType|processPath]->queryId
	cache.activeServices.IterCb(func(serviceId string, _ struct{}) {
		queryMap := cache.serviceQueryMap.Load()[serviceId]

		for queryId, query := range queryMap {
			if *query.QueryType != rest_model.PostureCheckTypePROCESS {
				activeQueryTypes[string(*query.QueryType)] = queryId
			} else {
				activeQueryTypes[query.Process.Path] = queryId
			}
		}
	})

	if len(activeQueryTypes) == 0 {
		return nil
	}

	var responses []rest_model.PostureResponseCreate
	if cache.currentData.Domain != cache.previousData.Domain {
		if queryId, ok := activeQueryTypes[string(rest_model.PostureCheckTypeDOMAIN)]; ok {
			domainResponse := &rest_model.PostureResponseDomainCreate{
				Domain: &cache.currentData.Domain,
			}
			domainResponse.SetID(&queryId)
			domainResponse.SetTypeID(rest_model.PostureCheckTypeDOMAIN)

			responses = append(responses, domainResponse)
		}
	}

	if !stringz.EqualSlices(cache.currentData.MacAddresses, cache.previousData.MacAddresses) {
		if queryId, ok := activeQueryTypes[string(rest_model.PostureCheckTypeMAC)]; ok {
			macResponse := &rest_model.PostureResponseMacAddressCreate{
				MacAddresses: cache.currentData.MacAddresses,
			}
			macResponse.SetID(&queryId)
			macResponse.SetTypeID(rest_model.PostureCheckTypeMAC)

			responses = append(responses, macResponse)
		}
	}

	if cache.previousData.Os.Version != cache.currentData.Os.Version || cache.previousData.Os.Type != cache.currentData.Os.Type {
		if queryId, ok := activeQueryTypes[string(rest_model.PostureCheckTypeOS)]; ok {
			osResponse := &rest_model.PostureResponseOperatingSystemCreate{
				Type:    &cache.currentData.Os.Type,
				Version: &cache.currentData.Os.Version,
				Build:   "",
			}
			osResponse.SetID(&queryId)
			osResponse.SetTypeID(rest_model.PostureCheckTypeOS)

			responses = append(responses, osResponse)
		}
	}

	cache.currentData.Processes.IterCb(func(processPath string, curState ProcessInfo) {
		queryId, isActive := activeQueryTypes[processPath]

		if !isActive {
			return
		}

		prevState, ok := cache.previousData.Processes.Get(processPath)

		sendResponse := false
		if !ok {
			//no prev state send
			sendResponse = true
		} else {
			sendResponse = prevState.IsRunning != curState.IsRunning || prevState.Hash != curState.Hash || !stringz.EqualSlices(prevState.SignerFingerprints, curState.SignerFingerprints)
		}

		if sendResponse {
			processResponse := &rest_model.PostureResponseProcessCreate{
				Path:               processPath,
				Hash:               curState.Hash,
				SignerFingerprints: curState.SignerFingerprints,
				IsRunning:          curState.IsRunning,
			}

			processResponse.SetID(&queryId)
			processResponse.SetTypeID(rest_model.PostureCheckTypePROCESS)
			responses = append(responses, processResponse)
		}
	})

	return responses
}

// Refresh refreshes posture data
func (cache *Cache) Refresh() {
	cache.previousData = cache.currentData

	cache.currentData = NewCacheData()
	cache.currentData.Os = Os()

	cache.currentData.Domain = cache.DomainFunc()
	cache.currentData.MacAddresses = MacAddresses()

	keys := cache.watchedProcesses.Keys()
	for _, processPath := range keys {
		cache.currentData.Processes.Set(processPath, Process(processPath))
	}
}

// SetServiceQueryMap receives of a list of serviceId -> queryId -> queries. Used to determine which queries are necessary
// to provide data for on a per-service basis.
func (cache *Cache) SetServiceQueryMap(serviceQueryMap map[string]map[string]rest_model.PostureQuery) {
	cache.serviceQueryMap.Store(serviceQueryMap)

	var processPaths []string
	for _, queryMap := range serviceQueryMap {
		for _, query := range queryMap {
			if *query.QueryType == rest_model.PostureCheckTypePROCESS && query.Process != nil {
				processPaths = append(processPaths, query.Process.Path)
			}
		}
	}
	cache.setWatchedProcesses(processPaths)
}

func (cache *Cache) AddActiveService(serviceId string) {
	cache.activeServices.Set(serviceId, struct{}{})
	cache.Evaluate()
}

func (cache *Cache) RemoveActiveService(serviceId string) {
	cache.activeServices.Remove(serviceId)
	cache.Evaluate()
}

func (cache *Cache) start() {
	cache.startOnce.Do(func() {
		ticker := time.NewTicker(10 * time.Second)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					pfxlog.Logger().Errorf("error during posture response streaming: %v", r)
				}
			}()

			for {
				select {
				case <-ticker.C:
					cache.Evaluate()
				case <-cache.closeNotify:
					return
				}
			}
		}()
	})
}

func (cache *Cache) SendResponses(responses []rest_model.PostureResponseCreate) []error {
	if cache.doSingleSubmissions {
		var allErrors []error
		for _, response := range responses {
			err := cache.ctrlClient.SendPostureResponse(response)

			if err != nil {
				allErrors = append(allErrors, err)
			}
		}

		return allErrors

	} else {
		err := cache.ctrlClient.SendPostureResponseBulk(responses)

		if apiErr, ok := err.(*runtime.APIError); ok && apiErr.Code == http.StatusNotFound {
			cache.doSingleSubmissions = true
			return cache.SendResponses(responses)
		}
		return []error{err}
	}
}

type Submitter interface {
	SendPostureResponse(response rest_model.PostureResponseCreate) error
	SendPostureResponseBulk(responses []rest_model.PostureResponseCreate) error
}

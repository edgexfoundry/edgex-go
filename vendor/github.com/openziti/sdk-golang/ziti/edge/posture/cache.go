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
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/foundation/v2/stringz"
	"github.com/openziti/sdk-golang/edge-apis"
	"github.com/openziti/sdk-golang/ziti/edge"
	cmap "github.com/orcaman/concurrent-map/v2"
)

const (
	// TotpAttemptDelta defines how far in advance of expiration the cache proactively requests
	// new TOTP tokens, ensuring tokens remain valid during authentication flows.
	TotpAttemptDelta = 5 * time.Minute

	// TotpPostureCheckNoTimeout indicates that a TOTP posture check does not expire and
	// does not require periodic token refresh.
	TotpPostureCheckNoTimeout = int64(-1)
)

// CacheData holds the current snapshot of device posture information including running processes,
// network configuration, operating system details, and authentication state.
type CacheData struct {
	Processes    cmap.ConcurrentMap[string, ProcessInfo] // map[processPath]ProcessInfo
	MacAddresses []string
	Os           OsInfo
	Domain       string
	TotpToken    edge_apis.TotpTokenResult
	OnWake       WakeEvent
	OnUnlock     UnlockEvent
	Index        uint64
	Responses    []rest_model.PostureResponseCreate
}

// NewCacheData creates an empty posture cache snapshot with initialized collections.
func NewCacheData() *CacheData {
	return &CacheData{
		Processes:    cmap.New[ProcessInfo](),
		MacAddresses: []string{},
		Os: OsInfo{
			Type:    "",
			Version: "",
		},
		Domain: "",
		Index:  0,
	}
}

// ActiveServiceProvider supplies information about services currently in use by the client,
// enabling the cache to determine which posture checks are relevant.
type ActiveServiceProvider interface {
	GetActiveDialServices() []*rest_model.ServiceDetail
	GetActiveBindServices() []*rest_model.ServiceDetail
}

// ActiveServiceProviderFunc is a function adapter that implements ActiveServiceProvider
// for both dial and bind service queries.
type ActiveServiceProviderFunc func() []*rest_model.ServiceDetail

func (f ActiveServiceProviderFunc) GetActiveDialServices() []*rest_model.ServiceDetail {
	return f()
}

// Cache manages device posture data collection, tracking changes over time and coordinating
// submission of posture responses when device state changes or policies require updates.
type Cache struct {
	currentData  *CacheData
	previousData *CacheData

	watchedProcesses cmap.ConcurrentMap[string, string] //map[processPath]queryId

	serviceProvider ActiveServiceProvider

	lastSent  cmap.ConcurrentMap[string, time.Time] //map[type|processQueryId]time.Time
	submitter Submitter

	startOnce           sync.Once
	doSingleSubmissions bool
	closeNotify         <-chan struct{}

	DomainProvider    DomainProvider
	MacProvider       MacProvider
	OsProvider        OsProvider
	ProcessProvider   ProcessProvider
	TotpTokenProvider edge_apis.TotpTokenProvider

	lock        sync.Mutex
	totpTimeout int64
	eventState  EventState
}

// NewCache creates a posture cache that monitors device state and coordinates posture response
// submission. The cache uses the provided service provider to determine which posture checks
// are active, the submitter to send responses, and the token provider for TOTP authentication.
func NewCache(activeServiceProvider ActiveServiceProvider, submitter Submitter, totpTokenProvider edge_apis.TotpTokenProvider, closeNotify <-chan struct{}) *Cache {
	cache := &Cache{
		currentData:      NewCacheData(),
		previousData:     NewCacheData(),
		watchedProcesses: cmap.New[string](),
		serviceProvider:  activeServiceProvider,
		lastSent:         cmap.New[time.Time](),
		submitter:        submitter,
		startOnce:        sync.Once{},
		closeNotify:      closeNotify,
		totpTimeout:      TotpPostureCheckNoTimeout,

		TotpTokenProvider: totpTokenProvider,
		DomainProvider:    NewDomainProvider(),
		MacProvider:       NewMacProvider(),
		OsProvider:        NewOsProvider(),
		ProcessProvider:   NewProcessProvider(),

		eventState: NewEventState(),
	}

	cache.currentData.Index = 1

	cache.start()

	return cache
}

// Evaluate refreshes all posture data and determines if new posture responses should be sent out
func (cache *Cache) Evaluate() {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	activeDialServices := cache.serviceProvider.GetActiveDialServices()
	activeBindServices := cache.serviceProvider.GetActiveBindServices()

	activeQueryTypes, activeProcesses, lowestTotpTimeout := getActiveQueryInfo(activeDialServices, activeBindServices)

	cache.setTotpTimeout(lowestTotpTimeout)

	candidateData := NewCacheData()
	candidateData.Index = cache.currentData.Index + 1

	candidateData.Os = cache.OsProvider.GetOsInfo()
	candidateData.Domain = cache.DomainProvider.GetDomain()
	candidateData.MacAddresses = sanitizeMacAddresses(cache.MacProvider.GetMacAddresses())

	for processPath, queryId := range activeProcesses {
		processInfo := cache.ProcessProvider.GetProcessInfo(processPath)
		processInfo.QueryId = queryId
		candidateData.Processes.Set(processPath, processInfo)
	}

	candidateData.TotpToken = cache.previousData.TotpToken

	if cache.TotpTokenProvider != nil && cache.totpTimeoutWindowEntered() {
		totpTokenResultCh := cache.TotpTokenProvider.Request()

		if totpTokenResultCh != nil {
			select {
			case totpTokenResult := <-totpTokenResultCh:
				if totpTokenResult.Err != nil {
					pfxlog.Logger().Errorf("error requesting totp token: %v", totpTokenResult.Err)
				} else {
					candidateData.TotpToken = totpTokenResult
				}
			case <-cache.closeNotify:
				return
			}
		}
	}

	responses := cache.GetChangedResponses(cache.currentData, candidateData, activeQueryTypes)
	if len(responses) > 0 {
		cache.previousData = cache.currentData
		cache.currentData = candidateData
		cache.currentData.Responses = responses

		if err := cache.SendResponses(responses); err != nil {
			pfxlog.Logger().Error(err)
		}
	}
}

func getActiveQueryInfo(dialServices []*rest_model.ServiceDetail, bindServices []*rest_model.ServiceDetail) (map[string]string, map[string]string, int64) {
	activeQueryTypes := map[string]string{} // map[queryType]->queryId'
	activeProcesses := map[string]string{}  // map[processPath]->queryId'

	lowestTotpTimeout := TotpPostureCheckNoTimeout
	for _, service := range dialServices {
		for _, postureQueryState := range service.PostureQueries {
			if postureQueryState.PolicyType == rest_model.DialBindDial {
				addQueryInfoToMaps(activeQueryTypes, activeProcesses, postureQueryState)
				lowestTotpTimeout = getLowestTotpTimeout(postureQueryState, lowestTotpTimeout)
			}
		}
	}

	for _, service := range bindServices {
		for _, postureQueryState := range service.PostureQueries {
			if postureQueryState.PolicyType == rest_model.DialBindBind {
				addQueryInfoToMaps(activeQueryTypes, activeProcesses, postureQueryState)
				lowestTotpTimeout = getLowestTotpTimeout(postureQueryState, lowestTotpTimeout)
			}
		}
	}

	return activeQueryTypes, activeProcesses, lowestTotpTimeout
}

func getLowestTotpTimeout(postureQueryState *rest_model.PostureQueries, curTimeout int64) int64 {
	for _, query := range postureQueryState.PostureQueries {

		if *query.QueryType == rest_model.PostureCheckTypeMFA {
			if curTimeout == TotpPostureCheckNoTimeout && query.Timeout != nil && *query.Timeout != TotpPostureCheckNoTimeout {
				curTimeout = *query.Timeout
			} else if query.Timeout != nil && *query.Timeout < curTimeout {
				curTimeout = *query.Timeout
			}
		}
	}

	return curTimeout
}

func addQueryInfoToMaps(activeQueryTypes map[string]string, activeProcesses map[string]string, postureQueryState *rest_model.PostureQueries) {
	for _, query := range postureQueryState.PostureQueries {
		activeQueryTypes[string(*query.QueryType)] = *query.ID

		switch *query.QueryType {
		case rest_model.PostureCheckTypePROCESS:
			activeQueryTypes[string(rest_model.PostureCheckTypeOS)] = *query.ID
			activeProcesses[query.Process.Path] = *query.ID
		case rest_model.PostureCheckTypePROCESSMULTI:
			activeQueryTypes[string(rest_model.PostureCheckTypeOS)] = *query.ID

			for _, process := range query.Processes {
				activeProcesses[process.Path] = *query.ID
			}
		}
	}
}

// GetChangedResponses determines if posture responses should be sent out.
func (cache *Cache) GetChangedResponses(currentData, candidateData *CacheData, activeQueryTypes map[string]string) []rest_model.PostureResponseCreate {
	if len(activeQueryTypes) == 0 && candidateData.Processes.Count() == 0 {
		return nil
	}

	var responses []rest_model.PostureResponseCreate

	wakeChanged := !currentData.OnWake.At.Equal(candidateData.OnWake.At)
	unlockChanged := !currentData.OnUnlock.At.Equal(candidateData.OnUnlock.At)

	if wakeChanged || unlockChanged {
		// TOTP MFA checks are the only checks that care about wake/unlock
		if queryId, ok := activeQueryTypes[string(rest_model.PostureCheckTypeMFA)]; ok {
			endpointState := &rest_model.PostureResponseEndpointStateCreate{
				Unlocked: unlockChanged,
				Woken:    wakeChanged,
			}
			endpointState.SetID(&queryId)
			responses = append(responses, endpointState)
		}
	}

	if currentData.Domain != candidateData.Domain {
		if queryId, ok := activeQueryTypes[string(rest_model.PostureCheckTypeDOMAIN)]; ok {
			domainResponse := &rest_model.PostureResponseDomainCreate{
				Domain: &candidateData.Domain,
			}
			domainResponse.SetID(&queryId)
			domainResponse.SetTypeID(rest_model.PostureCheckTypeDOMAIN)

			responses = append(responses, domainResponse)
		}
	}

	if currentData.TotpToken.Token != candidateData.TotpToken.Token {
		if queryId, ok := activeQueryTypes[string(rest_model.PostureCheckTypeMFA)]; ok {
			totpMfaResponse := &edge.PostureResponseTotp{
				TotpToken: candidateData.TotpToken.Token,
			}
			totpMfaResponse.SetID(&queryId)
			totpMfaResponse.SetTypeID(rest_model.PostureCheckTypeMFA)

			responses = append(responses, totpMfaResponse)
		}
	}

	if !stringz.EqualSlices(currentData.MacAddresses, candidateData.MacAddresses) {
		if queryId, ok := activeQueryTypes[string(rest_model.PostureCheckTypeMAC)]; ok {
			macResponse := &rest_model.PostureResponseMacAddressCreate{
				MacAddresses: candidateData.MacAddresses,
			}
			macResponse.SetID(&queryId)
			macResponse.SetTypeID(rest_model.PostureCheckTypeMAC)

			responses = append(responses, macResponse)
		}
	}

	if candidateData.Os.Version != currentData.Os.Version || candidateData.Os.Type != currentData.Os.Type {
		if queryId, ok := activeQueryTypes[string(rest_model.PostureCheckTypeOS)]; ok {
			osResponse := &rest_model.PostureResponseOperatingSystemCreate{
				Type:    &candidateData.Os.Type,
				Version: &candidateData.Os.Version,
				Build:   "",
			}
			osResponse.SetID(&queryId)
			osResponse.SetTypeID(rest_model.PostureCheckTypeOS)

			responses = append(responses, osResponse)
		}
	}

	candidateData.Processes.IterCb(func(processPath string, candidateProcessInfo ProcessInfo) {
		curProcessInfo, ok := currentData.Processes.Get(processPath)

		sendResponse := false
		if !ok {
			//no prev state send
			sendResponse = true
		} else {
			sendResponse = curProcessInfo.IsRunning != candidateProcessInfo.IsRunning || curProcessInfo.Hash != candidateProcessInfo.Hash || !stringz.EqualSlices(curProcessInfo.SignerFingerprints, candidateProcessInfo.SignerFingerprints)
		}

		if sendResponse {
			processResponse := &rest_model.PostureResponseProcessCreate{
				Path:               processPath,
				Hash:               candidateProcessInfo.Hash,
				SignerFingerprints: candidateProcessInfo.SignerFingerprints,
				IsRunning:          candidateProcessInfo.IsRunning,
			}

			processResponse.SetID(&candidateProcessInfo.QueryId)
			processResponse.SetTypeID(rest_model.PostureCheckTypePROCESS)
			responses = append(responses, processResponse)
		}
	})

	return responses
}

func (cache *Cache) totpTimeoutWindowEntered() bool {
	if cache.totpTimeout == TotpPostureCheckNoTimeout {
		return false
	}

	if cache.previousData.TotpToken.IssuedAt.IsZero() {
		return true
	}

	if cache.previousData.TotpToken.Token == "" {
		return true
	}

	effectiveTimeout := time.Duration(cache.totpTimeout)*time.Second - TotpAttemptDelta
	return cache.previousData.TotpToken.IssuedAt.Add(effectiveTimeout).Before(time.Now())
}

func (cache *Cache) start() {
	cache.startOnce.Do(func() {
		stopWake, err := cache.eventState.ListenForWake(cache.onWake)
		if err != nil {
			pfxlog.Logger().WithError(err).Error("error starting wake listener for posture")
		}

		stopUnlock, err := cache.eventState.ListenForUnlock(cache.OnUnlock)
		if err != nil {
			pfxlog.Logger().WithError(err).Error("error starting unlock listener for posture")
		}

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
					if stopWake != nil {
						stopWake()
					}

					if stopUnlock != nil {
						stopUnlock()
					}
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
			err := cache.submitter.SendPostureResponse(response)

			if err != nil {
				allErrors = append(allErrors, err)
			}
		}

		return allErrors

	} else {
		err := cache.submitter.SendPostureResponseBulk(responses)

		if err != nil {
			if apiErr, ok := err.(*runtime.APIError); ok && apiErr.Code == http.StatusNotFound {
				cache.doSingleSubmissions = true
				return cache.SendResponses(responses)
			}
			return []error{err}
		}
		return nil
	}
}

func (cache *Cache) InitializePostureOnEdgeRouter(conn edge.RouterConn) error {
	allResponses := cache.GetAllResponses()
	return cache.submitter.SendPostureResponseBulk(allResponses)
}

func (cache *Cache) GetAllResponses() []rest_model.PostureResponseCreate {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	var allResponses []rest_model.PostureResponseCreate

	allResponses = append(allResponses, cache.currentData.Responses...)

	return allResponses
}

func (cache *Cache) setTotpTimeout(timeout int64) {
	cache.totpTimeout = timeout
}

func (cache *Cache) onWake(event WakeEvent) {
	cache.currentData.OnWake = event
}

func (cache *Cache) OnUnlock(event UnlockEvent) {
	cache.currentData.OnUnlock = event
}

func (cache *Cache) SimulateWake() {
	cache.onWake(WakeEvent{At: time.Now().UTC()})
}

func (cache *Cache) SimulateUnlock() {
	cache.OnUnlock(UnlockEvent{At: time.Now().UTC()})
}

func (cache *Cache) SetTotpToken(token *rest_model.TotpToken) {
	cache.currentData.TotpToken = edge_apis.TotpTokenResult{
		Token:    *token.Token,
		IssuedAt: time.Time(*token.IssuedAt),
	}
}

func (cache *Cache) SetTotpProviderFunc(f func() <-chan edge_apis.TotpTokenResult) {
	p := edge_apis.TotpTokenProviderFunc(f)
	cache.TotpTokenProvider = &p
}

func (cache *Cache) SetDomainProviderFunc(f func() string) {
	p := DomainProviderFunc(f)
	cache.DomainProvider = &p
}

func (cache *Cache) SetMacProviderFunc(f func() []string) {
	p := MacProviderFunc(f)
	cache.MacProvider = &p
}

func (cache *Cache) SetOsProviderFunc(f func() OsInfo) {
	p := OsProviderFunc(f)
	cache.OsProvider = &p
}

func (cache *Cache) SetProcessProviderFunc(f func(string) ProcessInfo) {
	p := ProcessInfoFunc(f)
	cache.ProcessProvider = &p
}

func sanitizeMacAddresses(addresses []string) []string {
	result := make([]string, 0, len(addresses))

	for _, address := range addresses {
		address = strings.TrimSpace(address)
		address = strings.ToLower(address)
		address = strings.ReplaceAll(address, ":", "")
		result = append(result, address)
	}

	return result
}

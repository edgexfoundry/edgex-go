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

package ziti

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/kataras/go-events"
	"github.com/openziti/edge-api/rest_client_api_client/authentication"
	"github.com/openziti/edge-api/rest_client_api_client/service"
	rest_session "github.com/openziti/edge-api/rest_client_api_client/session"
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/openziti/foundation/v2/errorz"
	"github.com/openziti/foundation/v2/stringz"
	apis "github.com/openziti/sdk-golang/edge-apis"
	"github.com/openziti/secretstream/kx"
	"math"
	"math/rand"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/channel/v3/latency"
	"github.com/openziti/edge-api/rest_client_api_client/current_api_session"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/foundation/v2/versions"
	"github.com/openziti/identity"
	"github.com/openziti/metrics"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/sdk-golang/ziti/edge/network"
	"github.com/openziti/sdk-golang/ziti/signing"
	"github.com/openziti/transport/v2"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
	metrics2 "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

type SessionType rest_model.DialBind

const (
	LatencyCheckInterval = 30 * time.Second
	LatencyCheckTimeout  = 10 * time.Second

	ClientConfigV1 = "ziti-tunneler-client.v1"
	InterceptV1    = "intercept.v1"

	SessionDial = rest_model.DialBindDial
	SessionBind = rest_model.DialBindBind
)

// MfaCodeResponse is a handler used to return a string (TOTP) code
type MfaCodeResponse func(code string) error

// Context is the main interface for SDK instances that may be used to authenticate, connect to services, or host
// services.
type Context interface {
	// Authenticate attempts to use credentials configured on the Context to perform authentication. The authentication
	// implementation used is configured via the Credentials field on an Option struct provided during Context
	// creation.
	Authenticate() error

	// SetCredentials sets the credentials used to authenticate against the Edge Client API.
	SetCredentials(authenticator apis.Credentials)

	// GetCredentials returns the currently set credentials used to authenticate against the Edge Client API.
	GetCredentials() apis.Credentials

	// GetCurrentIdentity returns the Edge API details of the currently authenticated identity.
	GetCurrentIdentity() (*rest_model.IdentityDetail, error)

	// GetCurrentIdentityWithBackoff returns the Edge API details of the currently authenticated identity. with retry if necessary
	GetCurrentIdentityWithBackoff() (*rest_model.IdentityDetail, error)

	// Dial attempts to connect to a service using a given service name; authenticating as necessary in order to obtain
	// a service session, attach to Edge Routers, and connect to a service.
	Dial(serviceName string) (edge.Conn, error)

	// DialWithOptions performs the same logic as Dial but allows specification of DialOptions.
	DialWithOptions(serviceName string, options *DialOptions) (edge.Conn, error)

	// DialAddr finds the service for given address and performs a Dial for it.
	DialAddr(network string, addr string) (edge.Conn, error)

	// Listen attempts to host a service by the given service name;  authenticating as necessary in order to obtain
	// a service session, attach to Edge Routers, and bind (host) the service.
	Listen(serviceName string) (edge.Listener, error)

	// ListenWithOptions performs the same logic as Listen, but allows the specification of ListenOptions.
	ListenWithOptions(serviceName string, options *ListenOptions) (edge.Listener, error)

	// GetServiceId will return the id of a specific service by service name. If not found, false, will be returned
	// with an empty string.
	GetServiceId(serviceName string) (string, bool, error)

	// GetServices will return a slice of service details that the current authenticating identity can access for
	// dial (connect) or bind (host/listen).
	GetServices() ([]rest_model.ServiceDetail, error)

	// GetService will return the service details of a specific service by service name.
	GetService(serviceName string) (*rest_model.ServiceDetail, bool)

	// GetServiceForAddr finds the service with intercept that matches best to given address
	GetServiceForAddr(network, hostname string, port uint16) (*rest_model.ServiceDetail, int, error)

	// RefreshServices forces the context to refresh the list of services the current authenticating identity has access
	// to.
	RefreshServices() error

	// RefreshService forces the context to refresh just the service with the given name. If the given service isn't
	// found, a nil will be returned
	RefreshService(serviceName string) (*rest_model.ServiceDetail, error)

	// GetServiceTerminators will return a slice of rest_model.TerminatorClientDetail for a specific service name.
	// The offset and limit options can be used to page through excessive lists of items. A max of 500 is imposed on
	// limit.
	GetServiceTerminators(serviceName string, offset, limit int) ([]*rest_model.TerminatorClientDetail, int, error)

	// GetSession will return the session detail associated with a specific session id.
	GetSession(id string) (*rest_model.SessionDetail, error)

	// Metrics will return the current context's metrics Registry.
	Metrics() metrics.Registry

	// Close closes any connections open to edge routers
	Close()

	// Deprecated: AddZitiMfaHandler adds a Ziti MFA handler, invoked during authentication.
	// Replaced with event functionality. Use `zitiContext.Events().AddMfaTotpCodeListener(func(Context, *rest_model.AuthQueryDetail, MfaCodeResponse))` instead.
	AddZitiMfaHandler(handler func(query *rest_model.AuthQueryDetail, resp MfaCodeResponse) error)

	// EnrollZitiMfa will attempt to enable TOTP 2FA on the currently authenticating identity if not already enrolled.
	EnrollZitiMfa() (*rest_model.DetailMfa, error)

	// VerifyZitiMfa will attempt to complete enrollment of TOTP 2FA with the given code.
	VerifyZitiMfa(code string) error

	// RemoveZitiMfa will attempt to remove TOTP 2FA for the current identity
	RemoveZitiMfa(code string) error

	// GetId returns a unique context id
	GetId() string

	// SetId allows the setting of a context's id
	SetId(id string)

	Events() Eventer
}

var _ Context = &ContextImpl{}

type ContextImpl struct {
	options           *Options
	Id                string
	routerConnections cmap.ConcurrentMap[string, edge.RouterConn]

	CtrlClt *CtrlClient

	services   cmap.ConcurrentMap[string, *rest_model.ServiceDetail] // name -> Service
	sessions   cmap.ConcurrentMap[string, *rest_model.SessionDetail] // svcID:type -> Session
	intercepts cmap.ConcurrentMap[string, *edge.InterceptV1Config]

	metrics metrics.Registry

	firstAuthOnce sync.Once

	closed            atomic.Bool
	closeNotify       chan struct{}
	authQueryHandlers map[string]func(query *rest_model.AuthQueryDetail, response MfaCodeResponse) error

	events.EventEmmiter
	lastSuccessfulApiSessionRefresh time.Time
	routerProxy                     func(addr string) *transport.ProxyConfiguration
}

func (context *ContextImpl) AddServiceAddedListener(handler func(Context, *rest_model.ServiceDetail)) func() {
	listener := func(args ...interface{}) {
		details, ok := args[0].(*rest_model.ServiceDetail)

		if !ok {
			pfxlog.Logger().Fatalf("could not convert args[0] to %T was %T", details, args[0])
		}

		if details == nil {
			pfxlog.Logger().Fatalf("expected arg[0] was nil, unexpected")
		}

		handler(context, details)
	}

	context.AddListener(EventServiceAdded, listener)

	return func() {
		context.RemoveListener(EventServiceAdded, listener)
	}
}

func (context *ContextImpl) AddServiceChangedListener(handler func(Context, *rest_model.ServiceDetail)) func() {
	listener := func(args ...interface{}) {
		details, ok := args[0].(*rest_model.ServiceDetail)

		if !ok {
			pfxlog.Logger().Fatalf("could not convert args[0] to %T was %T", details, args[0])
		}

		if details == nil {
			pfxlog.Logger().Fatalf("expected arg[0] was nil, unexpected")
		}

		handler(context, details)
	}

	context.AddListener(EventServiceChanged, listener)

	return func() {
		context.RemoveListener(EventServiceChanged, listener)
	}
}

func (context *ContextImpl) AddServiceRemovedListener(handler func(Context, *rest_model.ServiceDetail)) func() {
	listener := func(args ...interface{}) {
		details, ok := args[0].(*rest_model.ServiceDetail)

		if !ok {
			pfxlog.Logger().Fatalf("could not convert args[0] to %T was %T", details, args[0])
		}

		if details == nil {
			pfxlog.Logger().Fatalf("expected arg[0] was nil, unexpected")
		}

		handler(context, details)
	}

	context.AddListener(EventServiceRemoved, listener)

	return func() {
		context.RemoveListener(EventServiceRemoved, listener)
	}
}

func (context *ContextImpl) AddRouterConnectedListener(handler func(Context, string, string)) func() {
	listener := func(args ...interface{}) {
		name, ok := args[0].(string)

		if !ok {
			pfxlog.Logger().Fatalf("could not convert args[0] to %T was %T", name, args[0])
		}

		addr, ok := args[1].(string)

		if !ok {
			pfxlog.Logger().Fatalf("could not convert args[1] to %T was %T", addr, args[1])
		}

		handler(context, name, addr)
	}

	context.AddListener(EventRouterConnected, listener)

	return func() {
		context.RemoveListener(EventRouterConnected, listener)
	}
}

func (context *ContextImpl) AddRouterDisconnectedListener(handler func(Context, string, string)) func() {
	listener := func(args ...interface{}) {
		name, ok := args[0].(string)

		if !ok {
			pfxlog.Logger().Fatalf("could not convert args[0] to %T was %T", name, args[0])
		}

		addr, ok := args[1].(string)

		if !ok {
			pfxlog.Logger().Fatalf("could not convert args[1] to %T was %T", addr, args[1])
		}

		handler(context, name, addr)
	}

	context.AddListener(EventRouterDisconnected, listener)

	return func() {
		context.RemoveListener(EventRouterDisconnected, listener)
	}
}

func (context *ContextImpl) AddMfaTotpCodeListener(handler func(Context, *rest_model.AuthQueryDetail, MfaCodeResponse)) func() {
	listener := func(args ...interface{}) {
		authQuery, ok := args[0].(*rest_model.AuthQueryDetail)

		if !ok {
			pfxlog.Logger().Fatalf("could not convert args[0] to %T was %T", authQuery, args[0])
		}

		if authQuery == nil {
			pfxlog.Logger().Fatalf("expected arg[0] was nil, unexpected")
		}

		responder, ok := args[1].(MfaCodeResponse)

		if !ok {
			pfxlog.Logger().Fatalf("could not convert args[1] to %T was %T", responder, args[1])
		}

		if responder == nil {
			pfxlog.Logger().Fatalf("expected arg[0] was nil, unexpected")
		}

		handler(context, authQuery, responder)
	}

	context.AddListener(EventMfaTotpCode, listener)

	return func() {
		context.RemoveListener(EventMfaTotpCode, listener)
	}
}

func (context *ContextImpl) AddAuthQueryListener(handler func(Context, *rest_model.AuthQueryDetail)) func() {
	listener := func(args ...interface{}) {
		authQuery, ok := args[0].(*rest_model.AuthQueryDetail)

		if !ok {
			pfxlog.Logger().Fatalf("could not convert args[0] to %T was %T", authQuery, args[0])
		}

		if authQuery == nil {
			pfxlog.Logger().Fatalf("expected arg[0] was nil, unexpected")
		}

		handler(context, authQuery)
	}

	context.AddListener(EventAuthQuery, listener)

	return func() {
		context.RemoveListener(EventAuthQuery, listener)
	}
}

func (context *ContextImpl) AddAuthenticationStatePartialListener(handler func(Context, apis.ApiSession)) func() {
	listener := func(args ...interface{}) {
		apiSession, ok := args[0].(apis.ApiSession)

		if !ok {
			pfxlog.Logger().Fatalf("could not convert args[0] to %T was %T", apiSession, args[0])
		}

		if apiSession == nil {
			pfxlog.Logger().Fatalf("expected arg[0] was nil, unexpected")
		}

		handler(context, apiSession)
	}

	context.AddListener(EventAuthenticationStatePartial, listener)

	return func() {
		context.RemoveListener(EventAuthenticationStatePartial, listener)
	}
}

func (context *ContextImpl) AddAuthenticationStateFullListener(handler func(Context, apis.ApiSession)) func() {
	listener := func(args ...interface{}) {
		apiSession, ok := args[0].(apis.ApiSession)

		if !ok {
			pfxlog.Logger().Fatalf("could not convert args[0] to %T was %T", apiSession, args[0])
		}

		if apiSession == nil {
			pfxlog.Logger().Fatalf("expected arg[0] was nil, unexpected")
		}

		handler(context, apiSession)
	}

	context.AddListener(EventAuthenticationStateFull, listener)

	return func() {
		context.RemoveListener(EventAuthenticationStateFull, listener)
	}
}

func (context *ContextImpl) AddAuthenticationStateUnauthenticatedListener(handler func(Context, apis.ApiSession)) func() {
	listener := func(args ...interface{}) {
		var apiSession apis.ApiSession

		if args[0] != nil {
			var ok bool
			apiSession, ok = args[0].(apis.ApiSession)

			if !ok {
				pfxlog.Logger().Fatalf("could not convert args[0] to %T was %T", apiSession, args[0])
			}
		}

		handler(context, apiSession)
	}

	context.AddListener(EventAuthenticationStateUnauthenticated, listener)

	return func() {
		context.RemoveListener(EventAuthenticationStateUnauthenticated, listener)
	}
}

func (context *ContextImpl) AddControllerUrlsUpdateListener(handler func(Context, []*url.URL)) func() {
	listener := func(args ...interface{}) {
		var apiUrls []*url.URL

		if args[0] != nil {
			var ok bool
			apiUrls, ok = args[0].([]*url.URL)

			if !ok {
				pfxlog.Logger().Fatalf("could not convert args[0] to %T was %T", apiUrls, args[0])
			}
		}

		handler(context, apiUrls)
	}

	context.AddListener(EventControllerUrlsUpdated, listener)

	return func() {
		context.RemoveListener(EventAuthenticationStateUnauthenticated, listener)
	}
}

func (context *ContextImpl) Events() Eventer {
	return context
}

func (context *ContextImpl) GetId() string {
	return context.Id
}

func (context *ContextImpl) SetId(id string) {
	context.Id = id
}

func (context *ContextImpl) SetCredentials(credentials apis.Credentials) {
	context.CtrlClt.Credentials = credentials
}

func (context *ContextImpl) GetCredentials() apis.Credentials {
	return context.CtrlClt.Credentials
}

func (context *ContextImpl) Sessions() ([]*rest_model.SessionDetail, error) {
	var sessions []*rest_model.SessionDetail
	context.sessions.IterCb(func(key string, s *rest_model.SessionDetail) {
		sessions = append(sessions, s)
	})
	return sessions, nil
}

func (context *ContextImpl) OnClose(routerConn edge.RouterConn) {
	logrus.Debugf("connection to router [%s] was closed", routerConn.Key())
	context.Emit(EventRouterDisconnected, routerConn.GetRouterName(), routerConn.Key())
	context.routerConnections.Remove(routerConn.Key())
}

func (context *ContextImpl) processServiceUpdates(services []*rest_model.ServiceDetail) {
	pfxlog.Logger().Debugf("processing service updates with %v services", len(services))

	idMap := make(map[string]*rest_model.ServiceDetail)
	for _, s := range services {
		idMap[*s.ID] = s
	}

	// process Deletes
	var deletes []string
	context.services.IterCb(func(key string, svc *rest_model.ServiceDetail) {
		if _, found := idMap[*svc.ID]; !found {
			deletes = append(deletes, key)
			if context.options.OnServiceUpdate != nil {
				context.options.OnServiceUpdate(ServiceRemoved, svc)
			}
			context.Emit(EventServiceRemoved, svc)

			context.deleteServiceSessions(*svc.ID)

		}
	})

	for _, deletedKey := range deletes {
		context.services.Remove(deletedKey)
		context.intercepts.Remove(deletedKey)
	}

	// Adds and Updates
	for _, s := range services {
		context.processServiceAddOrUpdated(s)
	}

	context.refreshServiceQueryMap()
}

func (context *ContextImpl) processSingleServiceUpdate(name string, s *rest_model.ServiceDetail) {
	// process Deletes
	if s == nil {
		var deletes []string
		context.services.IterCb(func(key string, svc *rest_model.ServiceDetail) {
			if *svc.Name == name {
				deletes = append(deletes, key)
				if context.options.OnServiceUpdate != nil {
					context.options.OnServiceUpdate(ServiceRemoved, svc)
				}
				context.Emit(EventServiceRemoved, svc)
				context.deleteServiceSessions(*svc.ID)
			}
		})

		for _, deletedKey := range deletes {
			context.services.Remove(deletedKey)
			context.intercepts.Remove(deletedKey)
		}
	} else {
		// Adds and Updates
		context.processServiceAddOrUpdated(s)
	}

	context.refreshServiceQueryMap()
}

func (context *ContextImpl) processServiceAddOrUpdated(s *rest_model.ServiceDetail) {
	isChange := false
	valuesDiffer := false

	_ = context.services.Upsert(*s.Name, s, func(exist bool, valueInMap *rest_model.ServiceDetail, newValue *rest_model.ServiceDetail) *rest_model.ServiceDetail {
		isChange = exist
		if isChange {
			valuesDiffer = !reflect.DeepEqual(newValue, valueInMap)
		}

		return newValue
	})

	if isChange {
		context.Emit(EventServiceChanged, s)
	} else {
		context.Emit(EventServiceAdded, s)
	}

	if context.options.OnServiceUpdate != nil {
		if isChange {
			if valuesDiffer {
				context.options.OnServiceUpdate(ServiceChanged, s)
			}
		} else {
			context.services.Set(*s.Name, s)
			context.options.OnServiceUpdate(ServiceAdded, s)
		}
	}

	intercept := &edge.InterceptV1Config{}
	ok, err := edge.ParseServiceConfig(s, InterceptV1, intercept)
	if err != nil {
		pfxlog.Logger().Warnf("failed to parse config[%s] for service[%s]", InterceptV1, *s.Name)
	} else if ok {
		intercept.Service = s
		context.intercepts.Set(*s.Name, intercept)
	} else {
		cltCfg := &edge.ClientConfig{}
		ok, err := edge.ParseServiceConfig(s, ClientConfigV1, cltCfg)
		if err == nil && ok {
			intercept = cltCfg.ToInterceptV1Config()
			intercept.Service = s
			context.intercepts.Set(*s.Name, intercept)
		}
	}
}

func (context *ContextImpl) refreshServiceQueryMap() {
	serviceQueryMap := map[string]map[string]rest_model.PostureQuery{} //serviceId -> queryId -> query

	context.services.IterCb(func(key string, svc *rest_model.ServiceDetail) {
		for _, querySets := range svc.PostureQueries {
			for _, query := range querySets.PostureQueries {
				var queryMap map[string]rest_model.PostureQuery
				var ok bool
				if queryMap, ok = serviceQueryMap[*svc.ID]; !ok {
					queryMap = map[string]rest_model.PostureQuery{}
					serviceQueryMap[*svc.ID] = queryMap
				}
				queryMap[*query.ID] = *query
			}
		}
	})

	context.CtrlClt.PostureCache.SetServiceQueryMap(serviceQueryMap)
}

func (context *ContextImpl) refreshSessions() {
	log := pfxlog.Logger()
	edgeRouters := make(map[string]string)
	var toDelete []string
	for entry := range context.sessions.IterBuffered() {
		key := entry.Key
		session := entry.Val
		log.Debugf("refreshing session for %s", key)

		if s, err := context.refreshSession(session); err != nil {
			log.WithError(err).Errorf("failed to refresh session for %s", key)
			toDelete = append(toDelete, *session.ID)
		} else {
			for _, er := range s.EdgeRouters {
				for _, u := range er.SupportedProtocols {
					if context.options.isEdgeRouterUrlAccepted(u) {
						edgeRouters[u] = *er.Name
					}
				}
			}
		}
	}

	for _, id := range toDelete {
		context.sessions.Remove(id)
	}

	for u, name := range edgeRouters {
		go context.handleConnectEdgeRouter(name, u, nil)
	}
}

func (context *ContextImpl) RefreshServices() error {
	return context.refreshServices(true)
}

func (context *ContextImpl) refreshServices(forceCheck bool) error {
	if err := context.ensureApiSession(); err != nil {
		return fmt.Errorf("failed to refresh services: %v", err)
	}

	var checkService bool
	var lastServiceUpdate *strfmt.DateTime
	var err error

	log := pfxlog.Logger()
	log.Debug("checking if service updates available")
	if checkService, lastServiceUpdate, err = context.CtrlClt.IsServiceListUpdateAvailable(); err != nil {
		log.WithError(err).Error("failed to check if service list update is available")
		target := &current_api_session.ListServiceUpdatesUnauthorized{}
		if errors.As(err, &target) {
			checkService = true
		} else {
			if err = context.Authenticate(); err != nil {
				log.WithError(err).Error("unable to re-authenticate during session refresh")
			} else {
				if checkService, lastServiceUpdate, err = context.CtrlClt.IsServiceListUpdateAvailable(); err != nil {
					checkService = true
				}
			}
		}
	}

	if checkService || forceCheck {
		log.Debug("refreshing services")

		services, err := context.CtrlClt.GetServices()
		if err != nil {
			target := &service.ListServicesUnauthorized{}
			if errors.As(err, &target) {
				log.Info("attempting to re-authenticate")
				if authErr := context.Authenticate(); authErr != nil {
					log.WithError(authErr).Error("unable to re-authenticate during services refresh")
					return err
				}
				if services, err = context.CtrlClt.GetServices(); err != nil {
					return err
				}

			} else {
				return err
			}
		}
		context.CtrlClt.lastServiceUpdate = lastServiceUpdate
		context.processServiceUpdates(services)
	}

	return nil
}

func (context *ContextImpl) RefreshService(serviceName string) (*rest_model.ServiceDetail, error) {
	if err := context.ensureApiSession(); err != nil {
		return nil, fmt.Errorf("failed to refresh service: %v", err)
	}

	var err error

	log := pfxlog.Logger().WithField("serviceName", serviceName)

	log.Debug("refreshing service")

	serviceDetail, err := context.CtrlClt.GetService(serviceName)
	if err != nil {
		target := &service.ListServicesUnauthorized{}
		if errors.As(err, &target) {
			log.Info("attempting to re-authenticate")
			if authErr := context.Authenticate(); authErr != nil {
				log.WithError(authErr).Error("unable to re-authenticate during service refresh")
				return nil, err
			}
			if serviceDetail, err = context.CtrlClt.GetService(serviceName); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	context.processSingleServiceUpdate(serviceName, serviceDetail)

	return serviceDetail, nil
}

func (context *ContextImpl) updateTokenOnAllErs(apiSession apis.ApiSession) {
	if apiSession.RequiresRouterTokenUpdate() {
		for tpl := range context.routerConnections.IterBuffered() {
			erConn := tpl.Val
			erKey := tpl.Key
			go func() {
				if err := erConn.UpdateToken(apiSession.GetToken(), 10*time.Second); err != nil {
					pfxlog.Logger().WithError(err).WithField("er", erKey).Warn("error updating apiSession token to connected ER")
				}
			}()
		}
	}
}

func (context *ContextImpl) runRefreshes() {
	log := pfxlog.Logger()
	svcRefreshInterval := context.options.RefreshInterval

	if svcRefreshInterval == 0 {
		svcRefreshInterval = DefaultServiceRefreshInterval
	}
	if svcRefreshInterval < MinRefreshInterval {
		svcRefreshInterval = MinRefreshInterval
	}
	svcRefreshTick := time.NewTicker(svcRefreshInterval)
	defer svcRefreshTick.Stop()

	sessionRefreshInterval := context.options.SessionRefreshInterval
	if sessionRefreshInterval == 0 {
		sessionRefreshInterval = DefaultSessionRefreshInterval
	}
	if sessionRefreshInterval < MinRefreshInterval {
		sessionRefreshInterval = MinRefreshInterval
	}

	sessionRefreshTick := time.NewTicker(sessionRefreshInterval)
	defer sessionRefreshTick.Stop()

	refreshAt := time.Now().Add(30 * time.Second)

	if currentApiSession := context.CtrlClt.GetCurrentApiSession(); currentApiSession != nil && currentApiSession.GetExpiresAt() != nil {
		refreshAt = (*currentApiSession.GetExpiresAt()).Add(-10 * time.Second)
	}

	for {
		select {
		case <-context.closeNotify:
			return

		case <-time.After(time.Until(refreshAt)):
			apiSession := context.CtrlClt.GetCurrentApiSession()

			if apiSession == nil {
				pfxlog.Logger().Warn("could not refresh api session, current api session is nil, authenticating")
				if err := context.Authenticate(); err != nil {
					pfxlog.Logger().WithError(err).Error("failed to authenticate")
				}
				refreshAt = time.Now().Add(5 * time.Second)
				continue
			}

			newApiSession, err := context.CtrlClt.Refresh()

			if err != nil {
				log.Errorf("could not refresh apiSession: %v", err)

				refreshAt = time.Now().Add(5 * time.Second)
			} else {
				exp := newApiSession.GetExpiresAt()
				refreshAt = exp.Add(-10 * time.Second)
				log.Debugf("apiSession refreshed, new expiration[%s]", *exp)

				context.updateTokenOnAllErs(newApiSession)
			}

		case <-svcRefreshTick.C:
			log.Debug("refreshing services")
			if err := context.refreshServices(false); err != nil {
				log.WithError(err).Error("failed to load service updates")
			}

		case <-sessionRefreshTick.C:
			log.Debug("refreshing sessions")
			context.refreshSessions()
		}
	}
}

func (context *ContextImpl) EnsureAuthenticated(options edge.ConnOptions) error {
	operation := func() error {
		pfxlog.Logger().Info("attempting to establish new api session")
		err := context.Authenticate()
		if err != nil {
			return backoff.Permanent(err)
		}

		return err
	}
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxInterval = 10 * time.Second
	expBackoff.MaxElapsedTime = options.GetConnectTimeout()

	return backoff.Retry(operation, expBackoff)
}

func (context *ContextImpl) GetCurrentIdentity() (*rest_model.IdentityDetail, error) {
	if err := context.ensureApiSession(); err != nil {
		return nil, errors.Wrap(err, "failed to establish api session")
	}

	return context.CtrlClt.GetCurrentIdentity()
}

func (context *ContextImpl) GetCurrentIdentityWithBackoff() (*rest_model.IdentityDetail, error) {
	expBackoff := backoff.NewExponentialBackOff()

	expBackoff.InitialInterval = time.Second
	expBackoff.MaxInterval = 30 * time.Second
	expBackoff.MaxElapsedTime = 5 * time.Minute

	var detail *rest_model.IdentityDetail
	operation := func() error {
		var err error
		detail, err = context.GetCurrentIdentity()
		return err
	}

	if err := backoff.Retry(operation, expBackoff); err != nil {
		return nil, err
	}

	return detail, nil
}

func (context *ContextImpl) setUnauthenticated() {
	prevApiSessionPtr := context.CtrlClt.ApiSession.Swap(nil)
	willEmit := prevApiSessionPtr != nil

	context.CtrlClt.ApiSessionCertificate = nil

	context.CloseAllEdgeRouterConns()
	context.sessions.Clear()

	if willEmit {
		context.Emit(EventAuthenticationStateUnauthenticated, *prevApiSessionPtr)
	}
}

func (context *ContextImpl) authenticate() error {
	logrus.Debug("attempting to authenticate")
	context.services = cmap.New[*rest_model.ServiceDetail]()
	context.sessions = cmap.New[*rest_model.SessionDetail]()
	context.intercepts = cmap.New[*edge.InterceptV1Config]()

	context.setUnauthenticated()

	apiSession, err := context.CtrlClt.Authenticate()

	if err != nil {
		return err
	}

	authQueries := apiSession.GetAuthQueries()
	if len(authQueries) != 0 {
		context.Emit(EventAuthenticationStatePartial, apiSession)
		for _, authQuery := range apiSession.GetAuthQueries() {
			if err := context.handleAuthQuery(authQuery); err != nil {
				return err
			}
		}

		return nil
	}

	return context.onFullAuth(apiSession)
}

func (context *ContextImpl) Reauthenticate() error {
	context.CtrlClt.ApiSession.Store(nil)
	context.CtrlClt.ApiSessionCertificate = nil

	return context.authenticate()
}

func (context *ContextImpl) Authenticate() error {
	if context.CtrlClt.GetCurrentApiSession() != nil {
		if time.Since(context.lastSuccessfulApiSessionRefresh) < 5*time.Second {
			return nil
		}
		logrus.Debug("previous apiSession detected, checking if valid")
		if err := context.RefreshApiSessionWithBackoff(); err == nil {
			logrus.Info("previous apiSession refreshed")
			context.lastSuccessfulApiSessionRefresh = time.Now()
			return nil
		} else {
			logrus.WithError(err).Info("previous apiSession failed to refresh, attempting to authenticate")
		}
	}

	return context.authenticate()
}

func (context *ContextImpl) RefreshApiSessionWithBackoff() error {
	expBackoff := backoff.NewExponentialBackOff()

	expBackoff.InitialInterval = 5 * time.Second
	expBackoff.MaxInterval = 5 * time.Minute
	expBackoff.MaxElapsedTime = 24 * time.Hour

	operation := func() error {
		newApiSession, err := context.CtrlClt.Refresh()
		if err == nil {
			context.updateTokenOnAllErs(newApiSession)
			return nil
		}

		unauthorizedErr := &current_api_session.GetCurrentAPISessionUnauthorized{}
		if errors.As(err, &unauthorizedErr) {
			logrus.Info("previous apiSession expired")
			return backoff.Permanent(err)
		}
		logrus.WithError(err).Info("unable to refresh apiSession, will retry")
		return err
	}

	return backoff.Retry(operation, expBackoff)
}

func (context *ContextImpl) CloseAllEdgeRouterConns() {
	for entry := range context.routerConnections.IterBuffered() {
		key, val := entry.Key, entry.Val
		if !val.IsClosed() {
			if err := val.Close(); err != nil {
				pfxlog.Logger().WithError(err).Error("error while closing edge router connection")
			}
		}

		context.routerConnections.Remove(key)
	}
}

func (context *ContextImpl) onFullAuth(apiSession apis.ApiSession) error {
	var doOnceErr error
	context.firstAuthOnce.Do(func() {
		if context.options.OnContextReady != nil {
			context.options.OnContextReady(context)
		}
		go context.runRefreshes()

		metricsTags := map[string]string{
			"srcId": apiSession.GetIdentityId(),
		}

		context.metrics = metrics.NewRegistry(apiSession.GetIdentityName(), metricsTags)
	})

	context.Emit(EventAuthenticationStateFull, apiSession)

	// get services
	if err := context.RefreshServices(); err != nil {
		doOnceErr = err
	}

	return doOnceErr
}

func (context *ContextImpl) AddZitiMfaHandler(handler func(query *rest_model.AuthQueryDetail, response MfaCodeResponse) error) {
	context.authQueryHandlers[string(rest_model.MfaProvidersZiti)] = handler
}

func (context *ContextImpl) authenticateMfa(code string) error {
	if err := context.CtrlClt.AuthenticateMFA(code); err != nil {
		return err
	}

	newApiSession, err := context.CtrlClt.Refresh()

	if err != nil {
		return err
	}
	context.updateTokenOnAllErs(newApiSession)

	apiSession := context.CtrlClt.GetCurrentApiSession()

	if apiSession != nil && len(apiSession.GetAuthQueries()) == 0 {
		return context.onFullAuth(apiSession)
	}

	return nil
}

func (context *ContextImpl) handleAuthQuery(authQuery *rest_model.AuthQueryDetail) error {
	context.Emit(EventAuthQuery, authQuery)

	if authQuery.Provider == nil {
		return fmt.Errorf("unhandled response from controller: authentication query has no provider specified")
	}

	if *authQuery.Provider == rest_model.MfaProvidersZiti {
		handler := context.authQueryHandlers[string(rest_model.MfaProvidersZiti)]

		context.Emit(EventMfaTotpCode, authQuery, MfaCodeResponse(context.authenticateMfa))

		if handler == nil {
			pfxlog.Logger().Debugf("no callback handler registered for provider: %v, event will still be emitted", *authQuery.Provider)
		} else {
			return handler(authQuery, context.authenticateMfa)
		}

		return nil
	}

	return fmt.Errorf("unsupported MFA provider: %v", *authQuery.Provider)
}

func (context *ContextImpl) Dial(serviceName string) (edge.Conn, error) {
	defaultOptions := &DialOptions{ConnectTimeout: 5 * time.Second}
	return context.DialWithOptions(serviceName, defaultOptions)
}

func (context *ContextImpl) DialWithOptions(serviceName string, options *DialOptions) (edge.Conn, error) {
	edgeDialOptions := &edge.DialOptions{
		ConnectTimeout:  options.ConnectTimeout,
		Identity:        options.Identity,
		AppData:         options.AppData,
		StickinessToken: options.StickinessToken,
	}
	if edgeDialOptions.GetConnectTimeout() == 0 {
		edgeDialOptions.ConnectTimeout = 15 * time.Second
	}

	if err := context.ensureApiSession(); err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	svc, ok := context.GetService(serviceName)
	if !ok {
		return nil, errors.Errorf("service '%s' not found", serviceName)
	}

	context.CtrlClt.PostureCache.AddActiveService(*svc.ID)

	edgeDialOptions.CallerId = context.CtrlClt.GetCurrentApiSession().GetIdentityName()

	session, err := context.GetSession(*svc.ID)
	if err != nil {
		context.deleteServiceSessions(*svc.ID)
		if session, err = context.createSessionWithBackoff(svc, SessionType(SessionDial), options); err != nil {
			return nil, errors.Wrapf(err, "unable to dial service '%v'", serviceName)
		}
	}

	pfxlog.Logger().WithField("sessionId", *session.ID).WithField("sessionToken", session.Token).Debug("connecting with session")
	conn, err := context.dialSession(svc, session, edgeDialOptions)
	if err == nil {
		return conn, nil
	}

	var refreshErr error
	if _, refreshErr = context.refreshSession(session); refreshErr == nil {
		// if the session wasn't expired, no reason to try again, return the failure
		return nil, errors.Wrapf(err, "unable to dial service '%s'", serviceName)
	}

	context.deleteServiceSessions(*svc.ID)
	if session, refreshErr = context.createSessionWithBackoff(svc, SessionType(SessionDial), options); refreshErr != nil {
		// couldn't create a new session, report the error
		return nil, errors.Wrapf(refreshErr, "unable to dial service '%s'", serviceName)
	}

	// retry with new session
	conn, err = context.dialSession(svc, session, edgeDialOptions)
	if err == nil {
		return conn, nil
	}

	return nil, errors.Wrapf(err, "unable to dial service '%s'", serviceName)
}

// GetServiceForAddr finds the service with intercept that matches best to given address
func (context *ContextImpl) GetServiceForAddr(network, hostname string, port uint16) (*rest_model.ServiceDetail, int, error) {
	var svc *rest_model.ServiceDetail
	score := math.MaxInt
	lowestFound := false
	context.intercepts.IterCb(func(key string, intercept *edge.InterceptV1Config) {
		if lowestFound {
			return
		}

		sc := intercept.Match(network, hostname, port)
		if sc != -1 {
			if score > sc {
				score = sc
				svc = intercept.Service
			} else if score == sc && *intercept.Service.Name < *svc.Name { // if score is the same, pick alphabetically first service
				score = sc
				svc = intercept.Service
			}

			if sc == 0 {
				lowestFound = true
			}
		}
	})

	if svc == nil {
		return nil, -1, errors.Errorf("no service for address[%s:%s:%d]", network, hostname, port)
	}

	return svc, score, nil
}

func (context *ContextImpl) dialServiceFromAddr(service, network, host string, port uint16) (edge.Conn, error) {
	appdata := make(map[string]any)
	appdata["dst_protocol"] = network
	appdata["dst_port"] = strconv.Itoa(int(port))
	ip := net.ParseIP(host)
	if len(ip) != 0 {
		appdata["dst_ip"] = host
	} else {
		appdata["dst_hostname"] = host
	}

	options := &DialOptions{
		ConnectTimeout: 5 * time.Second,
	}
	appdataJson, _ := json.Marshal(appdata)
	options.AppData = appdataJson

	return context.DialWithOptions(service, options)
}

func (context *ContextImpl) DialAddr(network string, addr string) (edge.Conn, error) {
	host, portStr, err := net.SplitHostPort(addr)

	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	network = normalizeProtocol(network)

	svc, _, err := context.GetServiceForAddr(network, host, uint16(port))
	if err != nil {
		return nil, err
	}

	return context.dialServiceFromAddr(*svc.Name, network, host, uint16(port))
}

func (context *ContextImpl) dialSession(service *rest_model.ServiceDetail, session *rest_model.SessionDetail, options *edge.DialOptions) (edge.Conn, error) {
	edgeConnFactory, err := context.getEdgeRouterConn(session, options)
	if err != nil {
		return nil, err
	}
	return edgeConnFactory.Connect(service, session, options)
}

func (context *ContextImpl) ensureApiSession() error {
	if context.CtrlClt.GetCurrentApiSession() == nil {
		if err := context.Authenticate(); err != nil {
			return fmt.Errorf("no apiSession, authentication attempt failed: %v", err)
		}
	}
	return nil
}

func (context *ContextImpl) Listen(serviceName string) (edge.Listener, error) {
	return context.ListenWithOptions(serviceName, DefaultListenOptions())
}

func (context *ContextImpl) ListenWithOptions(serviceName string, options *ListenOptions) (edge.Listener, error) {
	if err := context.ensureApiSession(); err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	if s, ok := context.GetService(serviceName); ok {
		return context.listenSession(s, options)
	}
	return nil, errors.Errorf("service '%s' not found in ziti network", serviceName)
}

func (context *ContextImpl) listenSession(service *rest_model.ServiceDetail, options *ListenOptions) (edge.Listener, error) {
	edgeListenOptions := edge.NewListenOptions()
	edgeListenOptions.Cost = options.Cost
	edgeListenOptions.Precedence = edge.Precedence(options.Precedence)
	edgeListenOptions.ConnectTimeout = options.ConnectTimeout
	if options.MaxTerminators != 0 {
		edgeListenOptions.MaxTerminators = options.MaxTerminators
	} else {
		edgeListenOptions.MaxTerminators = options.MaxConnections
	}
	edgeListenOptions.Identity = options.Identity
	edgeListenOptions.BindUsingEdgeIdentity = options.BindUsingEdgeIdentity
	edgeListenOptions.ManualStart = options.ManualStart

	if edgeListenOptions.ConnectTimeout == 0 {
		edgeListenOptions.ConnectTimeout = time.Minute
	}

	if edgeListenOptions.MaxTerminators < 1 {
		edgeListenOptions.MaxTerminators = 1
	}

	if listenerMgr, err := newListenerManager(service, context, edgeListenOptions, options.WaitForNEstablishedListeners); err != nil {
		return nil, err
	} else {
		return listenerMgr.listener, nil
	}
}

func (context *ContextImpl) getEdgeRouterConn(session *rest_model.SessionDetail, options edge.ConnOptions) (edge.RouterConn, error) {
	logger := pfxlog.Logger().WithField("sessionId", *session.ID)

	if len(session.EdgeRouters) == 0 {
		if refreshedSession, err := context.refreshSession(session); err != nil {
			target := &rest_session.DetailSessionNotFound{}
			if errors.As(err, &target) {
				sessionKey := fmt.Sprintf("%s:%s", session.Service.ID, *session.Type)
				context.sessions.Remove(sessionKey)
			}

			return nil, fmt.Errorf("no edge routers available, refresh errored: %v", err)
		} else {
			if len(refreshedSession.EdgeRouters) == 0 {
				return nil, errors.New("no edge routers available, refresh yielded no new edge routers")
			}

			session = refreshedSession
		}
	}

	// go through connected routers first
	bestLatency := time.Duration(math.MaxInt64)
	var bestER edge.RouterConn
	var unconnected []*rest_model.SessionEdgeRouter
	for _, edgeRouter := range session.EdgeRouters {
		for proto, addr := range edgeRouter.SupportedProtocols {
			addr = strings.Replace(addr, "://", ":", 1)
			edgeRouter.SupportedProtocols[proto] = addr
			if er, found := context.routerConnections.Get(addr); found {
				h := context.metrics.Histogram("latency." + addr).(metrics2.Histogram)
				if h.Mean() < float64(bestLatency) {
					bestLatency = time.Duration(int64(h.Mean()))
					bestER = er
				}
			} else {
				unconnected = append(unconnected, edgeRouter)
			}
		}
	}

	var ch chan *edgeRouterConnResult
	if bestER == nil {
		ch = make(chan *edgeRouterConnResult, len(unconnected))
	}

	for _, edgeRouter := range unconnected {
		for _, addr := range edgeRouter.SupportedProtocols {
			if context.options.isEdgeRouterUrlAccepted(addr) {
				go context.handleConnectEdgeRouter(*edgeRouter.Name, addr, ch)
			}
		}
	}

	if bestER != nil {
		logger.Debugf("selected router[%s@%s] for best latency(%d ms)",
			bestER.GetRouterName(), bestER.Key(), bestLatency.Milliseconds())
		return bestER, nil
	}

	timeout := time.After(options.GetConnectTimeout())
	for {
		select {
		case f := <-ch:
			if f.routerConnection != nil {
				logger.Debugf("using edgeRouter[%s]", f.routerConnection.Key())
				return f.routerConnection, nil
			}
		case <-timeout:
			return nil, errors.New("no edge routers connected in time")
		}
	}
}

func (context *ContextImpl) handleConnectEdgeRouter(routerName, ingressUrl string, ret chan *edgeRouterConnResult) {
	result := context.connectEdgeRouter(routerName, ingressUrl)

	if ret != nil {
		select {
		case ret <- result:
		case <-time.After(10 * time.Second):
		}
	}
}

func (context *ContextImpl) connectEdgeRouter(routerName, ingressUrl string) *edgeRouterConnResult {
	logger := pfxlog.Logger().WithField("router", routerName)

	if conn, found := context.routerConnections.Get(ingressUrl); found {
		if !conn.IsClosed() {
			return &edgeRouterConnResult{
				routerUrl:        ingressUrl,
				routerName:       routerName,
				routerConnection: conn,
			}
		} else {
			context.routerConnections.Remove(ingressUrl)
		}
	}

	ingAddr, err := transport.ParseAddress(ingressUrl)
	if err != nil {
		logger.WithError(err).Errorf("failed to parse url[%s]", ingressUrl)
		return &edgeRouterConnResult{
			routerUrl:  ingressUrl,
			routerName: routerName,
			err:        err,
		}
	}

	currentApiSession := context.CtrlClt.GetCurrentApiSession()
	if currentApiSession == nil {
		return &edgeRouterConnResult{
			routerUrl:  ingressUrl,
			routerName: routerName,
			err:        errors.New("not authenticated to controller"),
		}
	}

	logger.Debugf("connection to edge router using api session token %s", string(currentApiSession.GetToken()))
	id, err := context.CtrlClt.GetIdentity()

	if err != nil {
		return &edgeRouterConnResult{
			routerUrl:  ingressUrl,
			routerName: routerName,
			err:        err,
		}
	}

	dialerConfig := channel.DialerConfig{
		Identity: identity.NewIdentity(id),
		Endpoint: ingAddr,
		Headers: map[int32][]byte{
			edge.SessionTokenHeader: context.CtrlClt.GetCurrentApiSession().GetToken(),
		},
		TransportConfig: map[interface{}]interface{}{},
	}

	if context.routerProxy != nil {
		if proxyConfig := context.routerProxy(ingressUrl); proxyConfig != nil {
			dialerConfig.TransportConfig[transport.KeyCachedProxyConfiguration] = proxyConfig
		}
	}

	dialer := channel.NewClassicDialer(dialerConfig)

	start := time.Now().UnixNano()
	edgeConn := network.NewEdgeConnFactory(routerName, ingressUrl, context)
	options := channel.DefaultOptions()
	options.ConnectTimeout = 15 * time.Second
	ch, err := channel.NewChannel(fmt.Sprintf("ziti-sdk[router=%v]", ingressUrl), dialer, edgeConn, options)
	if err != nil {
		logger.Error(err)
		return &edgeRouterConnResult{
			routerUrl:  ingressUrl,
			routerName: routerName,
			err:        err,
		}
	}
	connectTime := time.Duration(time.Now().UnixNano() - start)
	logger.Debugf("routerConn[%s@%s] connected in %d ms", routerName, ingressUrl, connectTime.Milliseconds())

	if versionHeader, found := ch.Underlay().Headers()[channel.HelloVersionHeader]; found {
		versionInfo, err := versions.StdVersionEncDec.Decode(versionHeader)
		if err != nil {
			pfxlog.Logger().Errorf("could not parse hello version header: %v", err)
		} else {
			pfxlog.Logger().
				WithField("os", versionInfo.OS).
				WithField("arch", versionInfo.Arch).
				WithField("version", versionInfo.Version).
				WithField("revision", versionInfo.Revision).
				WithField("buildDate", versionInfo.BuildDate).
				Debug("connected to edge router")
		}
	}

	logger.Debugf("connected to %s", ingressUrl)

	context.Emit(EventRouterConnected, edgeConn.GetRouterName(), edgeConn.Key())

	useConn := context.routerConnections.Upsert(ingressUrl, edgeConn,
		func(exist bool, oldV edge.RouterConn, newV edge.RouterConn) edge.RouterConn {
			if exist { // use the routerConnection already in the map, close new one
				pfxlog.Logger().Infof("connection to %s already established, closing duplicate connection", ingressUrl)
				go func() {
					if err := newV.Close(); err != nil {
						pfxlog.Logger().Errorf("unable to close router connection (%v)", err)
					}
				}()
				return oldV
			}
			h := context.metrics.Histogram("latency." + ingressUrl)
			h.Update(int64(connectTime))

			latencyProbeConfig := &latency.ProbeConfig{
				Channel:  ch,
				Interval: LatencyCheckInterval,
				Timeout:  LatencyCheckTimeout,
				ResultHandler: func(resultNanos int64) {
					h.Update(resultNanos)
				},
				TimeoutHandler: func() {
					logrus.Errorf("latency timeout after [%s]", LatencyCheckTimeout)
					if ch.GetTimeSinceLastRead() > LatencyCheckInterval {
						// No traffic on channel, no response. Close the channel
						logrus.Error("no read traffic on channel since before latency probe was sent, closing channel")
						_ = ch.Close()
					}
				},
				ExitHandler: func() {
					h.Dispose()
				},
			}

			go latency.ProbeLatencyConfigurable(latencyProbeConfig)
			return newV
		})

	return &edgeRouterConnResult{
		routerUrl:        ingressUrl,
		routerName:       routerName,
		routerConnection: useConn,
	}
}

func (context *ContextImpl) GetServiceId(name string) (string, bool, error) {
	if err := context.ensureApiSession(); err != nil {
		return "", false, fmt.Errorf("failed to get service id: %v", err)
	}

	if svc, found := context.GetService(name); found {
		return *svc.ID, true, nil
	}

	return "", false, nil
}

func (context *ContextImpl) GetService(name string) (*rest_model.ServiceDetail, bool) {
	if err := context.ensureApiSession(); err != nil {
		pfxlog.Logger().Warnf("failed to get service: %v", err)
		return nil, false
	}

	if svc, found := context.services.Get(name); !found {
		return nil, false
	} else {
		return svc, true
	}
}

func (context *ContextImpl) GetServices() ([]rest_model.ServiceDetail, error) {
	if err := context.ensureApiSession(); err != nil {
		return nil, fmt.Errorf("failed to get services: %v", err)
	}

	var res []rest_model.ServiceDetail
	context.services.IterCb(func(key string, svc *rest_model.ServiceDetail) {
		res = append(res, *svc)
	})

	return res, nil
}

func (context *ContextImpl) GetServiceTerminators(serviceName string, offset, limit int) ([]*rest_model.TerminatorClientDetail, int, error) {
	svc, found := context.GetService(serviceName)
	if !found {
		return nil, 0, errors.Errorf("did not find service named %v", serviceName)
	}
	return context.CtrlClt.GetServiceTerminators(svc, offset, limit)
}

func (context *ContextImpl) GetSession(serviceId string) (*rest_model.SessionDetail, error) {
	return context.getOrCreateSession(serviceId, SessionType(SessionDial))
}

func (context *ContextImpl) getOrCreateSession(serviceId string, sessionType SessionType) (*rest_model.SessionDetail, error) {
	sessionKey := fmt.Sprintf("%s:%s", serviceId, sessionType)

	cache := string(sessionType) == string(SessionDial)

	// Can't cache Bind sessions, as we use session tokens for routing. If there are multiple binds on a single
	// session routing information will get overwritten
	if cache {
		session, ok := context.sessions.Get(sessionKey)
		if ok {
			return session, nil
		}
	}

	context.CtrlClt.PostureCache.AddActiveService(serviceId)
	session, err := context.CtrlClt.CreateSession(serviceId, sessionType)

	if err != nil {
		return nil, err
	}
	context.cacheSession("create", session)
	return session, nil
}

func (context *ContextImpl) createSessionWithBackoff(service *rest_model.ServiceDetail, sessionType SessionType, options edge.ConnOptions) (*rest_model.SessionDetail, error) {
	expBackoff := backoff.NewExponentialBackOff()

	if sessionType == SessionType(rest_model.DialBindDial) {
		expBackoff.InitialInterval = 50 * time.Millisecond
	} else {
		expBackoff.InitialInterval = time.Second
	}

	expBackoff.MaxInterval = 10 * time.Second
	expBackoff.MaxElapsedTime = options.GetConnectTimeout()

	var session *rest_model.SessionDetail
	operation := func() error {
		latestSvc, _ := context.services.Get(*service.Name)
		if latestSvc != nil && *latestSvc.ID != *service.ID {
			pfxlog.Logger().
				WithField("serviceName", *service.Name).
				WithField("oldServiceId", *service.ID).
				WithField("newServiceId", *latestSvc.ID).
				Info("service id changed, service was recreated")
			service = latestSvc
		}

		s, err := context.createSession(service, sessionType)
		if err != nil {
			return err
		}
		session = s
		return nil
	}

	if session != nil {
		context.CtrlClt.PostureCache.AddActiveService(*service.ID)
		context.cacheSession("create", session)
	}

	return session, backoff.Retry(operation, expBackoff)
}

func (context *ContextImpl) createSession(service *rest_model.ServiceDetail, sessionType SessionType) (*rest_model.SessionDetail, error) {
	start := time.Now()
	logger := pfxlog.Logger()
	logger.Debugf("establishing %s session to service %s", sessionType, *service.Name)
	session, err := context.getOrCreateSession(*service.ID, sessionType)
	if err != nil {
		logger.WithError(err).WithField("errorType", fmt.Sprintf("%T", err)).Warnf("failure creating %s session to service %s", sessionType, *service.Name)
		var target error = &rest_session.CreateSessionUnauthorized{}
		if errors.As(err, &target) {
			if err := context.Authenticate(); err != nil {
				target = &authentication.AuthenticateUnauthorized{}
				if errors.As(err, &target) {
					return nil, backoff.Permanent(err)
				}
				return nil, err
			}
		}

		target = &rest_session.CreateSessionNotFound{}
		if errors.As(err, &target) {
			if refreshErr := context.refreshServices(false); refreshErr != nil {
				logger.WithError(refreshErr).Info("failed to refresh services after create session returned 404 (likely for service)")
			}
		}

		return nil, err
	}
	elapsed := time.Since(start)
	logger.Debugf("successfully created %s session to service %s in %vms", sessionType, *service.Name, elapsed.Milliseconds())
	return session, nil
}

func (context *ContextImpl) refreshSession(session *rest_model.SessionDetail) (*rest_model.SessionDetail, error) {
	var refreshedSession *rest_model.SessionDetail
	var err error
	if strings.HasPrefix(*session.Token, apis.JwtTokenPrefix) {
		refreshedSession, err = context.CtrlClt.GetSessionFromJwt(*session.Token)
	} else {
		refreshedSession, err = context.CtrlClt.GetSession(*session.ID)
	}

	if err != nil {
		return nil, err
	}

	context.cacheSession("refresh", refreshedSession)
	return refreshedSession, nil
}

func (context *ContextImpl) cacheSession(op string, session *rest_model.SessionDetail) {
	sessionKey := fmt.Sprintf("%s:%s", *session.ServiceID, *session.Type)

	if *session.Type == SessionDial {
		if op == "create" {
			context.sessions.Set(sessionKey, session)
		} else if op == "refresh" {
			// N.B.: refreshed sessions do not contain token so update stored session object with updated edgeRouters
			isUpdate := false
			val := context.sessions.Upsert(sessionKey, session, func(exist bool, valueInMap *rest_model.SessionDetail, newValue *rest_model.SessionDetail) *rest_model.SessionDetail {
				isUpdate = exist
				return newValue
			})
			if isUpdate {
				existingSession := val
				existingSession.EdgeRouters = session.EdgeRouters
			}
		}
	}
}

func (context *ContextImpl) deleteServiceSessions(svcId string) {
	context.sessions.Remove(fmt.Sprintf("%s:%s", svcId, SessionBind))
	context.sessions.Remove(fmt.Sprintf("%s:%s", svcId, SessionDial))
}

func (context *ContextImpl) Close() {
	if context.closed.CompareAndSwap(false, true) {
		close(context.closeNotify)

		context.CloseAllEdgeRouterConns()

	}
}

func (context *ContextImpl) Metrics() metrics.Registry {
	return context.metrics
}

func (context *ContextImpl) EnrollZitiMfa() (*rest_model.DetailMfa, error) {
	return context.CtrlClt.EnrollMfa()
}

func (context *ContextImpl) VerifyZitiMfa(code string) error {
	return context.CtrlClt.VerifyMfa(code)
}
func (context *ContextImpl) RemoveZitiMfa(code string) error {
	return context.CtrlClt.RemoveMfa(code)
}

type waitForNHelper struct {
	count  uint
	mgr    *listenerManager
	notify chan struct{}
	closed atomic.Bool
}

func (self *waitForNHelper) Notify(eventType ListenEventType) {
	if eventType == ListenerEstablished && self.mgr.listener.GetEstablishedCount() >= self.count {
		if self.closed.CompareAndSwap(false, true) {
			close(self.notify)
		}
	}
}

func (self *waitForNHelper) WaitForN(timeout time.Duration) error {
	select {
	case <-time.After(timeout):
		return fmt.Errorf("timed out waiting for %v listeners to be established, only had %v", self.count, self.mgr.listener.GetEstablishedCount())
	case <-self.notify:

	}
	return nil
}

func newListenerManager(service *rest_model.ServiceDetail, context *ContextImpl, options *edge.ListenOptions, waitForN uint) (*listenerManager, error) {
	now := time.Now()

	var keyPair *kx.KeyPair
	if service.EncryptionRequired != nil && *service.EncryptionRequired {
		var err error
		keyPair, err = kx.NewKeyPair()
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create end-to-end encrytpion key-pair while hosting service '%s'", *service.Name)
		}
	}

	options.KeyPair = keyPair
	options.ListenerId = uuid.NewString()

	listenerMgr := &listenerManager{
		service:           service,
		context:           context,
		options:           options,
		routerConnections: map[string]edge.RouterConn{},
		connects:          map[string]time.Time{},
		connectChan:       make(chan *edgeRouterConnResult, 3),
		eventChan:         make(chan listenerEvent),
		disconnectedTime:  &now,
	}

	listenerMgr.listener = network.NewMultiListener(service, listenerMgr.GetCurrentSession)

	var helper *waitForNHelper
	if waitForN > 0 {
		helper = &waitForNHelper{
			count:  waitForN,
			mgr:    listenerMgr,
			notify: make(chan struct{}),
		}
		listenerMgr.AddObserver(helper)
		defer listenerMgr.RemoveObserver(helper)
	}

	go listenerMgr.run()

	if helper != nil {
		if err := helper.WaitForN(options.ConnectTimeout); err != nil {
			result := errorz.MultipleErrors{}
			result = append(result, err)
			if closeErr := listenerMgr.listener.Close(); closeErr != nil {
				result = append(result, closeErr)
			}
			return nil, result.ToError()
		}
	}

	return listenerMgr, nil
}

type listenerManager struct {
	service                *rest_model.ServiceDetail
	context                *ContextImpl
	session                *rest_model.SessionDetail
	options                *edge.ListenOptions
	routerConnections      map[string]edge.RouterConn
	connects               map[string]time.Time
	listener               network.MultiListener
	connectChan            chan *edgeRouterConnResult
	eventChan              chan listenerEvent
	sessionRefreshInterval time.Duration
	restartSessionRefresh  bool
	lastSessionRefresh     time.Time
	disconnectedTime       *time.Time
	observers              concurrenz.CopyOnWriteSlice[ListenEventObserver]
	sessionRefreshBaseLine time.Duration
}

func (mgr *listenerManager) AddObserver(observer ListenEventObserver) {
	mgr.observers.Append(observer)
}

func (mgr *listenerManager) RemoveObserver(observer ListenEventObserver) {
	mgr.observers.Delete(observer)
}

func (mgr *listenerManager) notify(eventType ListenEventType) {
	for _, observer := range mgr.observers.Value() {
		go observer.Notify(eventType)
	}
}

func (mgr *listenerManager) run() {
	log := pfxlog.Logger().WithField("service", stringz.OrEmpty(mgr.service.Name))
	// need to either establish a session, or fail if we can't create one
	for mgr.session == nil {
		mgr.createSessionWithBackoff()
	}

	mgr.makeMoreListeners()

	if mgr.options.BindUsingEdgeIdentity {
		mgr.options.Identity = mgr.context.CtrlClt.GetCurrentApiSession().GetIdentityName()
	}

	if mgr.options.Identity != "" {
		id, err := mgr.context.CtrlClt.GetIdentity()

		if err != nil {
			panic("could not get identity during run")
		}

		identitySecret, err := signing.AssertIdentityWithSecret(id.Cert().PrivateKey)
		if err != nil {
			log.WithError(err).Error("failed to sign identity")
		} else {
			mgr.options.IdentitySecret = string(identitySecret)
		}
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var refreshSessionChan <-chan time.Time

	for !mgr.listener.IsClosed() {
		if mgr.restartSessionRefresh {
			refreshSessionChan = time.After(mgr.sessionRefreshInterval)
			mgr.restartSessionRefresh = false
		}

		//goland:noinspection GoNilness
		select {
		case routerConnectionResult := <-mgr.connectChan:
			mgr.handleRouterConnectResult(routerConnectionResult)
		case event := <-mgr.eventChan:
			event.handle(mgr)
		case <-refreshSessionChan:
			mgr.refreshSession()
			log.Debugf("next refresh in %s", mgr.sessionRefreshInterval.String())
			refreshSessionChan = time.After(mgr.sessionRefreshInterval)
			mgr.sessionRefreshInterval *= 2
			if mgr.sessionRefreshInterval > 30*time.Minute {
				mgr.sessionRefreshInterval = 30 * time.Minute
			}
		case <-ticker.C:
			mgr.makeMoreListeners()
		case <-mgr.options.GetEventChannel():
			mgr.notify(ListenerEstablished)
		case <-mgr.context.closeNotify:
			mgr.listener.CloseWithError(errors.New("context closed"))
		}
	}
}

func (mgr *listenerManager) sessionRefreshed(session *rest_model.SessionDetail) {
	oldUsableCount := mgr.getUsableEndpointCount(mgr.session)
	newUsableCount := mgr.getUsableEndpointCount(session)

	if oldUsableCount >= 0 && newUsableCount == 0 {
		mgr.sessionRefreshInterval = time.Duration(5+rand.Intn(10)) * time.Second
	} else if newUsableCount < mgr.options.MaxTerminators {
		// if there's been a change, check reset baseline, as things seem to be influx
		// we'll back-off if there's no further change
		if oldUsableCount != newUsableCount {
			mgr.sessionRefreshBaseLine = 30 * time.Second
		}

		// vary refresh by half the baseline refresh interval
		halfInterval := mgr.sessionRefreshBaseLine / 2
		wiggleFactor := time.Duration(rand.Int63n(int64(halfInterval)))
		mgr.sessionRefreshInterval = halfInterval + (wiggleFactor * 2)

		if mgr.sessionRefreshBaseLine < 5*time.Minute {
			mgr.sessionRefreshBaseLine += 30 * time.Second
		}
	} else {
		mgr.sessionRefreshInterval = 30 * time.Minute
	}

	mgr.session = session
	mgr.restartSessionRefresh = true
	mgr.lastSessionRefresh = time.Now()

	log := pfxlog.Logger().
		WithField("service", stringz.OrEmpty(mgr.service.Name)).
		WithField("sessionId", stringz.OrEmpty(mgr.session.ID)).
		WithField("usableEndpoints", newUsableCount).
		WithField("nextRefresh", mgr.sessionRefreshInterval.String())
	log.Debug("session refreshed")
}

func (mgr *listenerManager) getUsableEndpointCount(session *rest_model.SessionDetail) int {
	if session == nil {
		return 0
	}

	count := 0
	for _, edgeRouter := range session.EdgeRouters {
		for _, routerUrl := range edgeRouter.SupportedProtocols {
			if mgr.context.options.isEdgeRouterUrlAccepted(routerUrl) {
				count++
			}
		}
	}
	return count
}

func (mgr *listenerManager) handleRouterConnectResult(result *edgeRouterConnResult) {
	log := pfxlog.Logger().
		WithField("serviceName", *mgr.service.Name).
		WithField("listenerCount", len(mgr.routerConnections)).
		WithField("router", result.routerName).
		WithField("routerUrl", result.routerUrl)

	log.Debugf("handling router connect result, success? %v", result.routerConnection != nil)

	delete(mgr.connects, result.routerUrl)
	routerConnection := result.routerConnection
	if routerConnection == nil {
		return
	}

	if len(mgr.routerConnections) < mgr.options.MaxTerminators {
		if _, ok := mgr.routerConnections[routerConnection.GetRouterName()]; !ok {
			mgr.routerConnections[routerConnection.GetRouterName()] = routerConnection
			log.WithField("listenerCount", len(mgr.routerConnections)).
				Debugf("establishing listener to %s", routerConnection.Key())
			go mgr.createListener(routerConnection, mgr.session)
		}
	} else {
		log.Debug("ignoring connection, already have max connections")
	}
}

func (mgr *listenerManager) createListener(routerConnection edge.RouterConn, session *rest_model.SessionDetail) {
	start := time.Now()
	logger := pfxlog.Logger().WithField("serviceName", *mgr.service.Name).
		WithField("router", routerConnection.GetRouterName())
	svc := mgr.listener.GetService()
	listener, err := routerConnection.Listen(svc, session, mgr.options)
	elapsed := time.Since(start)
	if err == nil {
		logger = logger.WithField("connId", listener.Id())
		logger.Debugf("listener established to %v in %vms", routerConnection.Key(), elapsed.Milliseconds())
		mgr.listener.AddListener(listener, func() {
			select {
			case mgr.eventChan <- &routerConnectionListenFailedEvent{router: routerConnection.GetRouterName()}:
			case <-mgr.context.closeNotify:
				logger.Debugf("listener closed, exiting from createListener")
			}
		})
		mgr.eventChan <- listenSuccessEvent{}
		if !routerConnection.GetBoolHeader(edge.SupportsBindSuccessHeader) {
			select {
			case mgr.options.GetEventChannel() <- &edge.ListenerEvent{EventType: edge.ListenerEstablished}:
			default:
			}
		}
	} else {
		logger.Errorf("creating listener failed after %vms: %v", elapsed.Milliseconds(), err)
		mgr.listener.NotifyOfChildError(err)
		select {
		case mgr.eventChan <- &routerConnectionListenFailedEvent{router: routerConnection.GetRouterName()}:
		case <-mgr.context.closeNotify:
			logger.Debugf("listener closed, exiting from createListener")
		}
	}
}

func (mgr *listenerManager) makeMoreListeners() {
	log := pfxlog.Logger().WithField("service", *mgr.service.Name).WithField("erCount", len(mgr.session.EdgeRouters))
	if mgr.listener.IsClosed() || len(mgr.routerConnections) >= mgr.options.MaxTerminators || len(mgr.session.EdgeRouters) <= len(mgr.routerConnections) {
		log.Trace("not trying to make more connections")
		return
	}

	for _, edgeRouter := range mgr.session.EdgeRouters {
		if _, ok := mgr.routerConnections[*edgeRouter.Name]; ok {
			log.WithField("router", *edgeRouter.Name).Trace("already connected")
			// already connected to this router
			continue
		}

		for _, routerUrl := range edgeRouter.SupportedProtocols {
			if !mgr.context.options.isEdgeRouterUrlAccepted(routerUrl) {
				log.WithField("router", *edgeRouter.Name).WithField("url", routerUrl).
					Trace("skipping unusable url")
				continue
			}

			if connectTime, ok := mgr.connects[routerUrl]; ok && time.Since(connectTime) < 30*time.Second {
				// this url already has a connect in progress
				log.WithField("router", *edgeRouter.Name).WithField("url", routerUrl).
					Trace("connect already in progress")
				continue
			}

			log.WithField("router", *edgeRouter.Name).WithField("url", routerUrl).
				Trace("attempting to connect to router")
			mgr.connects[routerUrl] = time.Now()
			go mgr.context.handleConnectEdgeRouter(*edgeRouter.Name, routerUrl, mgr.connectChan)
		}
	}
}

func (mgr *listenerManager) refreshSession() {
	if time.Since(mgr.lastSessionRefresh) < 30*time.Second {
		return
	}

	log := pfxlog.Logger().WithField("service", stringz.OrEmpty(mgr.service.Name))
	if mgr.session == nil {
		log.Debug("establishing initial session")
		mgr.createSessionWithBackoff()
		return
	}

	log = log.WithField("sessionId", stringz.OrEmpty(mgr.session.ID)).WithField("erCount", len(mgr.session.EdgeRouters))
	log.Debug("starting session refresh")
	session, err := mgr.context.refreshSession(mgr.session)

	if err != nil {
		var target error = &rest_session.DetailSessionNotFound{}
		if errors.As(err, &target) {
			// try to create new session
			mgr.createSessionWithBackoff()
			return
		}

		target = &rest_session.DetailSessionUnauthorized{}
		if errors.As(err, &target) {
			log.WithError(err).Debugf("failure refreshing bind session for service %v", mgr.listener.GetServiceName())
			if err := mgr.context.EnsureAuthenticated(mgr.options); err != nil {
				err := fmt.Errorf("unable to establish API session (%w)", err)
				if len(mgr.routerConnections) == 0 {
					mgr.listener.CloseWithError(err)
				}
				return
			}
		}

		session, err = mgr.context.refreshSession(mgr.session)
		if err != nil {
			target = &rest_session.DetailSessionUnauthorized{}
			if errors.As(err, &target) {
				log.WithError(err).Errorf(
					"failure refreshing bind session even after re-authenticating api session. service %v",
					mgr.listener.GetServiceName())
				if len(mgr.routerConnections) == 0 {
					mgr.listener.CloseWithError(err)
				}
				return
			}

			log.WithError(err).Errorf("failed to to refresh session %v", *mgr.session.ID)

			// try to create new session
			mgr.createSessionWithBackoff()
		}
	}

	// token only returned on created, so if we refreshed the session (as opposed to creating a new one) we have to back-fill it on lookups
	if session != nil {
		session.Token = mgr.session.Token
		mgr.sessionRefreshed(session)
	}
}

func (mgr *listenerManager) createSessionWithBackoff() {
	latestSvc, _ := mgr.context.services.Get(*mgr.service.Name)
	if latestSvc != nil && *latestSvc.ID != *mgr.service.ID {
		pfxlog.Logger().
			WithField("serviceName", *mgr.service.Name).
			WithField("oldServiceId", *mgr.service.ID).
			WithField("newServiceId", *latestSvc.ID).
			Info("service id changed, service was recreated")
		mgr.service = latestSvc
	}

	session, err := mgr.context.createSessionWithBackoff(mgr.service, SessionType(SessionBind), mgr.options)
	if session != nil {
		mgr.sessionRefreshed(session)
		pfxlog.Logger().WithField("session token", *session.Token).Info("new service session")
	} else {
		pfxlog.Logger().WithError(err).Errorf("failed to create bind session for service %v", mgr.service.Name)
	}
}

func (mgr *listenerManager) GetCurrentSession() *rest_model.SessionDetail {
	if mgr.listener.IsClosed() {
		return nil
	}
	event := &getSessionEvent{
		doneC: make(chan struct{}),
	}
	timeout := time.After(5 * time.Second)

	select {
	case mgr.eventChan <- event:
	case <-timeout:
		return nil
	}

	select {
	case <-event.doneC:
		return event.session
	case <-timeout:
	}
	return nil
}

type listenerEvent interface {
	handle(mgr *listenerManager)
}

type routerConnectionListenFailedEvent struct {
	router string
}

func (event *routerConnectionListenFailedEvent) handle(mgr *listenerManager) {
	delete(mgr.routerConnections, event.router)
	pfxlog.Logger().WithField("serviceName", *mgr.service.Name).
		WithField("listenerCount", len(mgr.routerConnections)).
		WithField("router", event.router).
		Debugf("child listener connection closed. parent listener closed: %v", mgr.listener.IsClosed())
	now := time.Now()
	if len(mgr.routerConnections) == 0 {
		mgr.disconnectedTime = &now
	}
	mgr.notify(ListenerRemoved)
	if mgr.sessionRefreshInterval > 10*time.Second && time.Since(mgr.lastSessionRefresh) > 10*time.Second {
		mgr.sessionRefreshInterval = time.Duration(100+(rand.Intn(10)*1000)) * time.Millisecond
		mgr.restartSessionRefresh = true
	}
	mgr.refreshSession()
	mgr.makeMoreListeners()
}

type edgeRouterConnResult struct {
	routerUrl        string
	routerName       string
	routerConnection edge.RouterConn
	err              error
}

type listenSuccessEvent struct{}

func (event listenSuccessEvent) handle(mgr *listenerManager) {
	mgr.disconnectedTime = nil
	mgr.notify(ListenerAdded)
}

type getSessionEvent struct {
	session *rest_model.SessionDetail
	doneC   chan struct{}
}

func (event *getSessionEvent) handle(mgr *listenerManager) {
	defer close(event.doneC)
	event.session = mgr.session
}

type ListenEventType int

const (
	ListenerAdded       ListenEventType = 1
	ListenerEstablished ListenEventType = 2
	ListenerRemoved     ListenEventType = 3
)

type ListenEventObserver interface {
	Notify(eventType ListenEventType)
}

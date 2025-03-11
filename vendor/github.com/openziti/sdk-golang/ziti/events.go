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
	"github.com/kataras/go-events"
	"github.com/openziti/edge-api/rest_model"
	edge_apis "github.com/openziti/sdk-golang/edge-apis"
)

const (
	// EventServiceAdded is emitted when a new service is detected by a Ziti SDK context.
	//
	// Arguments:
	// 1) Context - the context that triggered the listener
	// 2) serviceDetail`*rest_model.ServiceDetail` - The full detail record of the service
	EventServiceAdded = events.EventName("service-new")

	// EventServiceChanged is emitted when an existing service undergoes a change in its definition.
	//
	//Arguments:
	// 1) Context - the context that triggered the listener
	// 2) serviceDetail`*rest_model.ServiceDetail` - The full detail record of the service
	EventServiceChanged = events.EventName("service-changed")

	// EventServiceRemoved is emitted when a service is no longer available.
	//
	// Arguments:
	// 1) Context - the context that triggered the listener
	// 2) serviceDetail`*rest_model.ServiceDetail` - The full detail record of the service
	EventServiceRemoved = events.EventName("service-removed")

	// EventRouterConnected is emitted when a connection to an Edge Router is established.
	//
	// Arguments:
	// 1) Context - the context that triggered the listener
	// 2) routerName `string` - The string name of the target router
	// 3) routerKey `string` - A string that uniquely identifies a router connection
	EventRouterConnected = events.EventName("router-connected")

	// EventRouterDisconnected is emitted when a connection to an Edge Router is disconnected.
	//
	// Arguments:
	// 1) Context - the context that triggered the listener
	// 2) routerName `string` - The string name of the target router
	// 3) routerKey `string` - A string that uniquely identifies a router connection
	EventRouterDisconnected = events.EventName("router-disconnected")

	// EventMfaTotpCode is emitted when a Ziti context requires an MFA TOTP code to proceed with authentication.
	//
	// Arguments:
	// 1) Context - the context that triggered the listener
	// 2) query *rest_model.AuthQueryDetail - details the authentication query causing the MFA Code request
	// 3) codeResponse MfaCodeResponse - a function that accepts a string to return to the authentication process. This codeResponse should be invoked with the user supplied TOTP code.
	EventMfaTotpCode = events.EventName("mfa-totp-code")

	// EventAuthQuery is emitted when a Ziti context requires an answer to an authentication query. MFA TOTP is
	// modeled as an authentication query as well and will also trigger the event EventMfaTotpCode.
	//
	// Arguments:
	// 1) Context - the context that triggered the listener
	// 2) query `*rest_model.AuthQueryDetail` - the details of the authentication query
	//
	EventAuthQuery = events.EventName("auth-query")

	// EventAuthenticationStatePartial emitted if a context acquires an API Session that is in a partially authenticated state. Partial authentication
	// allows for interaction w/ MFA TOTP enrollment and answering authentication queries. It does not allow access to service.
	// This event may or may not be emitted depending on the authentication policy the identity is acting under.
	//
	// Arguments:
	// 1) Context - the context that triggered the listener
	// 2) apiSession *rest_model.CurrentAPISessionDetail - details of the current API Session
	EventAuthenticationStatePartial = events.EventName("auth-state-partial")

	// EventAuthenticationStateFull is emitted when a context acquires an API Session that is fully authenticated. The
	// context will have access to services.
	//
	// Arguments:
	// 1) Context - the context that triggered the listener
	// 2) apiSession *rest_model.CurrentApiSessionDetail - details of the current API Session
	EventAuthenticationStateFull = events.EventName("auth-state-full")

	// EventAuthenticationStateUnauthenticated is emitted when a context has reverted to an unauthenticated state after
	// being fully or partially authenticated.
	//
	// Arguments:
	// 1) Context - the context that triggered the listener
	// 2) apiSession *rest_model.CurrentApiSessionDetail - details of the invalid API Session
	EventAuthenticationStateUnauthenticated = events.EventName("auth-state-unauthenticated")

	// EventControllerUrlsUpdated is emitted when a new set of controllers is detected
	//
	// Arguments:
	// 1) Context - the context that triggered the listener
	// 2) apiUrls []*urls.URL - the URLs of the API for the available controllers
	EventControllerUrlsUpdated = events.EventName("controller-urls-updated")
)

// Eventer provides types methods for adding event listeners to a context and exposes some weakly typed functions
// that are useful for debugging/testing.
type Eventer interface {
	// AddServiceAddedListener adds an event listener for the EventServiceAdded event and returns a function to remove
	// the listener. It is emitted any time a new service definition is received. The service detail provided is the
	// service that was added.
	AddServiceAddedListener(func(Context, *rest_model.ServiceDetail)) func()

	// AddServiceChangedListener adds an event listener for the EventServiceChanged event and returns a function to remove
	// the listener. It is emitted any time a known service definition is updated with new values. The service detail
	// provided is the service that was changed.
	AddServiceChangedListener(func(Context, *rest_model.ServiceDetail)) func()

	// AddServiceRemovedListener adds an event listener for the EventServiceRemoved event and returns a function to remove
	// the listener. It is emitted any time known service definition is no longer accessible. The service detail
	// provided is the service that was removed.
	AddServiceRemovedListener(func(Context, *rest_model.ServiceDetail)) func()

	// AddRouterConnectedListener adds an event listener for the EventRouterConnected event and returns a function to remove
	// the listener. It is emitted any time a router connection is established. The strings provided are router name and connection address.
	AddRouterConnectedListener(func(ztx Context, name string, addr string)) func()

	// AddRouterDisconnectedListener adds an event listener for the EventRouterDisconnected event and returns a function to remove
	// the listener. It is emitted any time a router connection is closed. The strings provided are router name and connection address.
	AddRouterDisconnectedListener(func(ztx Context, name string, addr string)) func()

	// AddMfaTotpCodeListener adds an event listener for the EventMfaTotpCode event and returns a function to remove
	// the listener. It is emitted any time the currently authenticated API Session requires an MFA TOTP Code for
	// authentication. The authentication query detail and an MfaCodeResponse function are provided. The MfaCodeResponse
	// should be invoked to answer the MFA TOTP challenge.
	//
	// Authentication challenges for MFA are modeled as authentication queries, and is provided to listeners for
	// informational purposes. This event handler is a specific authentication query that responds to the internal Ziti
	// MFA TOTP challenge only. All authentication queries, including MFA TOTP ones, are also available through
	// AddAuthQueryListener, but does not provide typed response callbacks.
	AddMfaTotpCodeListener(func(Context, *rest_model.AuthQueryDetail, MfaCodeResponse)) func()

	// AddAuthQueryListener adds an event listener for the EventAuthQuery event and returns a function to remove
	// the listener. The event is emitted any time the current API Session is required to pass additional authentication
	// challenges - which enabled MFA functionality.
	AddAuthQueryListener(func(Context, *rest_model.AuthQueryDetail)) func()

	// AddAuthenticationStatePartialListener adds an event listener for the EventAuthenticationStatePartial event and
	// returns a function to remove the listener. Partial authentication occurs when there are unmet authentication
	// queries - which are defined by the authentication policy associated with the identity. The
	// EventAuthQuery or EventMfaTotpCode events will also coincide with this event. Additionally, the authentication
	// queries that triggered this event are available on the API Session detail in the `AuthQueries` field.
	//
	// In the partially authenticated state, a context will have reduced capabilities. It will not be able to
	// update/list services, create service sessions, etc. It will be able to enroll in TOTP MFA and answer
	// authentication queries.
	//
	// One all authentication queries are answered, the EventAuthenticationStateFull event will be emitted. For
	// identities that do not have secondary authentication challenges associated with them, this even will never
	// be emitted.
	AddAuthenticationStatePartialListener(func(Context, edge_apis.ApiSession)) func()

	// AddAuthenticationStateFullListener adds an event listener for the EventAuthenticationStateFull event and
	// returns a function to remove the listener. Full authentication occurs when there are no unmet authentication
	// queries - which are defined by the authentication policy associated with the identity. In a fully authenticated
	// state, the context will be able to perform all client actions.
	AddAuthenticationStateFullListener(func(Context, edge_apis.ApiSession)) func()

	// AddAuthenticationStateUnauthenticatedListener adds an event listener for the EventAuthenticationStateUnauthenticated
	// event and returns a function to remove the listener. The unauthenticated state occurs when the API session
	// currently being used is no longer valid. API Sessions may become invalid due to prolonged inactivity due to
	// network disconnection, the host machine entering a power saving/sleep mode, etc. It may also occur due to
	// administrative action such as removing specific API Sessions or removing entire identities.
	//
	// The API Session detail provided to the listener may be nil. If it is not nil, the API Session detail is the
	// now expired API Session.
	AddAuthenticationStateUnauthenticatedListener(func(Context, edge_apis.ApiSession)) func()

	// AddListener is an alias for .On(eventName, listener).
	AddListener(events.EventName, ...events.Listener)

	// EventNames returns an array listing the events for which the emitter has registered listeners.
	// The values in the array will be strings.
	EventNames() []events.EventName

	// GetMaxListeners returns the max listeners for this emmiter
	// see SetMaxListeners
	GetMaxListeners() int

	// ListenerCount returns the length of all registered listeners to a particular event
	ListenerCount(events.EventName) int

	// Listeners returns a copy of the array of listeners for the event named eventName.
	Listeners(events.EventName) []events.Listener

	// On registers a particular listener for an event, func receiver parameter(s) is/are optional
	On(events.EventName, ...events.Listener)

	// Once adds a one-time listener function for the event named eventName.
	// The next time eventName is triggered, this listener is removed and then invoked.
	Once(events.EventName, ...events.Listener)

	// RemoveAllListeners removes all listeners, or those of the specified eventName.
	// Note that it will remove the event itself.
	// Returns an indicator if event and listeners were found before the remove.
	RemoveAllListeners(events.EventName) bool

	// RemoveListener removes given listener from the event named eventName.
	// Returns an indicator whether listener was removed
	RemoveListener(events.EventName, events.Listener) bool
}

package ziti

import (
	"time"

	"github.com/openziti/edge-api/rest_model"
)

type ServiceEventType string

const (
	ServiceAdded   ServiceEventType = "Added"
	ServiceRemoved ServiceEventType = "Removed"
	ServiceChanged ServiceEventType = "Changed"

	DefaultServiceRefreshInterval = 5 * time.Minute
	DefaultSessionRefreshInterval = time.Hour
	MinRefreshInterval            = time.Second
	DefaultRefreshJitter          = 0.1
)

type serviceCB func(eventType ServiceEventType, service *rest_model.ServiceDetail)

type Options struct {
	// Service refresh interval. May not be less than 1 second
	RefreshInterval time.Duration

	// Edge session refresh interval. Edge session only need to be refreshed if the list of available
	// edge routers has changed. This should be a relatively rare occurrence. If a dial fails, the
	// edge session will be refreshed regardless.
	// May not be less than 1 second
	SessionRefreshInterval time.Duration

	// RefreshJitter is a fraction (0.0–0.5) representing ±percentage of the refresh interval
	// to randomize. For example, 0.1 means each refresh will fire at a random time within ±10%
	// of the configured interval, helping avoid thundering-herd load spikes on the controller.
	RefreshJitter float64

	// Deprecated: OnContextReady is a callback that is invoked after the first successful authentication request. It
	// does not delineate between fully and partially authenticated API Sessions. Use context.AddListener() with the events
	// EventAuthenticationStateFull, EventAuthenticationStatePartial, EventAuthenticationStateUnAuthenticated instead.
	OnContextReady func(ctx Context)

	// Deprecated: OnServiceUpdate is a callback that is invoked when a service changes its definition.
	// Use `zitiContext.AddListener(<eventName>, handler)` where `eventName` may be EventServiceAdded, EventServiceChanged, EventServiceRemoved.
	OnServiceUpdate     serviceCB
	EdgeRouterUrlFilter func(string) bool
}

func (self *Options) isEdgeRouterUrlAccepted(url string) bool {
	return self.EdgeRouterUrlFilter == nil || self.EdgeRouterUrlFilter(url)
}

var DefaultOptions = &Options{
	RefreshInterval:        DefaultServiceRefreshInterval,
	SessionRefreshInterval: DefaultSessionRefreshInterval,
	RefreshJitter:          DefaultRefreshJitter,
	OnServiceUpdate:        nil,
}

type DialOptions struct {
	ConnectTimeout  time.Duration
	Identity        string
	AppData         []byte
	StickinessToken []byte

	// WARNING: Experimental setting, may be removed and/or defaults may change
	//
	// If set to true, flow-control will be managed from the SDK instead of the router. This prevents
	// multiplexed circuits from interfering with each other. If the embedding application is only
	// opening one circuit at a time, then it should not make much difference. Setting this to true
	// may affect memory use and network traffic.
	//
	// NOTES:
	// 1. Currently defaults to false, but may change to true in the future.
	// 2. Requires router side support
	// 3. Note that if config.MaxDefaultConnects is set to a value greater than one, the SDK will
	//    always use sdk flow-control, as otherwise multiple data channels can lead to out-of-order
	//    data corruption.
	SdkFlowControl *bool
}

func (d DialOptions) GetConnectTimeout() time.Duration {
	return d.ConnectTimeout
}

type ListenOptions struct {
	// Initial static cost assigned to terminators for this service
	Cost uint16

	// Initial precedence assigned to terminators for this service
	Precedence Precedence

	// When using WaitForNEstablishedListeners, how long to wait before giving
	// if N listeners can't be established
	ConnectTimeout time.Duration

	// Maximum number of terminators to establish. If a value less than 1 is provided,
	// will default to 1. At most one terminator will be established per available
	// edge router. If both MaxConnections and MaxTerminators have non-zer values,
	// the value from MaxTerminators will be used
	//
	// Deprecated: used MaxTerminators instead.
	MaxConnections int

	// Maximum number of terminators to establish. If a value less than 1 is provided,
	// will default to 1. At most one terminator will be established per available
	// edge router. If both MaxConnections and MaxTerminators have non-zer values,
	// the value from MaxTerminators will be used
	MaxTerminators int

	// Instance name to assign to terminators for this service
	Identity string

	// Assign the name of the edge identity hosting the service to the terminator's instance name
	// Overrides any name specified using the Identity field in ListenOptions
	BindUsingEdgeIdentity bool

	// If set to true, requires that AcceptEdge is called on the edge.Listener
	ManualStart bool

	// NOTE: Experimental setting, may be removed and/or defaults may change
	//
	// If set to true, flow-control will be managed from the SDK instead of the router. This prevents
	// multiplexed circuits from interfering with each other. If the embedding application is only
	// opening one circuit at a time, then it should not make much difference. Setting this to true
	// may affect memory use and network traffic.
	//
	// Currently defaults to false, but may change to true in the future.
	// Requires router side support
	SdkFlowControl *bool

	// If set to true, the dialing identity's id and name will not be saved to the connection
	// when a dial request is received.
	DoNotSaveDialerIdentity bool

	// Wait for N listeners before returning from the Listen call. By default it will return
	// before any listeners have been established.
	WaitForNEstablishedListeners uint
}

func DefaultListenOptions() *ListenOptions {
	return &ListenOptions{
		Cost:           0,
		Precedence:     PrecedenceDefault,
		ConnectTimeout: 5 * time.Second,
		MaxTerminators: 3,
	}
}

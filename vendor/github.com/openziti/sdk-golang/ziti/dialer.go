package ziti

import (
	"context"
	"fmt"
	"math"
	"net"
	"strconv"

	"github.com/openziti/edge-api/rest_model"
)

type Dialer interface {
	Dial(network, address string) (net.Conn, error)
}

type ContextDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type dialer struct {
	fallback   Dialer
	context    context.Context
	collection *CtxCollection
}

// Deprecated: NewDialer will return a dialer from the DefaultCollection that will iterate over the Context instances
// inside the collection searching for the context that best matches the service.
//
// It is suggested that implementations construct their own CtxCollection and use the NewDialer/NewDialerWithFallback present there.
//
// If a matching service is not found, an error is returned. Matching is based on Match() logic in edge.InterceptV1Config.
func NewDialer() Dialer {
	return DefaultCollection.NewDialer()
}

// Deprecated: NewDialerWithFallback will return a dialer from the DefaultCollection that will iterate over the Context
// instances inside the collection searching for the context that best matches the service.
//
// It is suggested that implementations construct their own CtxCollection and use the NewDialer/NewDialerWithFallback present there.
//
// If a matching service is not found, a dial is attempted with the fallback dialer. Matching is based on Match() logic
// in edge.InterceptV1Config.
func NewDialerWithFallback(ctx context.Context, fallback Dialer) Dialer {
	return DefaultCollection.NewDialerWithFallback(ctx, fallback)
}

func (dialer *dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, portString, err := net.SplitHostPort(address)

	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, err
	}

	network = normalizeProtocol(network)

	var ztx Context
	var service *rest_model.ServiceDetail
	var bestFound = false
	best := math.MaxInt
	dialer.collection.ForAll(func(ctx Context) {
		if bestFound {
			return
		}

		srv, score, err := ctx.GetServiceForAddr(network, host, uint16(port))
		if err == nil {
			if score < best {
				best = score
				ztx = ctx
				service = srv
			}

			if score == 0 { // best possible score
				bestFound = true
			}
		}
	})

	if ztx != nil && service != nil {
		return ztx.(*ContextImpl).dialServiceFromAddrWithContext(ctx, *service.Name, network, host, uint16(port))
	}

	if dialer.fallback != nil {
		ctxDialer, ok := dialer.fallback.(ContextDialer)
		if ok {
			return ctxDialer.DialContext(ctx, network, address)
		}
		return dialer.fallback.Dial(network, address)
	}

	return nil, fmt.Errorf("address [%s:%s:%d] is not intercepted by any ziti context", network, host, port)
}

func (dialer *dialer) Dial(network, address string) (net.Conn, error) {
	ctx := dialer.context
	if ctx == nil {
		ctx = context.Background()
	}
	return dialer.DialContext(ctx, network, address)
}

func normalizeProtocol(proto string) string {
	switch proto {
	case "tcp", "tcp4", "tcp6":
		return "tcp"
	case "udp", "udp4", "udp6":
		return "udp"
	default:
		return proto
	}
}

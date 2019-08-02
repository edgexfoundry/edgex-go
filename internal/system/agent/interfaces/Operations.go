package interfaces

import "context"

// The abstraction(s) which should be accessed via a global var.

type ServiceStarter interface {
	Start(service string) error
}

type ServiceStopper interface {
	Stop(service string) error
}

type ServiceRestarter interface {
	Restart(service string) error
}

type MetricsFetcher interface {
	Metrics(ctx context.Context, service string) ([]byte, error)
}

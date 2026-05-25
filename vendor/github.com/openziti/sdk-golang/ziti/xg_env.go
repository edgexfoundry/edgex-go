package ziti

import (
	"github.com/openziti/metrics"
	"github.com/openziti/sdk-golang/xgress"
)

type xgEnv struct {
	retransmitter   *xgress.Retransmitter
	payloadIngester *xgress.PayloadIngester
	metrics         xgress.Metrics
}

func NewXgressEnv(closeNotify <-chan struct{}, registry metrics.Registry) xgress.Env {
	return &xgEnv{
		retransmitter:   xgress.NewRetransmitter(dummyRetransmitterFaultReporter{}, registry, closeNotify),
		payloadIngester: xgress.NewPayloadIngesterWithConfig(5, closeNotify),
		metrics:         xgress.NewMetrics(registry),
	}
}

func (x xgEnv) GetRetransmitter() *xgress.Retransmitter {
	return x.retransmitter
}

func (x xgEnv) GetPayloadIngester() *xgress.PayloadIngester {
	return x.payloadIngester
}

func (x xgEnv) GetMetrics() xgress.Metrics {
	return x.metrics
}

type dummyRetransmitterFaultReporter struct{}

func (d dummyRetransmitterFaultReporter) ReportForwardingFault(circuitId string, ctrlId string) {
	// the only way to get a fault is if the connection goes down, in which case the circuit will
	// get torn down anyway
}

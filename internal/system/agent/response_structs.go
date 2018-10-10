package agent

import (
	"encoding/json"
	"github.com/edgexfoundry/edgex-go/internal"
)

// For handling the response (containing configuration data) returned by the edgex-support-notifications service.
type ConfigRespMap struct {
	Config map[string]interface{}
}

// For handling the response (containing metrics data) returned by the edgex-support-notifications service.
type MetricsRespMap struct {
	Metrics map[string]interface{}
}

func (b ConfigRespMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Service string
		Adaptee1
	}{
		Service:  internal.SupportNotificationsServiceKey,
		Adaptee1: Adaptee1(b),
	})
}

func (m MetricsRespMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Service string
		Adaptee2
	}{
		Service:  internal.SupportNotificationsServiceKey,
		Adaptee2: Adaptee2(m),
	})
}

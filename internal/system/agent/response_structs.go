package agent

import (
	"encoding/json"
	"github.com/edgexfoundry/edgex-go/internal"
)

// For handling the response (containing, for example, configuration or metrics data) returned by the edgex-support-notifications service.
type RespMap struct {
	Config map[string]interface{}
}

func (m RespMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Service string
		Adaptee
	}{
		Service:  internal.SupportNotificationsServiceKey,
		Adaptee: Adaptee(m),
	})
}

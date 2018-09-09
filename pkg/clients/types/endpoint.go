package types

import (
	"fmt"
	"os"
	"time"

	//TODO: Get rid of this import. Should not be here. Perhaps eliminate the need for the consulclient below
	// by passing in a function pointer. Return type needs to me addressed for that to happen, redefined in a
	// place that doesn't cause a circular reference.
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
)

type Endpoint struct{}

func (e Endpoint) Monitor(params EndpointParams, ch chan string) {
	for {
		data, err := consulclient.GetServiceEndpoint(params.ServiceKey)
		if err != nil {
			fmt.Fprintln(os.Stdout, err.Error())
		}
		url := fmt.Sprintf("http://%s:%v%s", data.Address, data.Port, params.Path)
		ch <- url
		time.Sleep(time.Second * time.Duration(params.Interval))
	}
}

package types

import (
	"fmt"
	"os"
	"time"

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
		time.Sleep(time.Second * time.Duration(15))
	}
}

package types

import (
	"time"
	"fmt"
	"os"
	"github.com/edgexfoundry/edgex-go/support/consul-client"
)

type Endpoint struct {}

func(e Endpoint) Monitor(params EndpointParams, ch chan string) {
	check := time.Now()
	for true {
		if time.Now().After(check) {
			data, err := consulclient.GetServiceEndpoint(params.ServiceKey)
			if err != nil {
				fmt.Fprintln(os.Stdout, err.Error())
			}
			url := fmt.Sprintf("http://%s:%v%s", data.Address, data.Port, params.Path)
			ch <- url
			check = time.Now().Add(time.Second * time.Duration(15))
		}
	}
}
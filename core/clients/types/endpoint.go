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
			fmt.Fprintf(os.Stdout, "now:%s\r\ncheck:%s\r\n", time.Now().Format(time.RFC3339Nano), check.Format(time.RFC3339Nano))
			data, err := consulclient.GetServiceEndpoint(params.ServiceKey)
			if err != nil {
				fmt.Fprintln(os.Stdout, err.Error())
			}
			url := fmt.Sprintf("http://%s:%v%s", data.Address, data.Port, params.Path)
			ch <- url
			fmt.Fprintf(os.Stdout, "endpoint loaded %s\r\n", url)
			check = time.Now().Add(time.Second * time.Duration(15))
			fmt.Fprintf(os.Stdout, "setting check:%s\r\n", check.Format(time.RFC3339Nano))
		}
	}
}
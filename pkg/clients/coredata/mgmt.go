package coredata

import (
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

type MgmtClient interface {
	FetchConfiguration() (string, error)
	FetchMetrics() (string, error)
}

type MgmtRestClient struct {
	url      string
	endpoint clients.Endpointer
}

func NewMgmtClient(params types.EndpointParams, m clients.Endpointer) MgmtClient {
	cdc := MgmtRestClient{endpoint: m}
	cdc.initClient(params)
	return &cdc
}

func (cdc *MgmtRestClient) initClient(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go cdc.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
					cdc.url = url
				}
			}
		}(ch)
	} else {
		cdc.url = params.Url
	}
}

// Fetch configuration information from the command service.
func (cdc *MgmtRestClient) FetchConfiguration() (string, error) {

	var result string
	data, err := clients.GetRequest(cdc.url + clients.ApiConfigRoute)
	if err != nil {
		return "", err
	}
	result = string(data)

	return result, nil

}

// Fetch metrics information from the command service.
func (cdc *MgmtRestClient) FetchMetrics() (string, error) {

	var result string
	data, err := clients.GetRequest(cdc.url + clients.ApiMetricsRoute)
	if err != nil {
		return "", err
	}
	result = string(data)

	return result, nil

}

package command

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
	mcc := MgmtRestClient{endpoint: m}
	mcc.initClient(params)
	return &mcc
}

func (mcc *MgmtRestClient) initClient(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go mcc.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
					mcc.url = url
				}
			}
		}(ch)
	} else {
		mcc.url = params.Url
	}
}

// Fetch configuration information from the command service.
func (mcc *MgmtRestClient) FetchConfiguration() (string, error) {

	var result string
	data, err := clients.GetRequest(mcc.url + clients.ApiConfigRoute)
	if err != nil {
		return "", err
	}
	result = string(data)

	return result, nil

}

// Fetch metrics information from the command service.
func (mcc *MgmtRestClient) FetchMetrics() (string, error) {

	var result string
	data, err := clients.GetRequest(mcc.url + clients.ApiMetricsRoute)
	if err != nil {
		return "", err
	}
	result = string(data)

	return result, nil

}

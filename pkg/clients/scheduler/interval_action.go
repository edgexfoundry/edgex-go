package scheduler

import (
	"context"
	"encoding/json"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"net/url"
)

type IntervalActionClient interface {
	Add(dev *models.IntervalAction, ctx context.Context) (string, error)
	Delete(id string, ctx context.Context) error
	DeleteByName(name string, ctx context.Context) error
	IntervalAction(id string, ctx context.Context) (models.IntervalAction, error)
	IntervalActionForName(name string, ctx context.Context) (models.IntervalAction, error)
	IntervalActions(ctx context.Context) ([]models.IntervalAction, error)
	IntervalActionsForTargetByName(name string, ctx context.Context) ([]models.IntervalAction, error)
	Update(dev models.IntervalAction, ctx context.Context) error
}

// receiver for intervalActionClient
type IntervalActionRestClient struct {
	url      string
	endpoint clients.Endpointer
}

func NewIntervalActionClient(params types.EndpointParams, m clients.Endpointer) IntervalActionClient {
	s := IntervalActionRestClient{endpoint: m}
	s.init(params)
	return &s
}

func (s *IntervalActionRestClient) init(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go s.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
					s.url = url
				}
			}
		}(ch)
	} else {
		s.url = params.Url
	}
}

// Helper method to request and decode an interval action
func (s *IntervalActionRestClient) requestIntervalAction(url string, ctx context.Context) (models.IntervalAction, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return models.IntervalAction{}, err
	}

	ia := models.IntervalAction{}
	err = json.Unmarshal(data, &ia)
	return ia, err
}

// Helper method to request and decode an interval action slice
func (s *IntervalActionRestClient) requestIntervalActionSlice(url string, ctx context.Context) ([]models.IntervalAction, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return []models.IntervalAction{}, err
	}

	iaSlice := make([]models.IntervalAction, 0)
	err = json.Unmarshal(data, &iaSlice)
	return iaSlice, err
}

// Add a interval action.
func (s *IntervalActionRestClient) Add(ia *models.IntervalAction, ctx context.Context) (string, error) {
	return clients.PostJsonRequest(s.url, ia, ctx)
}

// Delete a interval action (specified by id).
func (s *IntervalActionRestClient) Delete(id string, ctx context.Context) error {
	return clients.DeleteRequest(s.url+"/id/"+id, ctx)
}

// Delete a interval action (specified by name).
func (s *IntervalActionRestClient) DeleteByName(name string, ctx context.Context) error {
	return clients.DeleteRequest(s.url+"/name/"+url.QueryEscape(name), ctx)
}

// IntervalAction returns the IntervalAction specified by id.
func (s *IntervalActionRestClient) IntervalAction(id string, ctx context.Context) (models.IntervalAction, error) {
	return s.requestIntervalAction(s.url+"/"+id, ctx)
}

// IntervalActionForName returns the IntervalAction specified by name.
func (s *IntervalActionRestClient) IntervalActionForName(name string, ctx context.Context) (models.IntervalAction, error) {
	return s.requestIntervalAction(s.url+"/name/"+url.QueryEscape(name), ctx)
}

// Get a list of all interval actions.
func (s *IntervalActionRestClient) IntervalActions(ctx context.Context) ([]models.IntervalAction, error) {
	return s.requestIntervalActionSlice(s.url, ctx)
}

// Get the interval action for service by name.
func (s *IntervalActionRestClient) IntervalActionsForTargetByName(name string, ctx context.Context) ([]models.IntervalAction, error) {
	return s.requestIntervalActionSlice(s.url+"/target/"+url.QueryEscape(name), ctx)
}

// Update an interval action.
func (s *IntervalActionRestClient) Update(ia models.IntervalAction, ctx context.Context) error {
	return clients.UpdateRequest(s.url, ia, ctx)
}

package scheduler

import (
	"encoding/json"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"net/url"
)

type IntervalActionClient interface {
	Add(dev *models.IntervalAction) (string, error)
	Delete(id string) error
	DeleteByName(name string) error
	IntervalAction(id string) (models.IntervalAction, error)
	IntervalActionForName(name string) (models.IntervalAction, error)
	IntervalActions() ([]models.IntervalAction, error)
	IntervalActionsForTargetByName(name string) ([]models.IntervalAction, error)
	Update(dev models.IntervalAction) error
}

// receiver for intervalActionClient
type IntervalActionRestClient struct {
	url string
	endpoint clients.Endpointer
}

func NewIntervalActionClient (params types.EndpointParams, m clients.Endpointer) IntervalActionClient {
	s := IntervalActionRestClient{endpoint:m}
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
func (s *IntervalActionRestClient) requestIntervalAction(url string) (models.IntervalAction, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return models.IntervalAction{}, err
	}

	ia := models.IntervalAction{}
	err = json.Unmarshal(data, &ia)
	return ia, err
}

// Helper method to request and decode an interval action slice
func (s *IntervalActionRestClient) requestIntervalActionSlice(url string) ([]models.IntervalAction, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return []models.IntervalAction{}, err
	}

	iaSlice := make([]models.IntervalAction, 0)
	err = json.Unmarshal(data, &iaSlice)
	return iaSlice, err
}



// Add a interval action.
func (s *IntervalActionRestClient) Add(ia *models.IntervalAction) (string, error) {
	return clients.PostJsonRequest(s.url, ia)
}

// Delete a interval action (specified by id).
func (s *IntervalActionRestClient) Delete(id string) error {
	return clients.DeleteRequest(s.url + "/id/" + id)
}

// Delete a interval action (specified by name).
func (s *IntervalActionRestClient) DeleteByName(name string) error {
	return clients.DeleteRequest(s.url + "/name/" + url.QueryEscape(name))
}

// IntervalAction returns the IntervalAction specified by id.
func (s *IntervalActionRestClient) IntervalAction(id string) (models.IntervalAction, error) {
	return s.requestIntervalAction(s.url + "/" + id)
}

// IntervalActionForName returns the IntervalAction specified by name.
func (s *IntervalActionRestClient) IntervalActionForName(name string) (models.IntervalAction, error) {
	return s.requestIntervalAction(s.url + "/name/" + url.QueryEscape(name))
}

// Get a list of all interval actions.
func (s *IntervalActionRestClient) IntervalActions() ([]models.IntervalAction, error) {
	return s.requestIntervalActionSlice(s.url)
}

// Get the interval action for service by name.
func (s *IntervalActionRestClient) IntervalActionsForTargetByName(name string) ([]models.IntervalAction, error) {
	return s.requestIntervalActionSlice(s.url + "/target/" + url.QueryEscape(name))
}

// Update an interval action.
func (s *IntervalActionRestClient) Update(ia models.IntervalAction) error {
	return clients.UpdateRequest(s.url, ia)
}


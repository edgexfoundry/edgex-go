package scheduler

import (
	"context"
	"encoding/json"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"net/url"
)

type IntervalClient interface {
	Add(dev *models.Interval, ctx context.Context) (string, error)
	Delete(id string, ctx context.Context) error
	DeleteByName(name string, ctx context.Context) error
	Interval(id string, ctx context.Context) (models.Interval, error)
	IntervalForName(name string, ctx context.Context) (models.Interval, error)
	Intervals(ctx context.Context) ([]models.Interval, error)
	Update(interval models.Interval, ctx context.Context) error
}

type IntervalRestClient struct {
	url      string
	endpoint clients.Endpointer
}

func NewIntervalClient(params types.EndpointParams, m clients.Endpointer) IntervalClient {
	s := IntervalRestClient{endpoint: m}
	s.init(params)
	return &s
}

func (s *IntervalRestClient) init(params types.EndpointParams) {
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

// interface implementations
func (s *IntervalRestClient) Add(interval *models.Interval, ctx context.Context) (string, error) {
	return clients.PostJsonRequest(s.url, interval, ctx)
}

// Delete a interval (specified by id).
func (s *IntervalRestClient) Delete(id string, ctx context.Context) error {
	return clients.DeleteRequest(s.url+"/id/"+id, ctx)
}

// Delete a interval (specified by name).
func (s *IntervalRestClient) DeleteByName(name string, ctx context.Context) error {
	return clients.DeleteRequest(s.url+"/name/"+url.QueryEscape(name), ctx)
}

// support-scheduler returns the interval specified by id.
func (s *IntervalRestClient) Interval(id string, ctx context.Context) (models.Interval, error) {
	return s.requestInterval(s.url+"/"+id, ctx)
}

// ScheduleForName returns the Schedule specified by name.
func (s *IntervalRestClient) IntervalForName(name string, ctx context.Context) (models.Interval, error) {
	return s.requestInterval(s.url+"/name/"+url.QueryEscape(name), ctx)
}

// Schedules returns the list of all schedules.
func (s *IntervalRestClient) Intervals(ctx context.Context) ([]models.Interval, error) {
	return s.requestIntervalSlice(s.url, ctx)
}

// Update a schedule.
func (s *IntervalRestClient) Update(interval models.Interval, ctx context.Context) error {
	return clients.UpdateRequest(s.url, interval, ctx)
}

//
// Helper functions
//

// helper request and decode an interval
func (s *IntervalRestClient) requestInterval(url string, ctx context.Context) (models.Interval, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return models.Interval{}, err
	}

	interval := models.Interval{}
	err = json.Unmarshal(data, &interval)
	return interval, err
}

// helper returns a slice of intervals
func (s *IntervalRestClient) requestIntervalSlice(url string, ctx context.Context) ([]models.Interval, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return []models.Interval{}, err
	}

	sSlice := make([]models.Interval, 0)
	err = json.Unmarshal(data, &sSlice)
	return sSlice, err
}

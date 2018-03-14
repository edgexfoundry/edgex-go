//
// Copyright (c) 2017 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"bytes"
	"fmt"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"io"
	"net/http"
)

// Common http const
const (
	HttpPostMethod     = "POST"
	HttpPutMethod      = "PUT"
	HttpDeleteMethod   = "DELETE"
	ContentType        = "Content-Type"
	ContentTypeJsonVal = "application/json"
)

const (
	ScheduleApiPath      = "/api/v1/schedule"
	ScheduleEventApiPath = "/api/v1/scheduleevent"
	UrlPattern           = "http://%s:%d%s"
)

// Struct to represent the scheduler client
type SchedulerClient struct {
	SchedulerServiceHost string
	SchedulerServicePort int
	OwningService        string
}

// Function to send a schedule to the remote scheduler server
func (schedulerClient SchedulerClient) SendSchedule(schedule models.Schedule) error {
	client := &http.Client{}

	requestBody, err := schedule.MarshalJSON()
	if err != nil {
		return err
	}

	remoteScheduleUrl := fmt.Sprintf(UrlPattern, schedulerClient.SchedulerServiceHost, schedulerClient.SchedulerServicePort, ScheduleApiPath)

	return doPost(remoteScheduleUrl, bytes.NewBuffer(requestBody), client)
}

// Function to update a schedule to the remote scheduler server
func (schedulerClient SchedulerClient) UpdateSchedule(schedule models.Schedule) error {
	client := &http.Client{}

	requestBody, err := schedule.MarshalJSON()
	if err != nil {
		return err
	}

	remoteScheduleUrl := fmt.Sprintf(UrlPattern, schedulerClient.SchedulerServiceHost, schedulerClient.SchedulerServicePort, ScheduleApiPath)

	return doPut(remoteScheduleUrl, bytes.NewBuffer(requestBody), client)
}

// Function to remove a schedule to the remote scheduler server
func (schedulerClient SchedulerClient) RemoveSchedule(id string) error {
	client := &http.Client{}

	remoteScheduleUrl := fmt.Sprintf(UrlPattern, schedulerClient.SchedulerServiceHost, schedulerClient.SchedulerServicePort, ScheduleApiPath)
	remoteScheduleUrl = remoteScheduleUrl + "/" + id

	return doDelete(remoteScheduleUrl, client)
}

// Function to send a schedule event to the remote scheduler server
func (schedulerClient SchedulerClient) SendScheduleEvent(scheduleEvent models.ScheduleEvent) error {
	client := &http.Client{}

	requestBody, err := scheduleEvent.MarshalJSON()
	if err != nil {
		return err
	}

	remoteScheduleEventUrl := fmt.Sprintf(UrlPattern, schedulerClient.SchedulerServiceHost, schedulerClient.SchedulerServicePort, ScheduleEventApiPath)

	return doPost(remoteScheduleEventUrl, bytes.NewBuffer(requestBody), client)
}

// Function to update a schedule event to the remote scheduler server
func (schedulerClient SchedulerClient) UpdateScheduleEvent(scheduleEvent models.ScheduleEvent) error {
	client := &http.Client{}

	requestBody, err := scheduleEvent.MarshalJSON()
	if err != nil {
		return err
	}

	remoteScheduleEventUrl := fmt.Sprintf(UrlPattern, schedulerClient.SchedulerServiceHost, schedulerClient.SchedulerServicePort, ScheduleEventApiPath)

	return doPut(remoteScheduleEventUrl, bytes.NewBuffer(requestBody), client)
}

// Function to remove a schedule event to the remote scheduler server
func (schedulerClient SchedulerClient) RemoveScheduleEvent(id string) error {
	client := &http.Client{}

	remoteScheduleEventUrl := fmt.Sprintf(UrlPattern, schedulerClient.SchedulerServiceHost, schedulerClient.SchedulerServicePort, ScheduleEventApiPath)
	remoteScheduleEventUrl = remoteScheduleEventUrl + "/" + id

	return doDelete(remoteScheduleEventUrl, client)
}

// Function to do post request
func doPost(url string, binaryReqBody io.Reader, client *http.Client) error {
	req, err := http.NewRequest(HttpPostMethod, url, binaryReqBody)
	req.Header.Add(ContentType, ContentTypeJsonVal)

	if err != nil {
		return err
	}

	return sendRequest(client, req)
}

// Function to do put request
func doPut(url string, binaryReqBody io.Reader, client *http.Client) error {
	req, err := http.NewRequest(HttpPutMethod, url, binaryReqBody)
	req.Header.Add(ContentType, ContentTypeJsonVal)

	if err != nil {
		return err
	}

	return sendRequest(client, req)
}

// Function to do delete request
func doDelete(url string, client *http.Client) error {
	req, err := http.NewRequest(HttpDeleteMethod, url, nil)

	if err != nil {
		return err
	}

	return sendRequest(client, req)
}

// Function to actually make the HTTP request
func sendRequest(client *http.Client, req *http.Request) error {
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		resp.Close = true
		return nil
	} else {
		return err
	}
}

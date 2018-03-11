/*******************************************************************************
 * Copyright 2017 Tencent Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * @microservice: support-scheduler-client-go library
 * @author: Vino Yang, Tencent
 * @version: 0.5.0
 *******************************************************************************/

package scheduler

import (
	"bytes"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/edgexfoundry/edgex-go/support/logging-client"
	"io"
	"net/http"
)

const (
	SchedulerClientName = "support-scheduler-client"
)

// Common http const
const (
	HttpPostMethod     = "POST"
	ContentType        = "Content-Type"
	ContentTypeJsonVal = "application/json"
)

var loggingClient logger.LoggingClient = logger.NewClient(SchedulerClientName, false, "")

// Struct to represent the scheduler client
type SchedulerClient struct {
	RemoteScheduleUrl      string
	RemoteScheduleEventUrl string
	RemoteCallbackAlertUrl string
	OwningService          string
}

// Function to send a schedule to the remote scheduler server
func (schedulerClient SchedulerClient) SendSchedule(schedule models.Schedule) error {
	client := &http.Client{}

	requestBody, err := schedule.MarshalJSON()
	if err != nil {
		loggingClient.Error(err.Error())
		return err
	}

	return doPost(schedulerClient.RemoteScheduleUrl, bytes.NewBuffer(requestBody), client)
}

// Function to send a schedule event to the remote scheduler server
func (schedulerClient SchedulerClient) SendScheduleEvent(scheduleEvent models.ScheduleEvent) error {
	client := &http.Client{}

	requestBody, err := scheduleEvent.MarshalJSON()
	if err != nil {
		loggingClient.Error(err.Error())
		return err
	}

	return doPost(schedulerClient.RemoteScheduleEventUrl, bytes.NewBuffer(requestBody), client)
}

// Function to do post request
func doPost(url string, binaryReqBody io.Reader, client *http.Client) error {
	req, err := http.NewRequest(HttpPostMethod, url, binaryReqBody)
	req.Header.Add(ContentType, ContentTypeJsonVal)

	if err != nil {
		loggingClient.Error(err.Error())
		return err
	}

	// send request call
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
		loggingClient.Error(err.Error())
		return err
	}
}

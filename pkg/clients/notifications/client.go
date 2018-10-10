/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
 *******************************************************************************/
package notifications

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

type CategoryEnum string

const (
	SECURITY  CategoryEnum = "SECURITY"
	HW_HEALTH CategoryEnum = "HW_HEALTH"
	SW_HEALTH CategoryEnum = "SW_HEALTH"
)

type SeverityEnum string

const (
	CRITICAL SeverityEnum = "CRITICAL"
	NORMAL   SeverityEnum = "NORMAL"
)

type StatusEnum string

const (
	NEW       StatusEnum = "NEW"
	PROCESSED StatusEnum = "PROCESSED"
	ESCALATED StatusEnum = "ESCALATED"
)

// Common http const
const (
	ContentType        = "Content-Type"
	ContentTypeJsonVal = "application/json"
)

// Interface defining behavior for the notifications client
type NotificationsClient interface {
	SendNotification(n Notification) error
}

// Type struct for REST-specific implementation of NotificationsClient interface
type notificationsRestClient struct {
	url      string
	endpoint clients.Endpointer
}

// Struct to represent a notification being sent to the notifications service
type Notification struct {
	Id          string       `json:"id,omitempty"` // Generated by the system, users can ignore
	Slug        string       `json:"slug"`         // A meaningful identifier provided by client
	Sender      string       `json:"sender"`
	Category    CategoryEnum `json:"category"`
	Severity    SeverityEnum `json:"severity"`
	Content     string       `json:"content"`
	Description string       `json:"description,omitempty"`
	Status      StatusEnum   `json:"status,omitempty"`
	Labels      []string     `json:"labels,omitempty"`
	Created     int          `json:"created,omitempty"`  // The creation timestamp
	Modified    int          `json:"modified,omitempty"` // The last modification timestamp
}

func NewNotificationsClient(params types.EndpointParams, m clients.Endpointer) NotificationsClient {
	n := notificationsRestClient{endpoint: m}
	n.init(params)
	return &n
}

func (n *notificationsRestClient) init(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go n.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
					n.url = url
				}
			}
		}(ch)
	} else {
		n.url = params.Url
	}
}

// Send a notification to the notifications service
func (nc *notificationsRestClient) SendNotification(n Notification) error {
	client := &http.Client{}

	// Get the JSON request body
	requestBody, err := json.Marshal(n)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, nc.url, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	req.Header.Add(ContentType, ContentTypeJsonVal)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

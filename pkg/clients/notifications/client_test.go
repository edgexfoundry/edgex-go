//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package notifications

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

// Test common const
const (
	TestUnexpectedMsg          = "unexpected result"
	TestUnexpectedMsgFormatStr = "unexpected result, active: '%s' but expected: '%s'"
)

// Test Notification model const fields
const (
	TestNotificationSender      = "Microservice Name"
	TestNotificationCategory    = SW_HEALTH
	TestNotificationSeverity    = NORMAL
	TestNotificationContent     = "This is a notification"
	TestNotificationDescription = "This is a description"
	TestNotificationStatus      = NEW
	TestNotificationLabel1      = "Label One"
	TestNotificationLabel2      = "Label Two"
)

func TestReceiveNotification(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ 'status' : 'OK' }"))
		if r.Method != http.MethodPost {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, http.MethodPost)
		}
		if r.URL.EscapedPath() != clients.ApiNotificationRoute {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), clients.ApiNotificationRoute)
		}

		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()

		var receivedNotification Notification
		json.Unmarshal([]byte(result), &receivedNotification)

		if receivedNotification.Sender != TestNotificationSender {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedNotification.Sender, TestNotificationSender)
		}

		if receivedNotification.Category != TestNotificationCategory {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedNotification.Category, TestNotificationCategory)
		}

		if receivedNotification.Severity != TestNotificationSeverity {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedNotification.Severity, TestNotificationSeverity)
		}

		if receivedNotification.Content != TestNotificationContent {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedNotification.Content, TestNotificationContent)
		}

		if receivedNotification.Description != TestNotificationDescription {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedNotification.Description, TestNotificationDescription)
		}

		if receivedNotification.Status != TestNotificationStatus {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedNotification.Status, TestNotificationStatus)
		}

		if len(receivedNotification.Labels) != 2 {
			t.Error(TestUnexpectedMsg)
		}

		if receivedNotification.Labels[0] != TestNotificationLabel1 {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedNotification.Labels[0], TestNotificationLabel1)
		}

		if receivedNotification.Labels[1] != TestNotificationLabel2 {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedNotification.Labels[1], TestNotificationLabel2)
		}

	}))

	defer ts.Close()

	url := ts.URL + clients.ApiNotificationRoute

	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        clients.ApiNotificationRoute,
		UseRegistry: false,
		Url:         url,
		Interval:    clients.ClientMonitorDefault,
	}

	nc := NewNotificationsClient(params, mockNotificationEndpoint{})

	notification := Notification{
		Sender:      TestNotificationSender,
		Category:    TestNotificationCategory,
		Severity:    TestNotificationSeverity,
		Content:     TestNotificationContent,
		Description: TestNotificationDescription,
		Status:      TestNotificationStatus,
		Labels:      []string{TestNotificationLabel1, TestNotificationLabel2},
	}

	nc.SendNotification(notification)
}

type mockNotificationEndpoint struct {
}

func (e mockNotificationEndpoint) Monitor(params types.EndpointParams, ch chan string) {
	switch params.ServiceKey {
	case internal.CoreMetaDataServiceKey:
		url := params.Url
		ch <- url
		break
	default:
		ch <- ""
	}
}

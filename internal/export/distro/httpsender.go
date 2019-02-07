//
// Copyright (c) 2017
// Cavium
// Mainflux
// IOTech
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

type httpSender struct {
	url    string
	method string
}

const mimeTypeJSON = "application/json"

// newHTTPSender - create http sender
func newHTTPSender(addr contract.Addressable) sender {

	sender := httpSender{
		url:    addr.Protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + addr.Path,
		method: addr.HTTPMethod,
	}
	return sender
}

// Send will send the optionally filtered, compressed, encypted contract.Event via HTTP POST
// The model.Event is provided in order to obtain the necessary correlation-id.
func (sender httpSender) Send(data []byte, event *models.Event) bool {

	switch sender.method {
	case http.MethodPost:
		ctx := context.WithValue(context.Background(), clients.CorrelationHeader, event.CorrelationId)
		req, err := http.NewRequest(http.MethodPost, sender.url, bytes.NewReader(data))
		if err != nil {
			return false
		}
		req.Header.Set("Content-Type", mimeTypeJSON)

		c := clients.NewCorrelatedRequest(req, ctx)
		client := &http.Client{}
		begin := time.Now()
		response, err := client.Do(c.Request)
		if err != nil {
			LoggingClient.Error(err.Error(), clients.CorrelationHeader, event.CorrelationId, internal.LogDurationKey, time.Since(begin).String())
			return false
		}
		defer response.Body.Close()
		LoggingClient.Info(fmt.Sprintf("Response: %s", response.Status), clients.CorrelationHeader, event.CorrelationId, internal.LogDurationKey, time.Since(begin).String())
	default:
		LoggingClient.Info(fmt.Sprintf("Unsupported method: %s", sender.method))
		return false
	}

	LoggingClient.Info(fmt.Sprintf("Sent data: %X", data))
	return true
}

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
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/pkg/models"
	"go.uber.org/zap"
)

type httpSender struct {
	url      string
	method   string
	user     string
	password string
}

const mimeTypeJSON = "application/json"

// NewHTTPSender - create http sender
func NewHTTPSender(addr models.Addressable) Sender {

	sender := httpSender{
		url:      addr.Protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + addr.Path,
		method:   addr.HTTPMethod,
		user:     addr.User,
		password: addr.Password,
	}

	return sender
}

func (sender httpSender) Send(data []byte, event *models.Event) bool {
	switch sender.method {
	case http.MethodPost:
		client := &http.Client{}
		request, _ := http.NewRequest("POST", sender.url, bytes.NewReader(data))
		request.Header.Set("Content-Type", mimeTypeJSON)
		if sender.user != "" {
			request.SetBasicAuth(sender.user, sender.password)
		}
		response, err := client.Do(request)
		if err != nil {
			logger.Error("Error: ", zap.Error(err))
			return false
		}
		defer response.Body.Close()
		logger.Info("Response: ", zap.String("status", response.Status))
	default:
		logger.Info("Unsupported method: ", zap.String("method", sender.method))
		return false
	}

	logger.Info("Sent data: ", zap.ByteString("data", data))
	return true
}

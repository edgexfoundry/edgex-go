//
// Copyright (c) 2017
// Cavium
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"go.uber.org/zap"
)

type httpSender struct {
	url    string
	method string
}

const mimeTypeJSON = "application/json"

// NewHTTPSender - create http sender
func NewHTTPSender(addr models.Addressable) Sender {

	sender := httpSender{
		url:    addr.Protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + addr.Path,
		method: addr.HTTPMethod,
	}
	return sender
}

func (sender httpSender) Send(data []byte) {
	switch sender.method {
	case http.MethodPost:
		response, err := http.Post(sender.url, mimeTypeJSON, bytes.NewReader(data))
		if err != nil {
			logger.Error("Error: ", zap.Error(err))
			return
		}
		defer response.Body.Close()
		logger.Info("Response: ", zap.String("status", response.Status))
	default:
		logger.Info("Unsupported method: ", zap.String("method", sender.method))
		return
	}

	logger.Info("Sent data: ", zap.ByteString("data", data))
}

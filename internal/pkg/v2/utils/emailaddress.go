//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// SendEmailWithAddress sends request with Email address
func SendEmailWithAddress(lc logger.LoggingClient, content string, contentType string, address models.EmailAddress) (res string, err errors.EdgeX) {
	// TODO Send Request With EmailAddress
	lc.Debugf("success to send email with address %v", address.BaseAddress)
	return "", nil
}

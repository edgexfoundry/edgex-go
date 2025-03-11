//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/labstack/echo/v4"
)

// SecretStoreAuthenticationHandlerFunc verifies the JWT with a OpenBao-based JWT authentication check
func SecretStoreAuthenticationHandlerFunc(secretProvider interfaces.SecretProviderExt, lc logger.LoggingClient, token string, c echo.Context) errors.EdgeX {
	r := c.Request()

	validToken, err := secretProvider.IsJWTValid(token)
	if err != nil {
		lc.Errorf("Error checking JWT validity by the secret provider: %v ", err)
		return errors.NewCommonEdgeX(errors.KindServerError, "Error checking JWT validity by the secret provider", err)
	} else if !validToken {
		lc.Warnf("Request to '%s' UNAUTHORIZED", r.URL.Path)
		return errors.NewCommonEdgeX(errors.KindUnauthorized, fmt.Sprintf("Request to '%s' UNAUTHORIZED", r.URL.Path), err)
	}
	lc.Debugf("Request to '%s' authorized", r.URL.Path)
	return nil
}

//
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// WrapHandler wraps `handler func(http.ResponseWriter, *http.Request)` into `echo.HandlerFunc`
func WrapHandler(handler func(http.ResponseWriter, *http.Request)) echo.HandlerFunc {
	return func(c echo.Context) error {
		handler(c.Response(), c.Request())
		return nil
	}
}

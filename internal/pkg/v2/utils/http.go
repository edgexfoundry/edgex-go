//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

func WriteHttpHeader(w http.ResponseWriter, ctx context.Context, statusCode int) {
	w.Header().Set(clients.CorrelationHeader, correlation.FromContext(ctx))
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(statusCode)
}

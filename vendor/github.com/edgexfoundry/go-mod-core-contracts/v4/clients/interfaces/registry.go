//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// RegistryClient defines the interface for interactions with the registry endpoint on the EdgeX core-keeper service.
type RegistryClient interface {
	Register(context.Context, requests.AddRegistrationRequest) errors.EdgeX
	UpdateRegister(context.Context, requests.AddRegistrationRequest) errors.EdgeX
	RegistrationByServiceId(context.Context, string) (responses.RegistrationResponse, errors.EdgeX)
	AllRegistry(context.Context, bool) (responses.MultiRegistrationsResponse, errors.EdgeX)
	Deregister(context.Context, string) errors.EdgeX
}

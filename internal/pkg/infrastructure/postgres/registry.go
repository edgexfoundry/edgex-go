//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

func (c *Client) AddRegistration(r models.Registration) (models.Registration, errors.EdgeX) {
	return models.Registration{}, nil
}

func (c *Client) DeleteRegistrationByServiceId(id string) errors.EdgeX { return nil }

func (c *Client) Registrations() ([]models.Registration, errors.EdgeX) { return nil, nil }

func (c *Client) RegistrationByServiceId(id string) (models.Registration, errors.EdgeX) {
	return models.Registration{}, nil
}

func (c *Client) UpdateRegistration(r models.Registration) errors.EdgeX { return nil }

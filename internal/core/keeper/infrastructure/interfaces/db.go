//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type DBClient interface {
	KeeperKeys(key string, keyOnly bool, isRaw bool) ([]models.KVResponse, errors.EdgeX)
	AddKeeperKeys(kv models.KVS, isFlatten bool) ([]models.KeyOnly, errors.EdgeX)
	DeleteKeeperKeys(key string, isRecurse bool) ([]models.KeyOnly, errors.EdgeX)

	AddRegistration(r models.Registration) (models.Registration, errors.EdgeX)
	DeleteRegistrationByServiceId(id string) errors.EdgeX
	Registrations() ([]models.Registration, errors.EdgeX)
	RegistrationByServiceId(id string) (models.Registration, errors.EdgeX)
	UpdateRegistration(r models.Registration) errors.EdgeX
}

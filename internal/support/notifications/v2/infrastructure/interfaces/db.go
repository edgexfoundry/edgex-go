//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

type DBClient interface {
	CloseSession()

	AddSubscription(e models.Subscription) (models.Subscription, errors.EdgeX)
	AllSubscriptions(offset int, limit int) ([]models.Subscription, errors.EdgeX)
}

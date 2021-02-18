//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

type DBClient interface {
	CloseSession()

	AddInterval(e model.Interval) (model.Interval, errors.EdgeX)
	IntervalByName(name string) (model.Interval, errors.EdgeX)
}

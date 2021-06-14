//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// Metrics defines a metrics gathering abstraction.
type Metrics interface {
	Get(ctx context.Context, services []string) ([]interface{}, errors.EdgeX)
}

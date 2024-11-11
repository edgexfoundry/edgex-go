//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	schedulerInterfaces "github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces"
)

// Check the implementation of Postgres satisfies the DB client
var _ schedulerInterfaces.DBClient = &Client{}

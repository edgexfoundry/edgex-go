//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	cronSchedulerInterfaces "github.com/edgexfoundry/edgex-go/internal/support/cronscheduler/infrastructure/interfaces"
)

// Check the implementation of Postgres satisfies the DB client
var _ cronSchedulerInterfaces.DBClient = &Client{}

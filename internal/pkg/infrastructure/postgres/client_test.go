//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	dataInterfaces "github.com/edgexfoundry/edgex-go/internal/core/data/infrastructure/interfaces"
	metadataInterfaces "github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces"
	notificationsInterfaces "github.com/edgexfoundry/edgex-go/internal/support/notifications/infrastructure/interfaces"
	schedulerInterfaces "github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces"
)

// Check the implementation of Postgres satisfies the DB client
var _ dataInterfaces.DBClient = &Client{}
var _ metadataInterfaces.DBClient = &Client{}
var _ schedulerInterfaces.DBClient = &Client{}
var _ notificationsInterfaces.DBClient = &Client{}

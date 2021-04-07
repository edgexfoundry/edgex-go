//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import "github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/infrastructure/interfaces"

var _ interfaces.SchedulerClient = &Client{}

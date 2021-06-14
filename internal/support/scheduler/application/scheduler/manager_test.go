//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/stretchr/testify/require"
)

var _ interfaces.SchedulerManager = &manager{}

func TestNewManager(t *testing.T) {
	lc := logger.NewMockClient()
	config := &config.ConfigurationStruct{
		Intervals:            nil,
		IntervalActions:      nil,
		ScheduleIntervalTime: 500,
	}
	manager := NewManager(lc, config)
	require.NotNil(t, manager)
}

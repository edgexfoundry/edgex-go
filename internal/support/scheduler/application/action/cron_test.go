//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cronDef(crontab string) models.CronScheduleDef {
	return models.CronScheduleDef{
		BaseScheduleDef: models.BaseScheduleDef{Type: common.DefCron},
		Crontab:         crontab,
	}
}

func TestWindowLocation(t *testing.T) {
	taipei, err := time.LoadLocation("Asia/Taipei")
	require.NoError(t, err)

	tests := []struct {
		name     string
		def      models.ScheduleDef
		expected *time.Location
	}{
		{"CRON_TZ prefix resolves to that zone", cronDef("CRON_TZ=Asia/Taipei 0 0 * * *"), taipei},
		{"TZ prefix resolves to that zone", cronDef("TZ=Asia/Taipei 0 0 * * *"), taipei},
		{"CRON without prefix falls back to Local", cronDef("0 0 * * *"), time.Local},
		{"INTERVAL falls back to Local", models.IntervalScheduleDef{
			BaseScheduleDef: models.BaseScheduleDef{Type: common.DefInterval}, Interval: "10m"}, time.Local},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := WindowLocation(tt.def)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, loc)
		})
	}
}

func TestWindowLocation_InvalidCrontabReturnsError(t *testing.T) {
	_, err := WindowLocation(cronDef("CRON_TZ=Not/A_Zone 0 0 * * *"))
	assert.Error(t, err, "an unloadable timezone must return an error rather than silently defaulting")
}

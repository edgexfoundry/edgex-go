//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// ParseCronExpression parses a crontab into a cron.Schedule. When the expression carries no TZ=/CRON_TZ=
// prefix, it defaults the location to time.Local, so the returned schedule always has a resolved timezone.
func ParseCronExpression(cronExpr string) (cron.Schedule, error) {
	var withLocation string
	if strings.HasPrefix(cronExpr, "TZ=") || strings.HasPrefix(cronExpr, "CRON_TZ=") {
		withLocation = cronExpr
	} else {
		withLocation = fmt.Sprintf("CRON_TZ=%s %s", time.Local.String(), cronExpr)
	}

	// An optional 6th field is used at the beginning since withSeconds is set to true: `* * * * * *`
	p := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := p.Parse(withLocation)
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

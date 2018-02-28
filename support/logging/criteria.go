//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"strings"

	"github.com/edgexfoundry/edgex-go/support/domain"
)

type matchCriteria struct {
	OriginServices []string
	LogLevels      []string
	Labels         []string
	Keywords       []string
	Start          int64
	End            int64
	Limit          int
}

func matchStringInSlice(s string, l []string) bool {
	if len(l) > 0 {
		for _, i := range l {
			if i == s {
				return true
			}
		}
		return false
	}
	return true
}

func (criteria matchCriteria) match(le support_domain.LogEntry) bool {
	if !matchStringInSlice(le.OriginService, criteria.OriginServices) {
		return false
	}
	if !matchStringInSlice(le.Level, criteria.LogLevels) {
		return false
	}

	if criteria.Start > 0 {
		if criteria.Start > le.Created {
			return false
		}
	}
	if criteria.End > 0 {
		if criteria.End < le.Created {
			return false
		}
	}
	if len(criteria.Keywords) > 0 {
		found := false
		for _, keyword := range criteria.Keywords {
			if strings.Contains(le.Message, keyword) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(criteria.Labels) > 0 {
		found := false
		for _, label := range criteria.Labels {
			if matchStringInSlice(label, le.Labels) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

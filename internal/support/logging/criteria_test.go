//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func TestCriteriaMatch(t *testing.T) {
	var services = []string{"service1", "service2"}
	var levels = []string{logger.TraceLog, logger.DebugLog}
	var keywords = []string{"2"}
	var keywordsEmptyString = []string{""}

	var tests = []struct {
		name     string
		log      models.LogEntry
		criteria matchCriteria
		result   bool
	}{
		{"empty/empty", models.LogEntry{}, matchCriteria{}, true},
		// services
		{"emptyService", models.LogEntry{}, matchCriteria{OriginServices: services}, false},
		{"wrongService", models.LogEntry{OriginService: "service11"}, matchCriteria{OriginServices: services}, false},
		{"matchService", models.LogEntry{OriginService: "service1"}, matchCriteria{OriginServices: services}, true},
		// Levels
		{"wrongLevel", models.LogEntry{Level: logger.WarnLog}, matchCriteria{LogLevels: levels}, false},
		{"matchLevel", models.LogEntry{Level: logger.DebugLog}, matchCriteria{LogLevels: levels}, true},
		// Start
		{"0Start", models.LogEntry{Created: 5}, matchCriteria{Start: 0}, true},
		{"wrongStart", models.LogEntry{Created: 5}, matchCriteria{Start: 6}, false},
		{"matchStart1", models.LogEntry{Created: 6}, matchCriteria{Start: 6}, true},
		{"matchStart2", models.LogEntry{Created: 7}, matchCriteria{Start: 6}, true},
		// End
		{"0End", models.LogEntry{Created: 5}, matchCriteria{End: 0}, true},
		{"matchEnd", models.LogEntry{Created: 5}, matchCriteria{End: 6}, true},
		{"matchEnd1", models.LogEntry{Created: 6}, matchCriteria{End: 6}, true},
		{"wrongEnd", models.LogEntry{Created: 7}, matchCriteria{End: 6}, false},
		// keywords
		{"noKeywords", models.LogEntry{Message: "111111"}, matchCriteria{}, true},
		{"wrongKeywords", models.LogEntry{Message: "111111"}, matchCriteria{Keywords: keywords}, false},
		{"matchKeywords", models.LogEntry{Message: "222222"}, matchCriteria{Keywords: keywords}, true},
		{"KeywordsEmptyString", models.LogEntry{Message: "222222"}, matchCriteria{Keywords: keywordsEmptyString}, true},
		{"KeywordsEmptyString2", models.LogEntry{Message: ""}, matchCriteria{Keywords: keywordsEmptyString}, true},
	}
	le := models.LogEntry{}

	criteria := matchCriteria{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result != tt.criteria.match(tt.log) {
				t.Errorf("matching log %v criteria %v should be %v",
					tt.log, tt.criteria, tt.result)
			}
		})
	}

	if !criteria.match(le) {
		t.Errorf("log %v should match criteria %v", le, criteria)
	}
}

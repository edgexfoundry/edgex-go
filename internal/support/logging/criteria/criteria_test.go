//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package criteria

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func TestCriteriaMatch(t *testing.T) {
	var services = []string{"service1", "service2"}
	var levels = []string{models.TraceLog, models.DebugLog}
	var keywords = []string{"2"}
	var keywordsEmptyString = []string{""}

	var tests = []struct {
		name     string
		log      models.LogEntry
		criteria Criteria
		result   bool
	}{
		{"empty/empty", models.LogEntry{}, Criteria{}, true},
		// services
		{"emptyService", models.LogEntry{}, Criteria{OriginServices: services}, false},
		{"wrongService", models.LogEntry{OriginService: "service11"}, Criteria{OriginServices: services}, false},
		{"matchService", models.LogEntry{OriginService: "service1"}, Criteria{OriginServices: services}, true},
		// Levels
		{"wrongLevel", models.LogEntry{Level: models.WarnLog}, Criteria{LogLevels: levels}, false},
		{"matchLevel", models.LogEntry{Level: models.DebugLog}, Criteria{LogLevels: levels}, true},
		// Start
		{"0Start", models.LogEntry{Created: 5}, Criteria{Start: 0}, true},
		{"wrongStart", models.LogEntry{Created: 5}, Criteria{Start: 6}, false},
		{"matchStart1", models.LogEntry{Created: 6}, Criteria{Start: 6}, true},
		{"matchStart2", models.LogEntry{Created: 7}, Criteria{Start: 6}, true},
		// End
		{"0End", models.LogEntry{Created: 5}, Criteria{End: 0}, true},
		{"matchEnd", models.LogEntry{Created: 5}, Criteria{End: 6}, true},
		{"matchEnd1", models.LogEntry{Created: 6}, Criteria{End: 6}, true},
		{"wrongEnd", models.LogEntry{Created: 7}, Criteria{End: 6}, false},
		// keywords
		{"noKeywords", models.LogEntry{Message: "111111"}, Criteria{}, true},
		{"wrongKeywords", models.LogEntry{Message: "111111"}, Criteria{Keywords: keywords}, false},
		{"matchKeywords", models.LogEntry{Message: "222222"}, Criteria{Keywords: keywords}, true},
		{"KeywordsEmptyString", models.LogEntry{Message: "222222"}, Criteria{Keywords: keywordsEmptyString}, true},
		{"KeywordsEmptyString2", models.LogEntry{Message: ""}, Criteria{Keywords: keywordsEmptyString}, true},
	}
	le := models.LogEntry{}

	criteria := Criteria{}

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

//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

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
		criteria MatchCriteria
		result   bool
	}{
		{"empty/empty", models.LogEntry{}, MatchCriteria{}, true},
		// services
		{"emptyService", models.LogEntry{}, MatchCriteria{OriginServices: services}, false},
		{"wrongService", models.LogEntry{OriginService: "service11"}, MatchCriteria{OriginServices: services}, false},
		{"matchService", models.LogEntry{OriginService: "service1"}, MatchCriteria{OriginServices: services}, true},
		// Levels
		{"wrongLevel", models.LogEntry{Level: models.WarnLog}, MatchCriteria{LogLevels: levels}, false},
		{"matchLevel", models.LogEntry{Level: models.DebugLog}, MatchCriteria{LogLevels: levels}, true},
		// Start
		{"0Start", models.LogEntry{Created: 5}, MatchCriteria{Start: 0}, true},
		{"wrongStart", models.LogEntry{Created: 5}, MatchCriteria{Start: 6}, false},
		{"matchStart1", models.LogEntry{Created: 6}, MatchCriteria{Start: 6}, true},
		{"matchStart2", models.LogEntry{Created: 7}, MatchCriteria{Start: 6}, true},
		// End
		{"0End", models.LogEntry{Created: 5}, MatchCriteria{End: 0}, true},
		{"matchEnd", models.LogEntry{Created: 5}, MatchCriteria{End: 6}, true},
		{"matchEnd1", models.LogEntry{Created: 6}, MatchCriteria{End: 6}, true},
		{"wrongEnd", models.LogEntry{Created: 7}, MatchCriteria{End: 6}, false},
		// keywords
		{"noKeywords", models.LogEntry{Message: "111111"}, MatchCriteria{}, true},
		{"wrongKeywords", models.LogEntry{Message: "111111"}, MatchCriteria{Keywords: keywords}, false},
		{"matchKeywords", models.LogEntry{Message: "222222"}, MatchCriteria{Keywords: keywords}, true},
		{"KeywordsEmptyString", models.LogEntry{Message: "222222"}, MatchCriteria{Keywords: keywordsEmptyString}, true},
		{"KeywordsEmptyString2", models.LogEntry{Message: ""}, MatchCriteria{Keywords: keywordsEmptyString}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result != tt.criteria.Match(tt.log) {
				t.Errorf("matching log %v criteria %v should be %v",
					tt.log, tt.criteria, tt.result)
			}
		})
	}
}
